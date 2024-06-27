import useUrlState from "@ahooksjs/use-url-state";
import { useSnackbar } from "notistack";
import { useEffect, useState } from "react";
import { ModuleContext } from "../../components/ContextProvider";
import { useHandleError, useLogin } from "../../components/LoginProvider";
import { useProgress } from "../../components/ProgressProvider";
import { Template } from "../../proto/gen/dashboard/v1alpha1/template_pb";
import { User, UserAddon } from "../../proto/gen/dashboard/v1alpha1/user_pb";
import {
  useTemplateService,
  useUserService,
} from "../../services/DashboardServices";

export const PrivilegedRole = "cosmo-admin";

const AdminRoleSufix = "-admin";

export const isPrivilegedRole = (role: string) => {
  return role === PrivilegedRole;
};

export const isAdminRole = (role: string) => {
  return role.endsWith(AdminRoleSufix);
};

export const hasPrivilegedRole = (roles: string[]) => {
  return roles.includes(PrivilegedRole);
};

export const isAdminUser = (user?: User) => {
  if (user && user.roles) {
    if (hasPrivilegedRole(user.roles)) {
      return true;
    }
    for (const role of user.roles) {
      if (isAdminRole(role)) {
        return true;
      }
    }
  }
  return false;
};

export const excludeAdminRolePrefix = (role: string): string => {
  // given "xxx-admin" return "xxx"
  return role.endsWith(AdminRoleSufix)
    ? role.slice(0, -AdminRoleSufix.length)
    : role;
};

const hasAdminForRole = (myRoles: string[], userrole: string) => {
  for (const myRole of myRoles) {
    if (myRole == userrole) {
      return true;
    }
    if (
      isAdminRole(myRole) &&
      userrole.startsWith(excludeAdminRolePrefix(myRole))
    ) {
      return true;
    }
  }
  return false;
};

function filterUsersByRoles(users: User[], myRoles: string[]) {
  return hasPrivilegedRole(myRoles)
    ? users
    : users.filter((u) => {
        for (const userRole of u.roles) {
          if (hasAdminForRole(myRoles, userRole)) {
            return true;
          }
        }
        return false;
      });
}

export function setUserStateFuncFilteredByLoginUserRole(
  users: User[],
  loginUser?: User
) {
  const f = (prev: User[]) => {
    const newUsers = users.sort((a, b) => (a.name < b.name ? -1 : 1));
    const roleFilteredNewUsers = filterUsersByRoles(
      newUsers,
      loginUser?.roles || []
    );
    return JSON.stringify(prev) === JSON.stringify(roleFilteredNewUsers)
      ? prev
      : roleFilteredNewUsers;
  };
  return f;
}

/**
 * hooks
 */
const useUser = () => {
  console.log("useUserModule");

  const { loginUser } = useLogin();
  const { enqueueSnackbar } = useSnackbar();
  const { setMask, releaseMask } = useProgress();
  const { handleError } = useHandleError();
  const [users, setUsers] = useState<User[]>([]);
  const userService = useUserService();
  const [existingRoles, setExistingRoles] = useState<string[]>([]);

  const [urlParam, setUrlParam] = useUrlState(
    {
      search: "",
      filterRoles: [],
    },
    {
      parseOptions: { arrayFormat: "comma" },
      stringifyOptions: { arrayFormat: "comma", skipEmptyString: true },
    }
  );

  const search: string = urlParam.search || "";
  const setSearch = (word: string) => setUrlParam({ search: word });

  const filterRoles: string[] =
    typeof urlParam.filterRoles === "string"
      ? [urlParam.filterRoles]
      : urlParam.filterRoles;

  const appendFilterRoles = (role: string) => {
    setUrlParam((prev) => {
      if (typeof prev.filterRoles === "string") {
        return prev.filterRoles === role
          ? prev
          : { filterRoles: [prev.filterRoles, role] };
      }
      return prev.filterRoles.includes(role)
        ? prev
        : {
            filterRoles: [...filterRoles, role],
          };
    });
  };

  const removeFilterRoles = (role?: string) => {
    if (role) {
      setUrlParam((prev) => {
        if (typeof prev.filterRoles === "string") {
          return prev.filterRoles === role ? { filterRoles: [] } : prev;
        }
        return prev.filterRoles.includes(role)
          ? {
              filterRoles: prev.filterRoles.filter((v: string) => v !== role),
            }
          : prev;
      });
    } else {
      setUrlParam({ filterRoles: [] });
      return;
    }
  };

  useEffect(() => {
    // init fetch users
    getUsers().then((users) => {
      if (
        loginUser &&
        users &&
        !hasPrivilegedRole(loginUser.roles) &&
        filterRoles.length === 0
      ) {
        setUrlParam({
          filterRoles: [
            ...new Set(users.map((user) => user.roles).flat()),
          ].filter((v) => hasAdminForRole(loginUser.roles, v)),
        });
      }
    });
  }, []);

  /**
   * UserList: user list
   */
  const getUsers = async (): Promise<User[] | undefined> => {
    console.log("getUsers");
    setMask();
    try {
      const result = await userService.getUsers({});
      setUsers(result.items?.sort((a, b) => (a.name < b.name ? -1 : 1)));
      updateExistingRoles(result.items);
      return result.items;
    } catch (error) {
      handleError(error);
    } finally {
      releaseMask();
    }
  };

  const updateExistingRoles = (users: User[]) => {
    setExistingRoles(
      [...new Set(users.map((user) => user.roles).flat())].sort((a, b) =>
        a < b ? -1 : 1
      )
    );
  };

  /**
   * CreateDialog: Add user
   */
  const createUser = async (
    userName: string,
    displayName: string,
    authType: string,
    roles?: string[],
    addons?: UserAddon[]
  ) => {
    console.log("addUser");
    setMask();
    try {
      const result = await userService.createUser({
        userName,
        displayName,
        authType,
        roles,
        addons,
      });
      enqueueSnackbar(result.message, { variant: "success" });
      return result.user;
    } catch (error) {
      handleError(error);
    } finally {
      releaseMask();
    }
  };

  /**
   * updateNameDialog: Update user name
   */
  const updateName = async (userName: string, displayName: string) => {
    console.log("updateUserName", userName, displayName);
    setMask();
    try {
      const result = await userService.updateUserDisplayName({
        userName,
        displayName,
      });
      const newUser = result.user;
      enqueueSnackbar(result.message, { variant: "success" });
      if (users && newUser) {
        setUsers((prev) =>
          prev.map((us) => (us.name === newUser.name ? new User(newUser) : us))
        );
      }
      return newUser;
    } catch (error) {
      handleError(error);
    } finally {
      releaseMask();
    }
  };

  /**
   * updateRoleDialog: Update user
   */
  const updateRole = async (userName: string, roles: string[]) => {
    console.log("updateRole", userName, roles);
    setMask();
    try {
      const result = await userService.updateUserRole({ userName, roles });
      const newUser = result.user;
      enqueueSnackbar(result.message, { variant: "success" });
      if (users && newUser) {
        const newUsers = users.map((us) =>
          us.name === newUser.name ? new User(newUser) : us
        );
        setUsers(newUsers);
        updateExistingRoles(newUsers);
      }
      return newUser;
    } catch (error) {
      handleError(error);
    } finally {
      releaseMask();
    }
  };

  /**
   * updateAddonsDialog: Update user
   */
  const updateAddons = async (userName: string, addons: UserAddon[]) => {
    console.log("updateAddons", userName, addons);
    setMask();
    try {
      const result = await userService.updateUserAddons({ userName, addons });
      const newUser = result.user;
      enqueueSnackbar(result.message, { variant: "success" });
      if (users && newUser) {
        const newUsers = users.map((us) =>
          us.name === newUser.name ? new User(newUser) : us
        );
        setUsers(newUsers);
      }
      setTimeout(() => {
        getUsers();
      }, 1000);
      return newUser;
    } catch (error) {
      handleError(error);
    } finally {
      releaseMask();
    }
  };

  /**
   * DeleteDialog: Delete user
   */
  const deleteUser = async (userName: string) => {
    console.log("deleteUser");
    setMask();
    try {
      try {
        const result = await userService.deleteUser({ userName });
        enqueueSnackbar(result.message, { variant: "success" });
        setUsers(users.filter((u) => u.name !== userName));
        return result;
      } catch (error) {
        handleError(error);
      }
    } finally {
      releaseMask();
    }
  };

  return {
    search,
    setSearch,
    filterRoles,
    appendFilterRoles,
    removeFilterRoles,
    existingRoles,
    users,
    getUsers,
    createUser,
    updateName,
    updateRole,
    updateAddons,
    deleteUser,
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

  const getAllUserAddonTemplates = () => {
    console.log("getUserAddonTemplates");
    return templateService
      .getUserAddonTemplates({})
      .then((result) => {
        setTemplates(result.items.sort((a, b) => (a.name < b.name ? -1 : 1)));
      })
      .catch((error) => {
        handleError(error);
      });
  };

  const getUserAddonTemplates = () => {
    console.log("getUserAddonTemplates");
    return templateService
      .getUserAddonTemplates({ useRoleFilter: true })
      .then((result) => {
        setTemplates(result.items.sort((a, b) => (a.name < b.name ? -1 : 1)));
      })
      .catch((error) => {
        handleError(error);
      });
  };

  return {
    templates,
    getUserAddonTemplates,
    getAllUserAddonTemplates,
  };
};

/**
 * UserProvider
 */
export const UserContext = ModuleContext(useUser);
export const useUserModule = UserContext.useContext;
