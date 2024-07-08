import {
  AddTwoTone,
  AdminPanelSettingsTwoTone,
  Badge as BadgeIcon,
  DeleteOutlined,
  DeleteSweepOutlined,
  Edit,
  ExpandLess,
  ExpandMore,
  FilterListOff,
  LockOpenOutlined,
  LockOutlined,
  MoreVert,
  Notifications,
  OpenInNewTwoTone,
  PushPinOutlined,
  RefreshTwoTone,
  Settings,
  Web,
} from "@mui/icons-material";
import {
  Badge,
  Box,
  Chip,
  Collapse,
  Divider,
  Fab,
  Grid,
  IconButton,
  ListItemIcon,
  ListItemText,
  Menu,
  MenuItem,
  Paper,
  Stack,
  Tooltip,
  Typography,
  styled,
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
import {
  DeletePolicy,
  User,
  UserAddon,
} from "../../proto/gen/dashboard/v1alpha1/user_pb";
import { EllipsisTypography } from "../atoms/EllipsisTypography";
import { NameAvatar } from "../atoms/NameAvatar";
import { SearchTextField } from "../atoms/SearchTextField";
import { SelectableChip } from "../atoms/SelectableChips";
import { PasswordDialogContext } from "../organisms/PasswordDialog";
import { RoleChangeDialogContext } from "../organisms/RoleChangeDialog";
import {
  UserCreateConfirmDialogContext,
  UserCreateDialogContext,
  UserDeleteDialogContext,
} from "../organisms/UserActionDialog";
import { UserAddonChangeDialogContext } from "../organisms/UserAddonsChangeDialog";
import { UserDeletePolicyChangeDialogContext } from "../organisms/UserDeletePolicyChangeDialog";
import { UserInfoDialogContext } from "../organisms/UserInfoDialog";
import {
  isAdminRole,
  isPrivilegedRole,
  useUserModule,
} from "../organisms/UserModule";
import { UserNameChangeDialogContext } from "../organisms/UserNameChangeDialog";
import { WorkspaceInfoDialogContext } from "../organisms/WorkspaceInfoDialog";
import { PageTemplate } from "../templates/PageTemplate";

/**
 * view
 */
const RotatingRefreshTwoTone = styled(RefreshTwoTone)({
  animation: "rotatingRefresh 1s linear infinite",
  "@keyframes rotatingRefresh": {
    to: {
      transform: "rotate(2turn)",
    },
  },
});

const UserMenu: React.VFC<{ user: User }> = ({ user: us }) => {
  const { loginUser } = useLogin();
  const [anchorEl, setAnchorEl] = React.useState<null | HTMLElement>(null);
  const roleChangeDialogDispatch = RoleChangeDialogContext.useDispatch();
  const userDeleteDialogDispatch = UserDeleteDialogContext.useDispatch();
  const userNameChangeDispatch = UserNameChangeDialogContext.useDispatch();
  const userAddonChangeDispatch = UserAddonChangeDialogContext.useDispatch();
  const userDeletePolicyChangeDispatch =
    UserDeletePolicyChangeDialogContext.useDispatch();

  return (
    <>
      <Box>
        <IconButton
          color="inherit"
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
              window.open(`/#/event?user=${us.name}`);
            }}
          >
            <ListItemIcon>
              <Notifications fontSize="small" />
            </ListItemIcon>
            <ListItemText>
              Open Events...
              {
                <OpenInNewTwoTone
                  fontSize="inherit"
                  sx={{ position: "relative", top: "0.2em" }}
                />
              }
            </ListItemText>
          </MenuItem>
          <MenuItem
            onClick={() => {
              window.open(`/#workspace?user=${us.name}`);
            }}
          >
            <ListItemIcon>
              <Web fontSize="small" />
            </ListItemIcon>
            <ListItemText>
              Open Workspaces...
              {
                <OpenInNewTwoTone
                  fontSize="inherit"
                  sx={{ position: "relative", top: "0.2em" }}
                />
              }
            </ListItemText>
          </MenuItem>
          <Divider />
          <MenuItem
            onClick={() => {
              userNameChangeDispatch(true, { user: us });
              setAnchorEl(null);
            }}
          >
            <ListItemIcon>
              <BadgeIcon fontSize="small" />
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
          <Divider />
          <MenuItem
            onClick={() => {
              userDeletePolicyChangeDispatch(true, { user: us });
              setAnchorEl(null);
            }}
          >
            <ListItemIcon>
              {us.deletePolicy === DeletePolicy.keep ? (
                <LockOutlined fontSize="small" />
              ) : (
                <LockOpenOutlined fontSize="small" />
              )}
            </ListItemIcon>
            <ListItemText>Manage Delete Policy...</ListItemText>
          </MenuItem>
          <Divider />
          <MenuItem
            onClick={() => {
              userDeleteDialogDispatch(true, { user: us });
              setAnchorEl(null);
            }}
            disabled={us.deletePolicy === DeletePolicy.keep}
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
  const userInfoDispatch = UserInfoDialogContext.useDispatch();
  const userDeletePolicyChangeDispatch =
    UserDeletePolicyChangeDialogContext.useDispatch();

  const theme = useTheme();
  const isUpSM = useMediaQuery(theme.breakpoints.up("sm"), { noSsr: true });

  const columns: GridColDef[] = [
    {
      field: "avator",
      headerName: "Avator",
      type: "singleSelect",
      width: 80,
      sortable: false,
      renderCell: (params: GridRenderCellParams<any, string>) => (
        <NameAvatar
          name={params.row.id}
          sx={{ width: 32, height: 32 }}
          onClick={() => {
            userInfoDispatch(true, { userName: params.row.id });
          }}
        />
      ),
    },
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
      field: "deletePolicy",
      headerName: "Delete Policy",
      flex: 0.5,
      renderCell: (params: GridRenderCellParams<any, DeletePolicy>) => (
        <>
          {params.hasFocus ? (
            <Stack alignItems="flex-start" spacing={1}>
              <Chip
                size="small"
                variant="outlined"
                label={params.value === DeletePolicy.keep ? "keep" : "delete"}
                avatar={
                  params.value === DeletePolicy.keep ? (
                    <PushPinOutlined />
                  ) : (
                    <DeleteSweepOutlined />
                  )
                }
              />
              <IconButton
                size="small"
                onClick={() =>
                  userDeletePolicyChangeDispatch(true, { user: params.row })
                }
              >
                <Edit />
              </IconButton>
            </Stack>
          ) : (
            <Chip
              size="small"
              variant="outlined"
              label={params.value === DeletePolicy.keep ? "keep" : "delete"}
              avatar={
                params.value === DeletePolicy.keep ? (
                  <PushPinOutlined />
                ) : (
                  <DeleteSweepOutlined />
                )
              }
            />
          )}
        </>
      ),
    },
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
      sortComparator: (v1: string[], v2: string[]) => {
        if (v1.length !== v2.length) return v1.length - v2.length;
        const v1str = v1.sort().join();
        const v2str = v2.sort().join();
        return v1str.length - v2str.length;
      },
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
      sortComparator: (v1: UserAddon[], v2: UserAddon[]) => {
        if (v1.length !== v2.length) return v1.length - v2.length;
        const v1str = v1
          .map((v) => v.template)
          .sort()
          .join();
        const v2str = v2
          .map((v) => v.template)
          .sort()
          .join();
        return v1str.length - v2str.length;
      },
      renderCell: (params: GridRenderCellParams<any, UserAddon[]>) => (
        <Stack>
          {params.value?.map((v, i) => (
            <Tooltip
              arrow
              title={
                Object.keys(v.vars).length > 0 && (
                  <Stack>
                    {Object.keys(v.vars).map((k) => (
                      <Typography
                        key={i}
                        variant="body2"
                      >{`${k} = ${v.vars[k]}`}</Typography>
                    ))}
                  </Stack>
                )
              }
              key={i}
            >
              <Typography key={i} variant="body2">
                - {v.template}
              </Typography>
            </Tooltip>
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
                roles: isUpSM,
                addons: isUpSM,
                displayName: isUpSM,
                authType: false,
                deletePolicy: false,
              },
            },
            pagination: { paginationModel: { pageSize: 10 } },
          }}
          pageSizeOptions={[10, 50, 100]}
          onRowDoubleClick={(params) => {
            userInfoDispatch(true, { userName: params.row.id });
          }}
        />
      </div>
    </>
  );
};

const UserList: React.VFC = () => {
  const {
    search,
    setSearch,
    users,
    getUsers,
    filterRoles,
    appendFilterRoles,
    removeFilterRoles,
    existingRoles,
    applyAdminRoleFilter,
  } = useUserModule();
  const userCreateDialogDispatch = UserCreateDialogContext.useDispatch();
  const [isLoading, setIsLoading] = React.useState(false);
  const [showFilter, setShowFilter] = React.useState<boolean>(false);

  const searchRegExp = new RegExp(search, "i");

  useEffect(applyAdminRoleFilter, []);

  return (
    <>
      <Paper sx={{ minWidth: 320, maxWidth: 1200, mb: 1, px: 2, py: 1 }}>
        <Stack direction="row" alignItems="center" spacing={2}>
          <SearchTextField search={search} setSearch={setSearch} />
          <Box sx={{ flexGrow: 1 }} />
          <Tooltip title="Refresh" placement="top">
            <IconButton
              color="inherit"
              onClick={() => {
                setIsLoading(true);
                setTimeout(() => {
                  setIsLoading(false);
                }, 1000);
                if (!isLoading) getUsers();
              }}
            >
              {isLoading ? <RotatingRefreshTwoTone /> : <RefreshTwoTone />}
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
        <Box
          component="div"
          sx={{
            justifyContent: "flex-end",
            textOverflow: "ellipsis",
            overflow: "hidden",
            pt: 1,
          }}
        >
          <IconButton
            size="small"
            color="inherit"
            onClick={() => {
              setShowFilter(!showFilter);
            }}
          >
            {showFilter ? <ExpandLess /> : <ExpandMore />}
          </IconButton>
          <Typography color="text.secondary" variant="caption">
            Filter by Roles
          </Typography>
          {filterRoles.length > 0 && (
            <IconButton
              size="small"
              color="inherit"
              onClick={() => {
                removeFilterRoles();
              }}
            >
              <Tooltip arrow title="Clear Role Filter" placement="top">
                <Badge badgeContent={filterRoles.length} color="error">
                  <FilterListOff />
                </Badge>
              </Tooltip>
            </IconButton>
          )}
          <Collapse in={showFilter} timeout="auto" unmountOnExit>
            <Grid container>
              {existingRoles.map((v, i) => (
                <SelectableChip
                  key={v}
                  label={v}
                  sx={{ m: 0.1 }}
                  color={
                    isPrivilegedRole(v)
                      ? "error"
                      : v.endsWith("-admin")
                      ? "warning"
                      : "primary"
                  }
                  checked={filterRoles?.includes(v)}
                  onChecked={(checked) => {
                    checked ? appendFilterRoles(v) : removeFilterRoles(v);
                  }}
                />
              ))}
            </Grid>
          </Collapse>
        </Box>
      </Paper>
      <UserDataGrid
        users={users
          .filter(
            (us) =>
              search === "" ||
              Boolean(us.name.match(searchRegExp)) ||
              Boolean(us.status.match(searchRegExp)) ||
              Boolean(us.authType.match(searchRegExp)) ||
              Boolean(us.displayName.match(searchRegExp)) ||
              Boolean(
                us.roles.filter((v) => v.match(searchRegExp)).length > 0
              ) ||
              Boolean(
                us.addons.filter((v) => v.template.match(searchRegExp)).length >
                  0
              )
          )
          .filter((us) => {
            if (filterRoles.length === 0) return true;
            for (const role of us.roles) {
              if (filterRoles.includes(role)) {
                return true;
              }
            }
            return false;
          })}
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
                  <UserDeletePolicyChangeDialogContext.Provider>
                    <WorkspaceInfoDialogContext.Provider>
                      <UserInfoDialogContext.Provider>
                        <UserList />
                      </UserInfoDialogContext.Provider>
                    </WorkspaceInfoDialogContext.Provider>
                  </UserDeletePolicyChangeDialogContext.Provider>
                </UserDeleteDialogContext.Provider>
              </UserAddonChangeDialogContext.Provider>
            </RoleChangeDialogContext.Provider>
          </UserCreateDialogContext.Provider>
        </UserCreateConfirmDialogContext.Provider>
      </PasswordDialogContext.Provider>
    </PageTemplate>
  );
};
