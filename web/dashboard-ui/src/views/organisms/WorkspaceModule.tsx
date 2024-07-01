import useUrlState from "@ahooksjs/use-url-state";
import { PartialMessage, protoInt64 } from "@bufbuild/protobuf";
import { useSnackbar } from "notistack";
import { useEffect, useState } from "react";
import { ModuleContext } from "../../components/ContextProvider";
import { useHandleError, useLogin } from "../../components/LoginProvider";
import { useProgress } from "../../components/ProgressProvider";
import { Event } from "../../proto/gen/dashboard/v1alpha1/event_pb";
import { Template } from "../../proto/gen/dashboard/v1alpha1/template_pb";
import { GetWorkspaceTemplatesRequest } from "../../proto/gen/dashboard/v1alpha1/template_service_pb";
import { User } from "../../proto/gen/dashboard/v1alpha1/user_pb";
import {
  NetworkRule,
  Workspace,
  WorkspaceStatus,
} from "../../proto/gen/dashboard/v1alpha1/workspace_pb";
import {
  useTemplateService,
  useUserService,
  useWorkspaceService,
} from "../../services/DashboardServices";
import { getTime, latestTime } from "./EventModule";
import {
  isAdminUser,
  setUsersFuncFilteredByAccesibleRoles,
  usersFilteredByAccesibleRoles,
} from "./UserModule";

export function computeStatus(workspace: Workspace) {
  const status = workspace.status!.phase;
  const replicas = workspace.spec!.replicas;
  if (replicas === protoInt64.zero) {
    return status === "Running" ? "Stopping" : status;
  } else if (replicas > 0) {
    return status === "Stopped" ? "Starting" : status;
  }
  return status;
}

export function wskey(ws: Workspace) {
  return `${ws.name}-${ws.ownerName}`;
}

export class WorkspaceWrapper extends Workspace {
  constructor(data: PartialMessage<Workspace>) {
    super({ ...data });
  }
  update(data: Workspace) {
    Object.assign(this, data);
  }
  public timer: NodeJS.Timeout | undefined;
  public progress = 0;
  public events: Event[] = [];
  isPolling(): boolean {
    return this.timer !== undefined;
  }
  warningEventsCount(clock: Date): number {
    const events = this.events;
    if (events === undefined || events.length === 0) {
      return 0;
    }
    if (["Stopped", "Stopping"].includes(this.status?.phase!)) {
      return 0;
    }
    return events
      .filter((e) => e.type === "Warning")
      .filter((e) => latestTime(e) - getTime(this.status?.lastStartedAt) >= 0)
      .filter(
        (e) => (clock.getTime() - latestTime(e)) / 1000 / 60 <= 5 // before 5 minutes ago
      ).length;
  }
  isSharedFor(user: User): boolean {
    if (this.ownerName == user.name) return false;
    const allowed =
      this?.spec?.network?.filter((r) => r.allowedUsers.length > 0) || [];
    return allowed.length > 0;
  }
  readonlyFor(user: User): boolean {
    if (!this.isSharedFor(user)) return false;

    const main = this.spec?.network?.find((r) => r.url == this.status?.mainUrl);
    const canUpdate = main?.allowedUsers.includes(user.name);
    return !canUpdate;
  }
  networkRules(user: User): NetworkRule[] {
    const rules = this.spec?.network || [];
    if (this.isSharedFor(user)) {
      rules.filter((r) => r.allowedUsers.includes(user.name));
    }
    return rules;
  }
  key(): string {
    return wskey(this);
  }
}

/**
 * useWorkspace
 */
const useWorkspace = () => {
  console.log("useWorkspace");

  const { enqueueSnackbar } = useSnackbar();
  const { setMask, releaseMask } = useProgress();
  const [workspaces, setWorkspaces] = useState<{
    [key: string]: WorkspaceWrapper;
  }>({});
  const { handleError } = useHandleError();
  const workspaceService = useWorkspaceService();
  const userService = useUserService();

  const { loginUser, updateClock } = useLogin();
  const isAdmin = isAdminUser(loginUser);
  const [users, setUsers] = useState<User[]>([loginUser || new User()]);

  const [urlParam, setUrlParam] = useUrlState(
    { search: "", user: loginUser?.name || "" },
    {
      stringifyOptions: { skipEmptyString: true },
    }
  );
  const search: string = urlParam.search || "";
  const setSearch = (word: string) => setUrlParam({ search: word });

  const userName: string = urlParam.user || "";
  const user = users.find((u) => u.name === userName) || new User();
  const setUser = (name: string) => setUrlParam({ user: name });

  const checkIsPolling = () => {
    return (
      Object.keys(workspaces).filter((key) => workspaces[key].isPolling())
        .length > 0
    );
  };

  const stopAllPolling = () => {
    Object.keys(workspaces).forEach((key) => {
      clearTimer(key);
    });
  };

  useEffect(() => {
    stopAllPolling();
    getWorkspaces(userName);
    isAdmin &&
      getUsers().then((users) => {
        usersFilteredByAccesibleRoles(users || [], loginUser).find(
          (u) => u.name === userName
        ) ||
          enqueueSnackbar(`User ${userName} is not found`, {
            variant: "error",
          });
      });
  }, [userName]);

  const upsertWorkspace = (ws: Workspace, events?: Event[]) => {
    setWorkspaces((prev) => {
      const pws = prev[wskey(ws)] || new WorkspaceWrapper(ws);
      if (prev[wskey(ws)]) pws.update(ws);
      if (events) pws.events = events;
      return { ...prev, [wskey(ws)]: pws };
    });
    updateClock();
  };

  /**
   * WorkspaceList: workspace list
   */
  const getWorkspaces = async (userName: string) => {
    console.log("getWorkspaces", userName);
    setMask();
    try {
      const getWorkspacesResult = await workspaceService.getWorkspaces({
        userName: userName,
        includeShared: true,
      });
      const getEventsResult = await userService.getEvents({
        userName: userName,
      });

      const workspaces = getWorkspacesResult.items;
      const events = getEventsResult.items;

      const wsEventMap: { [key: string]: Event[] } = {};
      for (const event of events) {
        if (event.regardingWorkspace) {
          wsEventMap[event.regardingWorkspace] = [
            ...(wsEventMap[event.regardingWorkspace] || []),
            event,
          ].sort((a, b) => latestTime(a) - latestTime(b));
        }
      }

      setWorkspaces((prev) => {
        const pwsArr = workspaces.map((ws) => {
          const pws = prev[wskey(ws)] || new WorkspaceWrapper(ws);
          if (prev[wskey(ws)]) pws.update(ws);
          pws.events = wsEventMap[wskey(ws)] || [];
          return pws;
        });
        const wsMap = pwsArr.reduce(
          (map: { [key: string]: WorkspaceWrapper }, pws) => {
            map[wskey(pws)] = pws;
            return map;
          },
          {}
        );
        console.log(wsMap);
        return wsMap;
      });
    } catch (error) {
      handleError(error);
    } finally {
      releaseMask();
    }
  };

  const refreshWorkspace = async (workspace: Workspace): Promise<Workspace> => {
    const getWorkspaceResult = await workspaceService.getWorkspace({
      userName: workspace.ownerName,
      wsName: workspace.name,
    });
    const ws = getWorkspaceResult.workspace!;

    const getEventsResult = await userService.getEvents({
      userName: user.name,
    });
    const events = getEventsResult.items.filter(
      (e) => e.regardingWorkspace === ws.name
    );

    upsertWorkspace(ws, events);
    return ws;
  };

  const setProgress = (key: string, progress: number) => {
    setWorkspaces((prev) => {
      if (prev[key]) {
        console.log(
          "### update progress",
          `${prev[key].progress} -> ${progress}`
        );
        const pws = prev[key];
        pws.progress = progress;
        return { ...prev, [key]: pws };
      }
      return prev;
    });
  };

  const setTimer = (key: string, timer: NodeJS.Timeout) => {
    setWorkspaces((prev) => {
      if (prev[key]) {
        const pws = prev[key];
        clearInterval(pws.timer);
        pws.timer = timer;
        return { ...prev, [key]: pws };
      }
      return prev;
    });
  };

  const clearTimer = (key: string) => {
    setWorkspaces((prev) => {
      if (prev[key]) {
        const pws = prev[key];
        clearInterval(pws.timer);
        pws.timer = undefined;
        pws.progress = 120;
        return { ...prev, [key]: pws };
      }
      return prev;
    });
  };

  /**
   * WorkspaceList: pollingWorkspace
   */
  const pollingWorkspace = async (ws: Workspace) => {
    ws = await refreshWorkspace(ws);

    let limit = 120;
    let progress = 0;
    setProgress(wskey(ws), progress);
    setTimer(
      wskey(ws),
      setInterval(async () => {
        try {
          progress = progress >= 100 ? 0 : progress + 20;
          console.log("polling", "progress", progress);
          setProgress(wskey(ws), progress);
          if (progress === 20) {
            console.log("polling", "do request");
            const newWs = await refreshWorkspace(ws);

            const undefinedURLs = (newWs.spec?.network || []).filter(
              (v) => !v.url
            ).length;
            const status = computeStatus(newWs);
            if (
              undefinedURLs === 0 &&
              ["Running", "Stopped", "Error", "CrashLoopBackOff"].includes(
                status
              )
            ) {
              console.log("polling", "timer stopped");
              clearTimer(wskey(ws));
            }
          }
        } catch (e) {
          if (computeStatus(ws) !== "Creating") {
            console.log("polling", "error", e);
            clearTimer(wskey(ws));
          }
        } finally {
          limit--;
          console.log("polling", "limit", limit);
          if (limit < 0) {
            console.log("polling", "reached limit");
            clearTimer(wskey(ws));
          }
        }
      }, 1000)
    );
  };

  /**
   * CreateDialog: Create workspace
   */
  const createWorkspace = async (
    userName: string,
    wsName: string,
    templateName: string,
    vars: { [key: string]: string }
  ) => {
    console.log("createWorkspace", wsName, templateName, vars);
    setMask();
    try {
      const result_1 = await workspaceService.createWorkspace({
        userName,
        wsName: wsName,
        template: templateName,
        vars: vars,
      });
      const stat = new WorkspaceStatus({
        ...result_1.workspace!.status,
        phase: "Creating",
      });
      const ws = new Workspace({ ...result_1.workspace!, status: stat });
      upsertWorkspace(ws);
      enqueueSnackbar(result_1.message, { variant: "success" });
      pollingWorkspace(ws);
    } catch (error) {
      handleError(error);
    } finally {
      releaseMask();
    }
  };

  /**
   * Run workspace
   */
  const runWorkspace = async (workspace: Workspace) => {
    console.log("runWorkspace", workspace.name);
    setMask();
    try {
      const res = await workspaceService.updateWorkspace({
        userName: workspace.ownerName!,
        wsName: workspace.name,
        replicas: protoInt64.parse(1),
      });
      enqueueSnackbar("Successfully run workspace", { variant: "success" });
      pollingWorkspace(res.workspace!);
    } catch (error) {
      handleError(error);
    } finally {
      releaseMask();
    }
  };

  /**
   * Stop workspace
   */
  const stopWorkspace = async (workspace: Workspace) => {
    console.log("stopWorkspace", workspace.name);
    setMask();
    try {
      const res = await workspaceService.updateWorkspace({
        userName: workspace.ownerName!,
        wsName: workspace.name,
        replicas: protoInt64.zero,
      });
      enqueueSnackbar("Successfully stopped workspace", { variant: "success" });
      pollingWorkspace(res.workspace!);
    } catch (error) {
      handleError(error);
    } finally {
      releaseMask();
    }
  };

  /**
   * DestroyDialog: Destroy workspace
   */
  const deleteWorkspace = async (workspace: Workspace) => {
    console.log("deleteWorkspace", workspace);
    try {
      setMask();
      const result = await workspaceService.deleteWorkspace({
        userName: workspace.ownerName!,
        wsName: workspace.name,
      });
      enqueueSnackbar(result.message, { variant: "success" });
      getWorkspaces(workspace.ownerName!);
    } catch (error) {
      handleError(error);
    } finally {
      releaseMask();
    }
  };

  /**
   * UserModule
   */
  const getUsers = async (): Promise<User[] | undefined> => {
    console.log("useWorkspaceUsers:getUsers");
    try {
      const result = await userService.getUsers({});
      setUsers(setUsersFuncFilteredByAccesibleRoles(result.items, loginUser));
      return result.items;
    } catch (error) {
      handleError(error);
    }
  };

  return {
    workspaces,
    getWorkspaces,
    createWorkspace,
    deleteWorkspace,
    runWorkspace,
    stopWorkspace,
    pollingWorkspace,
    checkIsPolling,
    stopAllPolling,
    user,
    setUser,
    users,
    getUsers,
    search,
    setSearch,
  };
};

/**
 * TemplateModule
 */
export const useTemplates = () => {
  console.log("useTemplates");

  const [templates, setTemplates] = useState<Template[]>([]);
  const templateService = useTemplateService();
  const { handleError } = useHandleError();

  const getTemplates = async (
    option?: PartialMessage<GetWorkspaceTemplatesRequest>
  ) => {
    console.log("getTemplates");
    try {
      const result = await templateService.getWorkspaceTemplates({
        ...option,
      });
      setTemplates(result.items.sort((a, b) => (a.name < b.name ? -1 : 1)));
    } catch (error) {
      handleError(error);
    }
  };

  return {
    templates,
    getTemplates,
  };
};

/**
 * useNetworkRule
 */
export const useNetworkRule = () => {
  console.log("useNetworkRule");

  const { enqueueSnackbar } = useSnackbar();
  const { setMask, releaseMask } = useProgress();
  const { handleError } = useHandleError();
  const workspaceService = useWorkspaceService();
  const workspaceModule = useWorkspaceModule();

  const upsertNetwork = async (
    workspace: Workspace,
    networkRule: NetworkRule,
    index: number
  ) => {
    console.log("upsertNetwork", workspace, networkRule);
    setMask();
    try {
      const result = await workspaceService.upsertNetworkRule({
        userName: workspace.ownerName!,
        wsName: workspace.name,
        networkRule: networkRule,
        index: index,
      });
      console.log(result);
      enqueueSnackbar(result.message, { variant: "success" });
      workspaceModule.pollingWorkspace(new WorkspaceWrapper(workspace));
    } catch (error) {
      handleError(error);
    } finally {
      releaseMask();
    }
  };

  const removeNetwork = async (workspace: Workspace, index: number) => {
    console.log("removeNetwork", workspace, index);
    setMask();
    try {
      const result_1 = await workspaceService.deleteNetworkRule({
        userName: workspace.ownerName!,
        wsName: workspace.name,
        index: index,
      });
      console.log(result_1);
      enqueueSnackbar(result_1.message, { variant: "success" });
      workspaceModule.pollingWorkspace(new WorkspaceWrapper(workspace));
    } catch (error) {
      handleError(error);
    } finally {
      releaseMask();
    }
  };

  return {
    upsertNetwork,
    removeNetwork,
  };
};

/**
 * WorkspaceProvider
 */
export const WorkspaceContext = ModuleContext(useWorkspace);
export const useWorkspaceModule = WorkspaceContext.useContext;
