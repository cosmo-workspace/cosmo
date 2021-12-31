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
    catch (error) {
      handleError(error);
    }
    finally {
      releaseMask();
    }
  }

  /**
   * updateRoleDialog: Update user 
   */
  const updateRole = (id: string, role: string) => {
    console.log('updateRole', id, role);
    setMask();
    return restUser.putUserRole(id, { role })
      .then(result => {
        enqueueSnackbar(result.data.message, { variant: 'success' });
        getUsers();
      })
      .catch(error => { handleError(error); })
      .finally(() => { releaseMask() });
  }

  /**
   * DeleteDialog: Delete user 
   */
  const deleteUser = (userId: string) => {
    console.log('deleteUser');
    setMask();
    return restUser.deleteUser(userId)
      .then(result => {
        enqueueSnackbar(result.data.message, { variant: 'success' });
        setUsers(users.filter((u) => u.id !== userId));
      })
      .catch(error => { handleError(error); })
      .finally(() => { releaseMask(); });
  }

  return (
    {
      users,
      getUsers,
      createUser,
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
