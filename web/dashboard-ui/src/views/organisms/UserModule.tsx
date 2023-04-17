import { Code, ConnectError } from "@bufbuild/connect-web";
import { useSnackbar } from "notistack";
import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { ModuleContext } from "../../components/ContextProvider";
import { useProgress } from "../../components/ProgressProvider";
import { Template } from "../../proto/gen/dashboard/v1alpha1/template_pb";
import { User, UserAddons } from "../../proto/gen/dashboard/v1alpha1/user_pb";
import { useTemplateService, useUserService } from "../../services/DashboardServices";

/**
 * hooks
 */
const useUser = () => {
  console.log('useUserModule');

  const { enqueueSnackbar } = useSnackbar();
  const { setMask, releaseMask } = useProgress();
  const { handleError } = useHandleError();
  const [users, setUsers] = useState<User[]>([]);
  const userService = useUserService();

  /**
   * WorkspaceList: workspace list 
   */
  const getUsers = async () => {
    console.log('getUsers');
    try {
      const result = await userService.getUsers({});
      setUsers(result.items?.sort((a, b) => (a.name < b.name) ? -1 : 1));
    } catch (error) {
      handleError(error);
    }
  }

  /**
   * CreateDialog: Add user 
   */
  const createUser = async (userName: string, displayName: string, roles?: string[], addons?: UserAddons[]) => {
    console.log('addUser');
    setMask();
    try {
      try {
        const result = await userService.createUser({ userName, displayName, roles, addons });
        enqueueSnackbar(result.message, { variant: 'success' });
        return result.user;
      }
      catch (error) {
        handleError(error);
      }
    }
    finally { releaseMask(); }
  }

  /**
   * updateNameDialog: Update user name
   */
  const updateName = async (userName: string, displayName: string) => {
    console.log('updateUserName', userName, displayName);
    setMask();
    try {
      try {
        const result = await userService.updateUserDisplayName({ userName, displayName });
        const newUser = result.user;
        enqueueSnackbar(result.message, { variant: 'success' });
        if (users && newUser) {
          setUsers(prev => prev.map(us => us.name === newUser.name ? new User(newUser) : us));
        }
        return newUser;
      }
      catch (error) {
        handleError(error);
      }
    }
    finally { releaseMask(); }
  }

  /**
   * updateRoleDialog: Update user 
   */
  const updateRole = async (userName: string, roles: string[]) => {
    console.log('updateRole', userName, roles);
    setMask();
    try {
      try {
        const result = await userService.updateUserRole({ userName, roles });
        const newUser = result.user;
        enqueueSnackbar(result.message, { variant: 'success' });
        if (users && newUser) {
          setUsers(prev => prev.map(us => us.name === newUser.name ? new User(newUser) : us));
        }
        return newUser;
      }
      catch (error) {
        handleError(error);
      }
    }
    finally { releaseMask(); }
  }

  /**
   * DeleteDialog: Delete user 
   */
  const deleteUser = async (userName: string) => {
    console.log('deleteUser');
    setMask();
    try {
      try {
        const result = await userService.deleteUser({ userName });
        enqueueSnackbar(result.message, { variant: 'success' });
        setUsers(users.filter((u) => u.name !== userName));
        return result;
      }
      catch (error) {
        handleError(error);
      }
    }
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
  const templateService = useTemplateService();
  const { handleError } = useHandleError();

  const getUserAddonTemplates = () => {
    console.log('getUserAddonTemplates');
    return templateService.getUserAddonTemplates({})
      .then(result => { setTemplates(result.items.sort((a, b) => (a.name < b.name) ? -1 : 1)); })
      .catch(error => { handleError(error) });
  }

  return ({
    templates,
    getUserAddonTemplates,
  });
}

/**
* error handler
*/
const useHandleError = () => {
  const { enqueueSnackbar } = useSnackbar();
  const navigate = useNavigate();

  const handleError = (error: any) => {
    console.log('handleError', error);

    if (error instanceof ConnectError &&
      error.code === Code.Unauthenticated) {
      navigate('/signin');
    }
    const msg = error?.message;
    msg && enqueueSnackbar(msg, { variant: 'error' });
    throw error;
  }
  return { handleError }
}

/**
 * UserProvider
 */
export const UserContext = ModuleContext(useUser);
export const useUserModule = UserContext.useContext;
