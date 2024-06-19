import useUrlState from "@ahooksjs/use-url-state";
import {
  AddTwoTone,
  AdminPanelSettingsTwoTone,
  Badge,
  Clear,
  DeleteOutlined,
  Edit,
  MoreVert,
  RefreshTwoTone,
  SearchTwoTone,
  Settings,
} from "@mui/icons-material";
import {
  Box,
  Chip,
  Fab,
  IconButton,
  InputAdornment,
  ListItemIcon,
  ListItemText,
  Menu,
  MenuItem,
  Paper,
  Stack,
  TextField,
  Tooltip,
  Typography,
  useMediaQuery,
  useTheme,
} from "@mui/material";
import {
  DataGrid,
  GridColDef,
  GridRenderCellParams,
  gridClasses,
} from "@mui/x-data-grid";
import React, { useEffect } from "react";
import { useLogin } from "../../components/LoginProvider";
import { User, UserAddon } from "../../proto/gen/dashboard/v1alpha1/user_pb";
import { EllipsisTypography } from "../atoms/EllipsisTypography";
import { PasswordDialogContext } from "../organisms/PasswordDialog";
import { RoleChangeDialogContext } from "../organisms/RoleChangeDialog";
import {
  UserCreateConfirmDialogContext,
  UserCreateDialogContext,
  UserDeleteDialogContext,
  UserInfoDialogContext,
} from "../organisms/UserActionDialog";
import { UserAddonChangeDialogContext } from "../organisms/UserAddonsChangeDialog";
import {
  hasAdminForRole,
  hasPrivilegedRole,
  isAdminRole,
  isPrivilegedRole,
  useUserModule,
} from "../organisms/UserModule";
import { UserNameChangeDialogContext } from "../organisms/UserNameChangeDialog";
import { PageTemplate } from "../templates/PageTemplate";

/**
 * view
 */
const UserMenu: React.VFC<{ user: User }> = ({ user: us }) => {
  const { loginUser } = useLogin();
  const [anchorEl, setAnchorEl] = React.useState<null | HTMLElement>(null);
  const roleChangeDialogDispatch = RoleChangeDialogContext.useDispatch();
  const userDeleteDialogDispatch = UserDeleteDialogContext.useDispatch();
  const userNameChangeDispatch = UserNameChangeDialogContext.useDispatch();
  const userAddonChangeDispatch = UserAddonChangeDialogContext.useDispatch();

  return (
    <>
      <Box>
        <IconButton
          color="inherit"
          disabled={loginUser?.name === us.name}
          onClick={(e) => setAnchorEl(e.currentTarget)}
        >
          <MoreVert fontSize="small" />
        </IconButton>
        <Menu
          anchorEl={anchorEl}
          open={Boolean(anchorEl)}
          onClose={() => setAnchorEl(null)}
        >
          <MenuItem
            onClick={() => {
              userNameChangeDispatch(true, { user: us });
              setAnchorEl(null);
            }}
          >
            <ListItemIcon>
              <Badge fontSize="small" />
            </ListItemIcon>
            <ListItemText>Change DisplayName...</ListItemText>
          </MenuItem>
          <MenuItem
            onClick={() => {
              roleChangeDialogDispatch(true, { user: us });
              setAnchorEl(null);
            }}
          >
            <ListItemIcon>
              <AdminPanelSettingsTwoTone fontSize="small" />
            </ListItemIcon>
            <ListItemText>Change Role...</ListItemText>
          </MenuItem>
          <MenuItem
            onClick={() => {
              userAddonChangeDispatch(true, { user: us });
              setAnchorEl(null);
            }}
          >
            <ListItemIcon>
              <Settings fontSize="small" />
            </ListItemIcon>
            <ListItemText>Change Addons...</ListItemText>
          </MenuItem>
          <MenuItem
            onClick={() => {
              userDeleteDialogDispatch(true, { user: us });
              setAnchorEl(null);
            }}
          >
            <ListItemIcon>
              <DeleteOutlined fontSize="small" />
            </ListItemIcon>
            <ListItemText>Remove User...</ListItemText>
          </MenuItem>
        </Menu>
      </Box>
    </>
  );
};

export type UserDataGridProp = {
  users: User[];
};

export const UserDataGrid: React.FC<UserDataGridProp> = ({ users }) => {
  const userNameChangeDispatch = UserNameChangeDialogContext.useDispatch();
  const roleChangeDialogDispatch = RoleChangeDialogContext.useDispatch();
  const userAddonChangeDispatch = UserAddonChangeDialogContext.useDispatch();

  const theme = useTheme();
  const isUpSM = useMediaQuery(theme.breakpoints.up("sm"), { noSsr: true });

  const columns: GridColDef[] = [
    {
      field: "id",
      headerName: "ID",
      flex: 0.8,
      renderCell: (params: GridRenderCellParams<any, string>) => (
        <EllipsisTypography variant="body2">{params.value}</EllipsisTypography>
      ),
    },
    {
      field: "status",
      headerName: "Status",
      flex: 0.6,
      minWidth: 80,
      renderCell: (params: GridRenderCellParams<any, string>) => (
        <Chip
          color={params.value === "Active" ? "success" : "error"}
          variant="outlined"
          size="small"
          label={params.value}
        />
      ),
    },
    { field: "authType", headerName: "Auth Type", flex: 0.8 },
    {
      field: "displayName",
      headerName: "Display Name",
      flex: 0.8,
      renderCell: (params: GridRenderCellParams<any, string>) => (
        <>
          {params.hasFocus ? (
            <Stack direction="row" alignItems="center" spacing={1}>
              <Box
                component="div"
                sx={{
                  justifyContent: "flex-end",
                  textOverflow: "ellipsis",
                  overflow: "hidden",
                }}
              >
                <Typography variant="body2">{params.value}</Typography>
              </Box>
              <IconButton
                size="small"
                onClick={() =>
                  userNameChangeDispatch(true, { user: params.row })
                }
              >
                <Edit />
              </IconButton>
            </Stack>
          ) : (
            <EllipsisTypography variant="body2">
              {params.value}
            </EllipsisTypography>
          )}
        </>
      ),
    },
    {
      field: "roles",
      headerName: "Roles",
      flex: 0.8,
      renderCell: (params: GridRenderCellParams<any, string[]>) => (
        <Box
          component="div"
          sx={{
            justifyContent: "flex-end",
            textOverflow: "ellipsis",
            overflow: "hidden",
          }}
        >
          {params.value?.map((v, i) => (
            <Chip
              color={
                isPrivilegedRole(v)
                  ? "error"
                  : isAdminRole(v)
                  ? "warning"
                  : "default"
              }
              variant="outlined"
              size="small"
              key={i}
              label={v}
            />
          ))}
          {params.hasFocus && (
            <IconButton
              size="small"
              onClick={() =>
                roleChangeDialogDispatch(true, { user: params.row })
              }
            >
              <Edit />
            </IconButton>
          )}
        </Box>
      ),
    },
    {
      field: "addons",
      headerName: "Addons",
      valueGetter: (addons: UserAddon[]) => addons.map((v) => v.template),
      renderCell: (params: GridRenderCellParams<any, string[]>) => (
        <Stack>
          {params.value?.map((v, i) => (
            <Typography key={i} variant="body2">
              {v}
            </Typography>
          ))}
          {params.hasFocus && (
            <Stack direction="row" alignItems="center" spacing={2}>
              <IconButton
                size="small"
                onClick={() =>
                  userAddonChangeDispatch(true, { user: params.row })
                }
              >
                <Edit />
              </IconButton>
              <Box flex={1} />
            </Stack>
          )}
        </Stack>
      ),
      flex: 1,
    },
    {
      field: "actions",
      type: "actions",
      getActions: () => [],
      flex: 0.2,
      renderCell: (params: GridRenderCellParams<any, string[]>) => (
        <UserMenu user={params.row} />
      ),
    },
  ];

  return (
    <>
      <div style={{ width: "100%", minHeight: 100 }}>
        <DataGrid
          autoHeight={true}
          rows={users.map((v) => ({ ...v, id: v.name }))}
          columns={columns}
          getRowHeight={() => "auto"}
          sx={{
            [`& .${gridClasses.cell}`]: {
              py: 1,
            },
          }}
          initialState={{
            columns: {
              columnVisibilityModel: {
                addons: isUpSM,
                displayName: isUpSM,
                authType: isUpSM,
              },
            },
            pagination: { paginationModel: { pageSize: 10 } },
          }}
          pageSizeOptions={[10, 50, 100]}
        />
      </div>
    </>
  );
};

const UserList: React.VFC = () => {
  const hooks = useUserModule();
  const { loginUser } = useLogin();
  const userCreateDialogDispatch = UserCreateDialogContext.useDispatch();
  const userInfoDialogDispatch = UserInfoDialogContext.useDispatch();

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

  const filterRoles: string[] =
    typeof urlParam.filterRoles === "string"
      ? [urlParam.filterRoles]
      : urlParam.filterRoles;

  useEffect(() => {
    if (
      loginUser &&
      !hasPrivilegedRole(loginUser.roles) &&
      filterRoles.length === 0
    ) {
      setUrlParam({
        filterRoles: hooks.existingRoles.filter((v) =>
          hasAdminForRole(loginUser.roles, v)
        ),
      });
    }
  }, []);

  const isUserMatchedToFilterRoles = (user: User) => {
    for (const v of user.roles) {
      for (const f of filterRoles) {
        if (v === f) {
          return true;
        }
      }
    }
    return false;
  };

  useEffect(() => {
    hooks.getUsers();
  }, []); // eslint-disable-line

  return (
    <>
      <Paper sx={{ minWidth: 320, maxWidth: 1200, mb: 1, p: 1 }}>
        <Stack direction="row" alignItems="center" spacing={2}>
          <TextField
            InputProps={
              urlParam.search !== ""
                ? {
                    startAdornment: (
                      <InputAdornment position="start">
                        <SearchTwoTone />
                      </InputAdornment>
                    ),
                    endAdornment: (
                      <InputAdornment position="end">
                        <IconButton
                          size="small"
                          tabIndex={-1}
                          onClick={() => {
                            setUrlParam({ search: "" });
                          }}
                        >
                          <Clear />
                        </IconButton>
                      </InputAdornment>
                    ),
                  }
                : {
                    startAdornment: (
                      <InputAdornment position="start">
                        <SearchTwoTone />
                      </InputAdornment>
                    ),
                  }
            }
            placeholder="Search"
            size="small"
            value={urlParam.search}
            onChange={(e) => setUrlParam({ search: e.target.value })}
            sx={{ flexGrow: 0.5 }}
          />
          <Box sx={{ flexGrow: 1 }} />
          <Tooltip title="Refresh" placement="top">
            <IconButton
              color="inherit"
              onClick={() => {
                hooks.getUsers();
              }}
            >
              <RefreshTwoTone />
            </IconButton>
          </Tooltip>
          <Tooltip title="Add new user" placement="top">
            <Fab
              size="small"
              color="primary"
              onClick={() => userCreateDialogDispatch(true)}
              sx={{ flexShrink: 0 }}
            >
              <AddTwoTone />
            </Fab>
          </Tooltip>
        </Stack>
      </Paper>
      <UserDataGrid
        users={hooks.users
          .filter(
            (us) =>
              urlParam.search === "" ||
              Boolean(us.name.match(urlParam.search)) ||
              Boolean(us.status.match(urlParam.search)) ||
              Boolean(us.authType.match(urlParam.search)) ||
              Boolean(us.displayName.match(urlParam.search)) ||
              Boolean(
                us.roles.filter((v) => v.match(urlParam.search)).length > 0
              ) ||
              Boolean(
                us.addons.filter((v) => v.template.match(urlParam.search))
                  .length > 0
              )
          )
          .filter(
            (us) => filterRoles.length == 0 || isUserMatchedToFilterRoles(us)
          )}
      />
    </>
  );
};

export const UserPage: React.VFC = () => {
  console.log("UserPage");

  return (
    <PageTemplate title="Users">
      <PasswordDialogContext.Provider>
        <UserCreateConfirmDialogContext.Provider>
          <UserCreateDialogContext.Provider>
            <RoleChangeDialogContext.Provider>
              <UserAddonChangeDialogContext.Provider>
                <UserDeleteDialogContext.Provider>
                  <UserInfoDialogContext.Provider>
                    <UserList />
                  </UserInfoDialogContext.Provider>
                </UserDeleteDialogContext.Provider>
              </UserAddonChangeDialogContext.Provider>
            </RoleChangeDialogContext.Provider>
          </UserCreateDialogContext.Provider>
        </UserCreateConfirmDialogContext.Provider>
      </PasswordDialogContext.Provider>
    </PageTemplate>
  );
};
