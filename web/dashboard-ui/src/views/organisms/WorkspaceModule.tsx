import { Code, ConnectError } from "@bufbuild/connect";
import { protoInt64 } from "@bufbuild/protobuf";
import { useSnackbar } from "notistack";
import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { ModuleContext } from "../../components/ContextProvider";
import { useLogin } from "../../components/LoginProvider";
import { useProgress } from "../../components/ProgressProvider";
import { Template } from "../../proto/gen/dashboard/v1alpha1/template_pb";
import { User } from "../../proto/gen/dashboard/v1alpha1/user_pb";
import { NetworkRule, Workspace, WorkspaceStatus } from "../../proto/gen/dashboard/v1alpha1/workspace_pb";
import { useTemplateService, useUserService, useWorkspaceService } from "../../services/DashboardServices";
import { hasAdminForRole, hasPrivilegedRole } from "./UserModule";

export function computeStatus(workspace: Workspace) {
  const status = workspace.status!.phase;
  const replicas = workspace.spec!.replicas;
  if (replicas === protoInt64.zero) {
    return status === 'Running' ? 'Stopping' : status;
  } else if (replicas > 0) {
    return status === 'Stopped' ? 'Starting' : status;
  }
  return status;
}

/**
 * useWorkspace
 */
const useWorkspace = () => {
  console.log('useWorkspace');

  const { enqueueSnackbar } = useSnackbar();
  const { setMask, releaseMask } = useProgress();
  const [workspaces, setWorkspaces] = useState<Workspace[]>([]);
  const { handleError } = useHandleError();
  const workspaceService = useWorkspaceService();
  /**
   * WorkspaceList: workspace list
   */
  const getWorkspaces = async (userName: string) => {
    console.log('getWorkspaces', userName);
    setMask();
    try {
      try {
        const result_1 = await workspaceService.getWorkspaces({ userName });
        const datas = result_1.items.sort((a, b) => (a.name < b.name) ? -1 : 1);
        setWorkspaces(datas);
      } catch (error) {
        handleError(error);
      }
    } finally {
      releaseMask();
    }
  }

  /**
   * WorkspaceList: refresh
   */
  const refreshWorkspaces = (username: string) => {
    console.log('refreshWorkspace');
    getWorkspaces(username);
  }

  /**
   * WorkspaceList: refreshWorkspace
   */
  const refreshWorkspace = async (workspace: Workspace, timeout?: number) => {
    console.log('refreshWorkspace', computeStatus(workspace), timeout);

    if (timeout === undefined) {
      setTimeout(() => refreshWorkspace(workspace, 120000), 500);
      return;
    }
    const tout = timeout! - 1000;   // 1 x 120 seconds
    if (tout < 0) return;

    let newWorkspace = workspace;
    try {
      const result = await workspaceService.getWorkspace({ userName: workspace.ownerName, wsName: workspace.name });
      newWorkspace = result.workspace!;
    }
    catch (e) {
      if (computeStatus(workspace) !== 'Creating') {
        console.log('handleError', e);
        return;
      }
    }
    if ((newWorkspace.spec?.network &&
      newWorkspace.spec.network.filter((v) => (!v.url)).length !== 0) ||
      (!['Running', 'Stopped', 'Error', 'CrashLoopBackOff'].includes(computeStatus(newWorkspace)))) {
      setTimeout(() => refreshWorkspace(newWorkspace, tout), 1000);
    }

    const reducer = (wspaces: Workspace[]) => {
      const index = wspaces.findIndex(ws => ws.ownerName === workspace.ownerName && ws.name === workspace.name);
      if (index >= 0 && !wspaces[index].equals(newWorkspace)) {
        wspaces[index] = newWorkspace;
        return [...wspaces];
      }
      return wspaces;
    }
    setWorkspaces(reducer);
  }


  /**
   * CreateDialog: Create workspace 
   */
  const createWorkspace = async (userName: string, wsName: string, templateName: string, vars: { [key: string]: string }) => {
    console.log('createWorkspace', wsName, templateName, vars);
    setMask();
    try {
      const result_1 = await workspaceService.createWorkspace({ userName, wsName: wsName, template: templateName, vars: vars });
      const stat = new WorkspaceStatus({ ...result_1.workspace!.status, phase: 'Creating' })
      const newWs = new Workspace({ ...result_1.workspace!, status: stat });
      workspaces.push(newWs);
      setWorkspaces([...workspaces.sort((a, b) => (a.name < b.name) ? -1 : 1)]);
      enqueueSnackbar(result_1.message, { variant: 'success' });
      refreshWorkspace(newWs);
    } catch (error) {
      handleError(error);
    } finally {
      releaseMask();
    }
  }

  /**
   * Run workspace
   */
  const runWorkspace = async (workspace: Workspace) => {
    console.log('runWorkspace', workspace.name);
    setMask();
    try {
      await workspaceService.updateWorkspace({ userName: workspace.ownerName!, wsName: workspace.name, replicas: protoInt64.parse(1) });
      enqueueSnackbar('Successfully run workspace', { variant: 'success' });
      refreshWorkspace(workspace);
    } catch (error) {
      handleError(error);
    } finally {
      releaseMask();
    }
  }

  /**
   * Stop workspace
   */
  const stopWorkspace = async (workspace: Workspace) => {
    console.log('stopWorkspace', workspace.name);
    setMask();
    try {
      await workspaceService.updateWorkspace({ userName: workspace.ownerName!, wsName: workspace.name, replicas: protoInt64.zero });
      enqueueSnackbar('Successfully stopped workspace', { variant: 'success' });
      refreshWorkspace(workspace);
    } catch (error) {
      handleError(error);
    } finally {
      releaseMask();
    }
  }

  /**
   * DestroyDialog: Destroy workspace 
   */
  const deleteWorkspace = async (workspace: Workspace) => {
    console.log('deleteWorkspace', workspace);
    try {
      setMask();
      const result = await workspaceService.deleteWorkspace({ userName: workspace.ownerName!, wsName: workspace.name });
      enqueueSnackbar(result.message, { variant: 'success' });
      refreshWorkspaces(workspace.ownerName!);
    }
    catch (error) { handleError(error); }
    finally { releaseMask(); }
  }

  return ({
    workspaces,
    getWorkspaces,
    createWorkspace,
    deleteWorkspace,
    runWorkspace,
    stopWorkspace,
    refreshWorkspace,
    refreshWorkspaces,
  });
}

/**
 * TemplateModule
 */
export const useTemplates = () => {
  console.log('useTemplates');

  const [templates, setTemplates] = useState<Template[]>([]);
  const templateService = useTemplateService();
  const { handleError } = useHandleError();

  const getTemplates = async () => {
    console.log('getTemplates');
    try {
      const result = await templateService.getWorkspaceTemplates({});
      setTemplates(result.items.sort((a, b) => (a.name < b.name) ? -1 : 1));
    } catch (error) {
      handleError(error);
    }
  }

  return ({
    templates,
    getTemplates,
  });
}


/**
 * useNetworkRule
 */
export const useNetworkRule = () => {
  console.log('useNetworkRule');

  const { enqueueSnackbar } = useSnackbar();
  const { setMask, releaseMask } = useProgress();
  const { handleError } = useHandleError();
  const workspaceService = useWorkspaceService();
  const workspaceModule = useWorkspaceModule();

  const upsertNetwork = async (workspace: Workspace, networkRule: NetworkRule, index: number) => {
    console.log('upsertNetwork', workspace, networkRule);
    setMask();
    try {
      const result = await workspaceService.upsertNetworkRule({ userName: workspace.ownerName!, wsName: workspace.name, networkRule: networkRule, index: index });
      console.log(result);
      enqueueSnackbar(result.message, { variant: 'success' });
      workspaceModule.refreshWorkspace(workspace);
    } catch (error) {
      handleError(error);
    } finally {
      releaseMask();
    }
  }

  const removeNetwork = async (workspace: Workspace, index: number) => {
    console.log('removeNetwork', workspace, index);
    setMask();
    try {
      const result_1 = await workspaceService.deleteNetworkRule({ userName: workspace.ownerName!, wsName: workspace.name, index: index });
      console.log(result_1);
      enqueueSnackbar(result_1.message, { variant: 'success' });
      workspaceModule.refreshWorkspace(workspace);
    } catch (error) {
      handleError(error);
    } finally {
      releaseMask();
    }
  }

  return ({
    upsertNetwork,
    removeNetwork,
  });
}


/**
 * useWorkspaceUser
 */
const useWorkspaceUsers = () => {
  console.log('useWorkspaceUser');

  const { loginUser } = useLogin();
  const [user, setUser] = useState<User>(loginUser || new User());
  const [users, setUsers] = useState<User[]>([loginUser || new User()]);
  const { handleError } = useHandleError();
  const userService = useUserService();

  const filterUsersByRoles = (users: User[], myRoles: string[]) => {
    return hasPrivilegedRole(myRoles) ? users : users.filter((u) => {
      for (const userRole of u.roles) {
        if (hasAdminForRole(myRoles, userRole)) {
          return true
        }
      }
      return false
    })
  }

  const getUsers = async () => {
    console.log('useWorkspaceUsers:getUsers');
    try {
      const result = await userService.getUsers({});
      setUsers(prev => {
        const newUsers = result.items.sort((a, b) => (a.name < b.name) ? -1 : 1);
        const roleFilteredNewUsers = filterUsersByRoles(newUsers, (loginUser?.roles || []))
        return JSON.stringify(prev) === JSON.stringify(roleFilteredNewUsers) ? prev : roleFilteredNewUsers;
      });
    }
    catch (error) { handleError(error) };
  }

  return ({
    user, setUser,
    users,
    getUsers,
  });
}


/**
* error handler
*/
const useHandleError = () => {
  const { enqueueSnackbar } = useSnackbar();
  const navigate = useNavigate();
  const { clearLoginUser } = useLogin();

  const handleError = (error: any) => {
    console.log('handleError', error);

    if (error instanceof ConnectError &&
      error.code === Code.Unauthenticated) {
      clearLoginUser();
      navigate('/signin');
    }
    const msg = error?.message;
    msg && enqueueSnackbar(msg, { variant: 'error' });
    throw error;
  }
  return { handleError }
}


/**
 * WorkspaceProvider
 */
export const WorkspaceContext = ModuleContext(useWorkspace);
export const useWorkspaceModule = WorkspaceContext.useContext;

export const WorkspaceUsersContext = ModuleContext(useWorkspaceUsers);
export const useWorkspaceUsersModule = WorkspaceUsersContext.useContext;
