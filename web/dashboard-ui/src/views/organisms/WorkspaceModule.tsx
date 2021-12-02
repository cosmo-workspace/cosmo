import { useSnackbar } from "notistack";
import { useState } from "react";
import { useHistory } from "react-router-dom";
import { NetworkRule, UpsertNetworkRuleRequest, Template, TemplateApiFactory, User, UserApiFactory, Workspace, WorkspaceApiFactory } from "../../api/dashboard/v1alpha1";
import { ModuleContext } from "../../components/ContextProvider";
import { useLogin } from "../../components/LoginProvider";
import { useProgress } from "../../components/ProgressProvider";

export function computeStatus(workspace: Workspace) {
  const status = workspace.status!.phase;
  const replicas = workspace.spec!.replicas;
  if (replicas === 0) {
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
  const restWS = WorkspaceApiFactory(undefined, "");
  /**
   * WorkspaceList: workspace list
   */
  const getWorkspaces = (userId: string) => {
    console.log('getWorkspaces', userId);
    setMask();
    return restWS.getWorkspaces(userId)
      .then(result => {
        const datas = result.data.items.sort((a, b) => (a.name < b.name) ? -1 : 1);
        setWorkspaces(datas);
      })
      .catch(error => { handleError(error) })
      .finally(() => { releaseMask() });
  }

  /**
   * WorkspaceList: refresh
   */
  const refreshWorkspaces = (userId: string) => {
    console.log('refreshWorkspace');
    getWorkspaces(userId);
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
      const result = await restWS.getWorkspace(workspace.ownerID!, workspace.name);
      newWorkspace = result.data.workspace!;
    }
    catch (e) {
      if (computeStatus(workspace) !== 'Creating') {
        console.log('handleError', e);
        return;
      }
    }
    if ((newWorkspace.spec?.additionalNetwork &&
      newWorkspace.spec.additionalNetwork.filter((v) => (!v.url)).length !== 0) ||
      (!['Running', 'Stopped', 'Error', 'CrashLoopBackOff'].includes(computeStatus(newWorkspace)))) {
      setTimeout(() => refreshWorkspace(newWorkspace, tout), 1000);
    }

    const reducer = (wspaces: Workspace[]) => {
      const index = wspaces.findIndex(ws => ws.ownerID === workspace.ownerID && ws.name === workspace.name);
      if (index >= 0 && JSON.stringify(wspaces[index]) !== JSON.stringify(newWorkspace)) {
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
  const createWorkspace = (userId: string, wsName: string, templateName: string, vars: { [key: string]: string }) => {
    console.log('createWorkspace', wsName, templateName, vars);
    setMask();
    return restWS.postWorkspace(userId, { name: wsName, template: templateName, vars: vars })
      .then(result => {
        const newWs: Workspace = { ...result.data.workspace!, status: { ...result.data.workspace!.status, phase: 'Creating' } };
        workspaces.push(newWs);
        setWorkspaces([...workspaces.sort((a, b) => (a.name < b.name) ? -1 : 1)]);
        enqueueSnackbar(result.data.message, { variant: 'success' });
        refreshWorkspace(newWs);
      })
      .catch(error => { handleError(error); })
      .finally(() => { releaseMask() });
  }

  /**
   * Run workspace
   */
  const runWorkspace = (workspace: Workspace) => {
    console.log('runWorkspace', workspace.name);
    setMask();
    return restWS.patchWorkspace(workspace.ownerID!, workspace.name, { replicas: 1 })
      .then(() => {
        enqueueSnackbar('Successfully run workspace', { variant: 'success' });
        refreshWorkspace(workspace);
      })
      .catch(error => { handleError(error); })
      .finally(() => { releaseMask() });
  }

  /**
   * Stop workspace
   */
  const stopWorkspace = (workspace: Workspace) => {
    console.log('stopWorkspace', workspace.name);
    setMask();
    return restWS.patchWorkspace(workspace.ownerID!, workspace.name, { replicas: 0 })
      .then(() => {
        enqueueSnackbar('Successfully stopped workspace', { variant: 'success' });
        refreshWorkspace(workspace);
      })
      .catch(error => { handleError(error); })
      .finally(() => { releaseMask() });
  }

  /**
   * DestroyDialog: Destroy workspace 
   */
  const deleteWorkspace = async (workspace: Workspace) => {
    console.log('deleteWorkspace', workspace);
    try {
      setMask();
      const result = await restWS.deleteWorkspace(workspace.ownerID!, workspace.name);
      enqueueSnackbar(result.data.message, { variant: 'success' });
      refreshWorkspaces(workspace.ownerID!);
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
  const restTmpl = TemplateApiFactory(undefined, "");
  const { handleError } = useHandleError();

  const getTemplates = () => {
    console.log('getTemplates');
    return restTmpl.getWorkspaceTemplates()
      .then(result => { setTemplates(result.data.items.sort((a, b) => (a.name < b.name) ? -1 : 1)); })
      .catch(error => { handleError(error) });
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
  const restNetwork = WorkspaceApiFactory(undefined, "");
  const workspaceModule = useWorkspaceModule();

  const upsertNetwork = (workspace: Workspace, networkRule: NetworkRule) => {
    console.log('upsertNetwork', workspace, networkRule);
    setMask();
    const nwReq: UpsertNetworkRuleRequest = {
      portNumber: networkRule.portNumber,
      group: networkRule.group,
      httpPath: networkRule.httpPath,
    }
    return restNetwork.putNetworkRule(workspace.ownerID!, workspace.name, networkRule.portName, nwReq)
      .then(result => {
        console.log(result)
        enqueueSnackbar(result.data.message, { variant: 'success' });
        workspaceModule.refreshWorkspace(workspace);
      })
      .catch(error => { handleError(error); })
      .finally(() => { releaseMask() });
  }

  const removeNetwork = (workspace: Workspace, netRuleName: string) => {
    console.log('removeNetwork', workspace, netRuleName);
    setMask();
    return restNetwork.deleteNetworkRule(workspace.ownerID!, workspace.name, netRuleName)
      .then(result => {
        console.log(result)
        enqueueSnackbar(result.data.message, { variant: 'success' });
        workspaceModule.refreshWorkspace(workspace);
      })
      .catch(error => { handleError(error); })
      .finally(() => { releaseMask() });
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
  const [user, setUser] = useState<User>(loginUser || { id: '' });
  const [users, setUsers] = useState<User[]>([loginUser || { id: '' }]);
  const { handleError } = useHandleError();
  const restUser = UserApiFactory(undefined, "");

  const getUsers = async () => {
    console.log('getUsers');
    try {
      const result = await restUser.getUsers();
      setUsers(prev => {
        const newUsers = result.data.items.sort((a, b) => (a.id < b.id) ? -1 : 1);
        return JSON.stringify(prev) === JSON.stringify(newUsers) ? prev : newUsers;
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
  const history = useHistory();

  const handleError = (error: any) => {
    console.log('handleError', error);
    console.log('handleError', error.response);
    if (error?.response?.status === 401) {
      history.push('/signin');
    }
    const msg = error?.response?.data?.message || error?.message;
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
