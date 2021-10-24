import { useSnackbar } from "notistack";
import { useState } from "react";
import { useHistory } from "react-router-dom";
import { UserApiFactory, User } from "../../api/dashboard/v1alpha1";
import { ModuleContext } from "../../components/ContextProvider";
import { useProgress } from "../../components/ProgressProvider";

/**
 * hooks
 */
const useUser = () => {
  console.log('useUserModule');

  const { enqueueSnackbar } = useSnackbar();
  const { setMask, releaseMask } = useProgress();
  const history = useHistory();
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
  const createUser = async (id: string, displayName: string, role?: string) => {
    console.log('addUser');
    try {
      setMask();
      const result = await restUser.postUser({ id, displayName, role });
      enqueueSnackbar(result.data.message, { variant: 'success' });
      getUsers();
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

  /**
   * error handler
   */
  const handleError = (error: any) => {
    console.log(error);
    if (error?.response?.status === 401) {
      history.push('/signin');
    }
    const msg = error?.response?.data?.message || error?.message;
    msg && enqueueSnackbar(msg, { variant: 'error' });
    throw error;
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
 * UserProvider
 */
export const UserContext = ModuleContext(useUser);
export const useUserModule = UserContext.useContext;
