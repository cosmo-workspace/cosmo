import { PartialMessage, protoInt64 } from "@bufbuild/protobuf";
import { useSnackbar } from "notistack";
import { useEffect, useState } from "react";
import { ModuleContext } from "../../components/ContextProvider";
import { useHandleError, useLogin } from "../../components/LoginProvider";
import { useProgress } from "../../components/ProgressProvider";
import { Event } from "../../proto/gen/dashboard/v1alpha1/event_pb";
import { Template } from "../../proto/gen/dashboard/v1alpha1/template_pb";
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
import { setUserStateFuncFilteredByLoginUserRole } from "./UserModule";

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
  hasWarningEvents(clock: Date): boolean {
    const events = this.events;
    if (events === undefined || events.length === 0) {
      return false;
    }
    if (["Stopped", "Stopping"].includes(this.status?.phase!)) {
      return false;
    }
    return events.filter((e) => e.type === "Warning").filter((e) =>
      latestTime(e) - getTime(this.status?.lastStartedAt) >= 0
    ).filter((e) => (clock.getTime() - latestTime(e)) / 1000 / 60 <= 5 // before 5 minutes ago
    ).length > 0;
  }
}

/**
 * useWorkspace
 */
const useWorkspace = () => {
  console.log("useWorkspace");

  const { enqueueSnackbar } = useSnackbar();
  const { setMask, releaseMask } = useProgress();
  const [workspaces, setWorkspaces] = useState<
    { [key: string]: WorkspaceWrapper }
  >({});
  // order
  // const wss = result.items.sort((a, b) => (a.name < b.name) ? -1 : 1);
  const { handleError } = useHandleError();
  const workspaceService = useWorkspaceService();
  const userService = useUserService();

  const { loginUser, myEvents, updateClock } = useLogin();
  const [user, setUser] = useState<User>(loginUser || new User());
  const [users, setUsers] = useState<User[]>([loginUser || new User()]);

  const checkIsPolling = () => {
    return Object.keys(workspaces).filter((name) =>
      workspaces[name].isPolling()
    ).length > 0;
  };

  const stopAllPolling = () => {
    Object.keys(workspaces).forEach((name) => {
      clearTimer(name);
    });
    enqueueSnackbar("Stop all polling", { variant: "info" });
  };

  useEffect(() => {
    getWorkspaces();
  }, [user]);

  const upsertWorkspace = (ws: Workspace, events?: Event[]) => {
    setWorkspaces((prev) => {
      const pws = prev[ws.name] || new WorkspaceWrapper(ws);
      if (prev[ws.name]) pws.update(ws);
      if (events) pws.events = events;
      return { ...prev, [ws.name]: pws };
    });
  };

  /**
   * WorkspaceList: workspace list
   */
  const getWorkspaces = async (userName?: string) => {
    console.log("getWorkspaces", userName);
    setMask();
    try {
      const getWorkspacesResult = await workspaceService.getWorkspaces({
        userName: userName || user.name,
      });
      const getEventsResult = await userService.getEvents({
        userName: userName || user.name,
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
          const pws = prev[ws.name] || new WorkspaceWrapper(ws);
          if (prev[ws.name]) pws.update(ws);
          pws.events = wsEventMap[ws.name] || [];
          return pws;
        });
        const wsMap = pwsArr.reduce(
          (map: { [key: string]: WorkspaceWrapper }, pws) => {
            map[pws.name] = pws;
            return map;
          },
          {},
        );
        console.log(wsMap);
        return wsMap;
      });
    } catch (error) {
      handleError(error);
    } finally {
      updateClock();
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
      userName: ws.ownerName,
    });
    const events = getEventsResult.items.filter((e) =>
      e.regardingWorkspace === ws.name
    );

    upsertWorkspace(ws, events);
    return ws;
  };

  const setProgress = (wsName: string, progress: number) => {
    setWorkspaces((prev) => {
      if (prev[wsName]) {
        console.log(
          "### update progress",
          `${prev[wsName].progress} -> ${progress}`,
        );
        const pws = prev[wsName];
        pws.progress = progress;
        return { ...prev, [wsName]: pws };
      }
      return prev;
    });
  };

  const setTimer = (wsName: string, timer: NodeJS.Timeout) => {
    setWorkspaces((prev) => {
      if (prev[wsName]) {
        const pws = prev[wsName];
        pws.timer = timer;
        return { ...prev, [wsName]: pws };
      }
      return prev;
    });
  };

  const clearTimer = (wsName: string) => {
    setWorkspaces((prev) => {
      if (prev[wsName]) {
        const pws = prev[wsName];
        clearInterval(pws.timer);
        pws.timer = undefined;
        pws.progress = 120;
        return { ...prev, [wsName]: pws };
      }
      return prev;
    });
  };

  /**
   * WorkspaceList: pollingWorkspace
   */
  const pollingWorkspace = async (
    ws: Workspace,
  ) => {
    ws = await refreshWorkspace(ws);

    let limit = 120;
    let progress = 0;
    setProgress(ws.name, progress);
    setTimer(
      ws.name,
      setInterval(async () => {
        try {
          progress = progress >= 100 ? 0 : progress + 20;
          console.log("polling", "progress", progress);
          setProgress(ws.name, progress);
          if (progress === 20) {
            console.log("polling", "do request");
            const newWs = await refreshWorkspace(ws);

            const undefinedURLs =
              (newWs.spec?.network || []).filter((v) => (!v.url)).length;
            const status = computeStatus(newWs);
            if (
              undefinedURLs === 0 &&
              ["Running", "Stopped", "Error", "CrashLoopBackOff"].includes(
                status,
              )
            ) {
              console.log("polling", "timer stopped");
              clearTimer(ws.name);
            }
          }
        } catch (e) {
          if (computeStatus(ws) !== "Creating") {
            console.log("polling", "error", e);
            clearTimer(ws.name);
          }
        } finally {
          limit--;
          console.log("polling", "limit", limit);
          if (limit < 0) {
            console.log("polling", "reached limit");
            clearTimer(ws.name);
          }
        }
      }, 1000),
    );
  };

  /**
   * CreateDialog: Create workspace
   */
  const createWorkspace = async (
    userName: string,
    wsName: string,
    templateName: string,
    vars: { [key: string]: string },
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
  const getUsers = async () => {
    console.log("useWorkspaceUsers:getUsers");
    try {
      const result = await userService.getUsers({});
      setUsers(
        setUserStateFuncFilteredByLoginUserRole(result.items, loginUser),
      );
    } catch (error) {
      handleError(error);
    }
  };

  return ({
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
  });
};

/**
 * TemplateModule
 */
export const useTemplates = () => {
  console.log("useTemplates");

  const [templates, setTemplates] = useState<Template[]>([]);
  const templateService = useTemplateService();
  const { handleError } = useHandleError();

  const getTemplates = async () => {
    console.log("getTemplates");
    try {
      const result = await templateService.getWorkspaceTemplates({
        useRoleFilter: true,
      });
      setTemplates(result.items.sort((a, b) => (a.name < b.name) ? -1 : 1));
    } catch (error) {
      handleError(error);
    }
  };

  return ({
    templates,
    getTemplates,
  });
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
    index: number,
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

  return ({
    upsertNetwork,
    removeNetwork,
  });
};

/**
 * WorkspaceProvider
 */
export const WorkspaceContext = ModuleContext(useWorkspace);
export const useWorkspaceModule = WorkspaceContext.useContext;
