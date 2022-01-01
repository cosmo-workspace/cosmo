import { useSnackbar } from "notistack";
import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { ApiV1alpha1UserAddons, Template, TemplateApiFactory, User, UserApiFactory } from "../../api/dashboard/v1alpha1";
import { ModuleContext } from "../../components/ContextProvider";
import { useProgress } from "../../components/ProgressProvider";

/**
   * error handler
   */
const useHandleError = () => {
  const { enqueueSnackbar } = useSnackbar();
  const navigate = useNavigate();

  const handleError = (error: any) => {
    console.log('handleError', error);
    console.log('handleError', error.response);
    if (error?.response?.status === 401) {
      navigate('/signin');
    }
    const msg = error?.response?.data?.message || error?.message;
    msg && enqueueSnackbar(msg, { variant: 'error' });
    throw error;
  }
  return { handleError }
}

/**
 * hooks
 */
const useUser = () => {
  console.log('useUserModule');

  const { enqueueSnackbar } = useSnackbar();
  const { setMask, releaseMask } = useProgress();
  const { handleError } = useHandleError();
  const [users, setUsers] = useState<User[]>([]);
  const restUser = UserApiFactory(undefined, "");

  /**
   * WorkspaceList: workspace list 
   */
  const getUsers = () => {
    console.log('getUsers');
    return restUser.getUsers()
      .then(result => setUsers(result.data.items?.sort((a, b) => (a.id < b.id) ? -1 : 1) || []))
      .catch(error => { handleError(error); })
  }

  /**
   * CreateDialog: Add user 
   */
  const createUser = async (id: string, displayName: string, role?: string, addons?: ApiV1alpha1UserAddons[]) => {
    console.log('addUser');
    try {
      setMask();
      const result = await restUser.postUser({ id, displayName, role, addons });
      enqueueSnackbar(result.data.message, { variant: 'success' });
      return result.data.user;
    }
    catch (error) { handleError(error); }
    finally { releaseMask(); }
  }

  /**
   * updateNameDialog: Update user name
   */
  const updateName = async (id: string, userName: string) => {
    console.log('updateUserName', id, userName);
    try {
      setMask();
      const result = await restUser.putUserName(id, { displayName: userName });
      const newUser = result.data.user;
      enqueueSnackbar(result.data.message, { variant: 'success' });
      if (users && newUser) {
        setUsers(prev => prev.map(us => us.id === newUser.id ? { ...newUser } : us));
      }
      return newUser;
    }
    catch (error) { handleError(error); }
    finally { releaseMask(); }
  }

  /**
   * updateRoleDialog: Update user 
   */
  const updateRole = async (id: string, role: string) => {
    console.log('updateRole', id, role);
    try {
      setMask();
      const result = await restUser.putUserRole(id, { role });
      const newUser = result.data.user;
      enqueueSnackbar(result.data.message, { variant: 'success' });
      if (users && newUser) {
        setUsers(prev => prev.map(us => us.id === newUser.id ? { ...newUser } : us));
      }
      return newUser;
    }
    catch (error) { handleError(error); }
    finally { releaseMask(); }
  }

  /**
   * DeleteDialog: Delete user 
   */
  const deleteUser = async (userId: string) => {
    console.log('deleteUser');
    try {
      setMask();
      const result = await restUser.deleteUser(userId);
      enqueueSnackbar(result.data.message, { variant: 'success' });
      setUsers(users.filter((u) => u.id !== userId));
      return result;
    }
    catch (error) { handleError(error); }
    finally { releaseMask(); }
  }

  return (
    {
      users,
      getUsers,
      createUser,
      updateName,
      updateRole,
      deleteUser,
    }
  );
}

/**
 * TemplateModule
 */
export const useTemplates = () => {
  console.log('useTemplates');

  const [templates, setTemplates] = useState<Template[]>([]);
  const restTmpl = TemplateApiFactory(undefined, "");
  const { handleError } = useHandleError();

  const getUserAddonTemplates = () => {
    console.log('getUserAddonTemplates');
    return restTmpl.getUserAddonTemplates()
      .then(result => { setTemplates(result.data.items.sort((a, b) => (a.name < b.name) ? -1 : 1)); })
      .catch(error => { handleError(error) });
  }

  return ({
    templates,
    getUserAddonTemplates,
  });
}

/**
 * UserProvider
 */
export const UserContext = ModuleContext(useUser);
export const useUserModule = UserContext.useContext;
