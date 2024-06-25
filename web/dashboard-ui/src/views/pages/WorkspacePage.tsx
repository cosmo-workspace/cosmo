import {
  AddTwoTone,
  CheckCircleOutlined,
  Clear,
  CoPresentTwoTone,
  ContentCopy,
  DeleteTwoTone,
  EditTwoTone,
  ErrorOutline,
  ExpandLessTwoTone,
  ExpandMoreTwoTone,
  InfoOutlined,
  KeyboardArrowDownTwoTone,
  KeyboardArrowUpTwoTone,
  LockOutlined,
  MoreVertTwoTone,
  OpenInNewTwoTone,
  Person,
  PlayCircleFilledWhiteTwoTone,
  PublicOutlined,
  RefreshTwoTone,
  SearchTwoTone,
  StopCircleOutlined,
  StopCircleTwoTone,
  UnfoldLessOutlined,
  UnfoldMoreOutlined,
  VerifiedUserOutlined,
  WebTwoTone,
} from "@mui/icons-material";
import {
  Avatar,
  Badge,
  Box,
  Card,
  CardContent,
  CardHeader,
  Chip,
  CircularProgress,
  Fab,
  Grid,
  IconButton,
  InputAdornment,
  Link,
  ListItemIcon,
  ListItemText,
  Menu,
  MenuItem,
  Paper,
  Stack,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  TextField,
  Tooltip,
  Typography,
  styled,
  useMediaQuery,
  useTheme,
} from "@mui/material";
import copy from "copy-to-clipboard";
import { useSnackbar } from "notistack";
import React, { useRef, useState } from "react";
import { useLogin } from "../../components/LoginProvider";
import { Event } from "../../proto/gen/dashboard/v1alpha1/event_pb";
import { User } from "../../proto/gen/dashboard/v1alpha1/user_pb";
import { EventsDataGrid } from "../atoms/EventsDataGrid";
import { NameAvatar } from "../atoms/NameAvatar";
import { EventDetailDialogContext } from "../organisms/EventDetailDialog";
import {
  NetworkRuleDeleteDialogContext,
  NetworkRuleUpsertDialogContext,
} from "../organisms/NetworkRuleActionDialog";
import { hasPrivilegedRole } from "../organisms/UserModule";
import {
  WorkspaceCreateDialogContext,
  WorkspaceDeleteDialogContext,
  WorkspaceStartDialogContext,
  WorkspaceStopDialogContext,
} from "../organisms/WorkspaceActionDialog";
import { WorkspaceInfoDialogContext } from "../organisms/WorkspaceInfoDialog";
import {
  WorkspaceContext,
  WorkspaceWrapper,
  computeStatus,
  useWorkspaceModule,
} from "../organisms/WorkspaceModule";
import { PageTemplate } from "../templates/PageTemplate";

/**
 * view
 */
const RotatingRefreshTwoTone = styled(RefreshTwoTone)({
  animation: "rotatingRefresh 2s linear infinite",
  "@keyframes rotatingRefresh": {
    to: {
      transform: "rotate(2turn)",
    },
  },
});

const StatusChip: React.VFC<{ ws: WorkspaceWrapper }> = ({ ws }) => {
  const statusLabel = computeStatus(ws);

  switch (statusLabel) {
    case "Running":
      return (
        <Chip
          variant="outlined"
          size="small"
          icon={<CheckCircleOutlined />}
          color="success"
          label={statusLabel}
        />
      );
    case "Stopped":
      return (
        <Chip
          variant="outlined"
          size="small"
          icon={<StopCircleOutlined />}
          color="error"
          label={statusLabel}
        />
      );
    case "Error":
    case "CrashLoopBackOff":
      return (
        <Chip
          variant="outlined"
          size="small"
          icon={<ErrorOutline />}
          color="error"
          label={statusLabel}
        />
      );
    default:
      return (
        <Chip
          variant="outlined"
          size="small"
          icon={
            ws.progress > 0 ? (
              ws.progress > 100 ? (
                <InfoOutlined />
              ) : (
                <CircularProgress
                  color="info"
                  size={13}
                  variant="determinate"
                  value={ws.progress}
                />
              )
            ) : (
              <CircularProgress color="info" size={13} />
            )
          }
          color="info"
          label={statusLabel}
        />
      );
  }
};

const WorkspaceMenu: React.VFC<{ workspace: WorkspaceWrapper; user: User }> = ({
  workspace,
  user,
}) => {
  const [anchorEl, setAnchorEl] = React.useState<null | HTMLElement>(null);
  const startDialogDispatch = WorkspaceStartDialogContext.useDispatch();
  const stopDialogDispatch = WorkspaceStopDialogContext.useDispatch();
  const deleteDialogDispatch = WorkspaceDeleteDialogContext.useDispatch();

  return (
    <>
      <IconButton
        color="inherit"
        onClick={(e) => {
          setAnchorEl(e.currentTarget);
          e.stopPropagation();
        }}
        disabled={workspace.readonlyFor(user)}
      >
        <MoreVertTwoTone />
      </IconButton>
      <Menu
        anchorEl={anchorEl}
        open={Boolean(anchorEl)}
        onClose={() => setAnchorEl(null)}
      >
        <MenuItem
          onClick={() => {
            setAnchorEl(null);
            startDialogDispatch(true, { workspace: workspace });
          }}
          disabled={!Boolean(workspace.name)}
        >
          <ListItemIcon>
            <PlayCircleFilledWhiteTwoTone fontSize="small" />
          </ListItemIcon>
          <ListItemText>Start workspace...</ListItemText>
        </MenuItem>
        <MenuItem
          onClick={() => {
            setAnchorEl(null);
            stopDialogDispatch(true, { workspace: workspace });
          }}
          disabled={!Boolean(workspace.name)}
        >
          <ListItemIcon>
            <StopCircleTwoTone fontSize="small" />
          </ListItemIcon>
          <ListItemText>Stop workspace...</ListItemText>
        </MenuItem>
        <MenuItem
          onClick={() => {
            setAnchorEl(null);
            deleteDialogDispatch(true, { workspace: workspace });
          }}
          disabled={!Boolean(workspace.name) || workspace.isSharedFor(user)}
        >
          <ListItemIcon>
            <DeleteTwoTone fontSize="small" />
          </ListItemIcon>
          <ListItemText>Remove workspace...</ListItemText>
        </MenuItem>
      </Menu>
    </>
  );
};

const UserSelect: React.VFC = () => {
  const { user, setUser, users, getUsers } = useWorkspaceModule();
  const [anchorEl, setAnchorEl] = React.useState<null | HTMLElement>(null);
  const chipReff = useRef(null);
  return (
    <>
      <Tooltip title="Change User" placement="top">
        <Chip
          ref={chipReff}
          label={user.name}
          avatar={
            <NameAvatar name={user.displayName} typographyVariant="body2" />
          }
          onClick={(e) => {
            e.stopPropagation();
            getUsers().then(() => setAnchorEl(chipReff.current));
          }}
          onDelete={(e) => {
            e.stopPropagation();
            getUsers().then(() => setAnchorEl(chipReff.current));
          }}
          deleteIcon={anchorEl ? <ExpandLessTwoTone /> : <ExpandMoreTwoTone />}
        />
      </Tooltip>
      <Menu
        anchorEl={anchorEl}
        open={Boolean(anchorEl)}
        onClose={() => setAnchorEl(null)}
      >
        {users.map((user, ind) => (
          <MenuItem
            key={ind}
            value={user.name}
            onClick={() => {
              setAnchorEl(null);
              setUser(user.name);
            }}
          >
            <Stack>
              <Typography>{user.name}</Typography>
              <Typography color="gray" fontSize="small">
                {user.displayName}
              </Typography>
            </Stack>
          </MenuItem>
        ))}
      </Menu>
    </>
  );
};

const NetworkRuleList: React.FC<{
  workspace: WorkspaceWrapper;
  user: User;
}> = ({ workspace, user }) => {
  const upsertDialogDispatch = NetworkRuleUpsertDialogContext.useDispatch();
  const deleteDialogDispatch = NetworkRuleDeleteDialogContext.useDispatch();
  const { enqueueSnackbar } = useSnackbar();
  const onCopy = (text: string) => {
    copy(text);
    enqueueSnackbar("Copied!", { variant: "success" });
  };
  const theme = useTheme();
  const isUpSM = useMediaQuery(theme.breakpoints.up("sm"), { noSsr: true });

  const readonly = workspace.readonlyFor(user);

  return (
    <TableContainer
      sx={{
        border: "1px solid",
        borderRadius: "4px",
        borderColor:
          theme.palette.mode === "light"
            ? "rgba(224,224,224,1)"
            : "rgba(81,81,81,1)",
      }}
    >
      <Table size="small">
        <TableHead sx={{ backgroundColor: theme.palette.background.default }}>
          <TableRow>
            <TableCell align="center">Mode</TableCell>
            <TableCell align="left">URL</TableCell>
            {isUpSM && <TableCell align="right"></TableCell>}
            <TableCell align="center">Port #</TableCell>
            <TableCell align="center">
              {
                <IconButton
                  disabled={readonly}
                  onClick={() => {
                    upsertDialogDispatch(true, {
                      workspace: workspace,
                      index: -1,
                    });
                  }}
                >
                  <AddTwoTone />
                </IconButton>
              }
            </TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {workspace.spec?.network
            .filter((v) => (readonly ? v.allowedUsers.includes(user.name) : v))
            .map((networkRule, index) => {
              return (
                <TableRow key={index}>
                  <TableCell align="center">
                    {networkRule.public ? (
                      <Tooltip title="No authentication is required for this URL">
                        <PublicOutlined />
                      </Tooltip>
                    ) : networkRule.allowedUsers.length == 0 ? (
                      <Tooltip title="Private URL">
                        <LockOutlined />
                      </Tooltip>
                    ) : (
                      <Tooltip title="Limited Access">
                        <VerifiedUserOutlined />
                      </Tooltip>
                    )}
                  </TableCell>
                  <TableCell align="left">
                    {
                      <>
                        <Link href={networkRule.url || ""} target="_blank">
                          {networkRule.url}
                          <OpenInNewTwoTone
                            fontSize="inherit"
                            sx={{ position: "relative", top: "0.2em" }}
                          />
                        </Link>
                        <IconButton
                          size="small"
                          sx={{ ml: 1 }}
                          onClick={() => {
                            onCopy(networkRule.url);
                          }}
                        >
                          <ContentCopy fontSize="inherit" />
                        </IconButton>
                      </>
                    }
                  </TableCell>
                  {isUpSM && (
                    <TableCell align="right">
                      {!readonly && (
                        <Box sx={{ maxWidth: 150, wordWrap: "break-word" }}>
                          {networkRule.allowedUsers.map((allowedUser) => (
                            <Chip
                              sx={{ ml: 0.5 }}
                              key={allowedUser}
                              color="secondary"
                              size="small"
                              label={allowedUser}
                              icon={<Person />}
                            />
                          ))}
                        </Box>
                      )}
                    </TableCell>
                  )}
                  <TableCell align="center">{networkRule.portNumber}</TableCell>
                  <TableCell align="center">
                    {
                      <>
                        <IconButton
                          disabled={readonly}
                          onClick={() => {
                            upsertDialogDispatch(true, {
                              workspace: workspace,
                              networkRule: networkRule,
                              defaultOpenHttpOptions:
                                (networkRule.customHostPrefix !== "" &&
                                  networkRule.customHostPrefix !== "main") ||
                                networkRule.httpPath !== "/",
                              index: index,
                              isMain:
                                networkRule.url == workspace.status?.mainUrl,
                            });
                          }}
                        >
                          <EditTwoTone />
                        </IconButton>
                        <IconButton
                          disabled={
                            readonly ||
                            networkRule.url == workspace.status?.mainUrl
                          }
                          onClick={() => {
                            deleteDialogDispatch(true, {
                              workspace: workspace,
                              networkRule: networkRule,
                              index: index,
                            });
                          }}
                        >
                          <DeleteTwoTone />
                        </IconButton>
                      </>
                    }
                  </TableCell>
                </TableRow>
              );
            })}
        </TableBody>
      </Table>
    </TableContainer>
  );
};

const WorkspaceItem: React.VFC<{
  workspace: WorkspaceWrapper;
  events: Event[];
  user: User;
  defaultExpandState?: {
    networkRule?: boolean;
    event?: boolean;
  };
  expandState?: {
    networkRule?: boolean;
    event?: boolean;
  };
}> = ({ workspace: ws, events, user, defaultExpandState, expandState }) => {
  console.log("WorkspaceItem", ws.status?.phase, ws.spec?.replicas);
  const [networkRuleExpanded, setNetworkRuleExpanded] = useState(
    defaultExpandState?.networkRule || false
  );
  const [eventExpanded, setEventExpanded] = useState(
    defaultExpandState?.event || false
  );
  if (
    expandState?.networkRule !== undefined &&
    networkRuleExpanded !== expandState.networkRule
  ) {
    setNetworkRuleExpanded(expandState.networkRule);
  }
  if (expandState?.event !== undefined && eventExpanded !== expandState.event) {
    setEventExpanded(expandState.event);
  }

  const { clock } = useLogin();

  const theme = useTheme();
  const isUpSM = useMediaQuery(theme.breakpoints.up("sm"), { noSsr: true });

  const sharedWorkspace = ws.isSharedFor(user);
  const readonly = ws.readonlyFor(user);

  const wsInfoDialogDispatch = WorkspaceInfoDialogContext.useDispatch();

  return (
    <Grid item key={ws.name} xs={12}>
      <Card>
        <CardHeader
          sx={{
            borderBottom: "1px solid",
            borderColor:
              theme.palette.mode === "light"
                ? theme.palette.grey[300]
                : theme.palette.grey["A700"],
          }}
          avatar={
            <Avatar sx={{ backgroundColor: theme.palette.primary.main }}>
              {sharedWorkspace ? <CoPresentTwoTone /> : <WebTwoTone />}
            </Avatar>
          }
          onClick={() => {
            wsInfoDialogDispatch(true, { ws: ws });
          }}
          title={
            <Box display="flex" alignItems="center">
              {ws.status && ws.status.mainUrl && !readonly ? (
                <Link
                  variant="h6"
                  target="_blank"
                  href={ws.status.mainUrl}
                  onClick={(e: any) => e.stopPropagation()}
                  mr={2}
                >
                  {ws.name}{" "}
                  <OpenInNewTwoTone
                    fontSize="inherit"
                    sx={{ position: "relative", top: "0.2em" }}
                  />
                </Link>
              ) : (
                <Typography variant="h6" mr={2}>
                  {ws.name}
                </Typography>
              )}
              {sharedWorkspace && (
                <Stack direction="row" spacing={0.5}>
                  <Chip
                    color="secondary"
                    size="small"
                    label={`shared by ${ws.ownerName}`}
                  />
                  {readonly && (
                    <Chip color="default" size="small" label="readonly" />
                  )}
                </Stack>
              )}
            </Box>
          }
          subheader={ws.spec && ws.spec.template}
          action={
            <Stack direction="row" spacing={2} alignItems="center">
              <Badge
                variant="dot"
                color="error"
                invisible={ws.warningEventsCount(clock) === 0}
                badgeContent={" "}
              >
                <StatusChip ws={ws} />
              </Badge>
              <Box onClick={(e) => e.stopPropagation()}>
                <WorkspaceMenu workspace={ws} user={user} />
              </Box>
            </Stack>
          }
        />
        <CardContent>
          <Grid
            container
            rowSpacing={1}
            columnSpacing={{ xs: 1, sm: 2, md: 2 }}
          >
            <Grid item xs={12}>
              <Stack direction="row">
                <Box display="flex" alignItems="center">
                  <IconButton
                    onClick={() => setNetworkRuleExpanded(!networkRuleExpanded)}
                  >
                    {networkRuleExpanded ? (
                      <KeyboardArrowUpTwoTone />
                    ) : (
                      <KeyboardArrowDownTwoTone />
                    )}
                  </IconButton>
                  <Typography variant="body2">Network Rules</Typography>
                </Box>
                {!sharedWorkspace && (
                  <Box display="flex" alignItems="center">
                    <IconButton
                      onClick={() => setEventExpanded(!eventExpanded)}
                    >
                      {eventExpanded ? (
                        <KeyboardArrowUpTwoTone />
                      ) : (
                        <KeyboardArrowDownTwoTone />
                      )}
                    </IconButton>
                    <Typography variant="body2">Events</Typography>
                  </Box>
                )}
              </Stack>
            </Grid>
            {networkRuleExpanded && (
              <Grid item xs={12} mb={2}>
                <NetworkRuleList workspace={ws} user={user} />
              </Grid>
            )}
            {eventExpanded && (
              <EventsDataGrid
                events={events}
                maxHeight={300}
                clock={clock}
                sx={{ ml: 2 }}
                dataGridProps={{
                  disableColumnMenu: true,
                  columnVisibilityModel: {
                    type: false,
                    reportingController: false,
                    regardingWorkspace: false,
                    series: false,
                    note: isUpSM,
                  },
                  initialState: {
                    sorting: {
                      sortModel: [{ field: "eventTime", sort: "desc" }],
                    },
                  },
                }}
              />
            )}
          </Grid>
        </CardContent>
      </Card>
    </Grid>
  );
};

const WorkspaceList: React.VFC = () => {
  console.log("WorkspaceList");
  const {
    workspaces,
    getWorkspaces,
    user,
    checkIsPolling,
    stopAllPolling,
    search,
    setSearch,
  } = useWorkspaceModule();
  const { loginUser } = useLogin();
  const { enqueueSnackbar } = useSnackbar();
  const isPriv = hasPrivilegedRole(loginUser?.roles || []);

  const [isSearchFocused, setIsSearchFocused] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [isHoverRefreshIcon, setIsHoverRefreshIcon] = useState(false);

  const [expandState, setExpandState] = useState<{
    currentState: boolean | undefined;
    beforeState: boolean | undefined;
  }>({ currentState: undefined, beforeState: true });
  const changeExpandState = (state: boolean) => {
    setExpandState((prev) => {
      return { currentState: state, beforeState: prev.currentState };
    });
    setTimeout(
      () =>
        setExpandState((prev) => {
          return { currentState: undefined, beforeState: prev.currentState };
        }),
      500
    );
  };

  const createDialogDisptch = WorkspaceCreateDialogContext.useDispatch();

  const theme = useTheme();
  const isUpSM = useMediaQuery(theme.breakpoints.up("sm"), { noSsr: true });

  const isPolling = checkIsPolling();

  return (
    <>
      <Paper sx={{ minWidth: 320, maxWidth: 1200, mb: 1, px: 2, py: 1 }}>
        <Stack direction="row" alignItems="center" spacing={2}>
          <TextField
            InputProps={
              search !== ""
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
                            setSearch("");
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
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            onFocus={() => setIsSearchFocused(true)}
            onBlur={() => setIsSearchFocused(false)}
            sx={{ flexGrow: 0.5 }}
          />
          <Box sx={{ flexGrow: 1 }} />
          {isPriv && (isUpSM || (!isSearchFocused && search === "")) && (
            <UserSelect />
          )}
          {isUpSM &&
            (expandState.beforeState === true ? (
              <Tooltip title="Collapse all" placement="top">
                <span>
                  <IconButton
                    onClick={() => changeExpandState(false)}
                    disabled={expandState.beforeState === undefined}
                  >
                    <UnfoldLessOutlined />
                  </IconButton>
                </span>
              </Tooltip>
            ) : (
              <Tooltip title="Expand all" placement="top">
                <span>
                  <IconButton
                    onClick={() => changeExpandState(true)}
                    disabled={expandState.beforeState === undefined}
                  >
                    <UnfoldMoreOutlined />
                  </IconButton>
                </span>
              </Tooltip>
            ))}
          <Tooltip
            title={isPolling ? "Cancel polling" : "Refresh"}
            placement="top"
          >
            {isHoverRefreshIcon && isPolling ? (
              <IconButton
                color="inherit"
                onClick={() => {
                  stopAllPolling();
                  enqueueSnackbar("Stop all polling", { variant: "info" });
                }}
                onMouseLeave={() => setIsHoverRefreshIcon(false)}
              >
                <Clear />
              </IconButton>
            ) : (
              <IconButton
                color="inherit"
                onClick={() => {
                  setIsLoading(true);
                  setTimeout(() => {
                    setIsLoading(false);
                  }, 2000);
                  if (!isLoading) getWorkspaces(user.name);
                }}
                onMouseEnter={() =>
                  setIsHoverRefreshIcon((prev) => (prev ? prev : true))
                }
                onMouseLeave={() =>
                  setIsHoverRefreshIcon((prev) => (prev ? false : prev))
                }
              >
                {isPolling || isLoading ? (
                  <RotatingRefreshTwoTone />
                ) : (
                  <RefreshTwoTone />
                )}
              </IconButton>
            )}
          </Tooltip>
          <Tooltip title="Create Workspace" placement="top">
            <Fab
              size="small"
              color="primary"
              onClick={() => {
                createDialogDisptch(true);
              }}
              sx={{ flexShrink: 0 }}
            >
              <AddTwoTone />
            </Fab>
          </Tooltip>
        </Stack>
      </Paper>
      {!Object.keys(workspaces).filter(
        (wsName) => search === "" || Boolean(wsName.match(search))
      ).length && (
        <Paper sx={{ minWidth: 320, maxWidth: 1200, mb: 1, p: 4 }}>
          <Typography
            variant="subtitle1"
            sx={{ color: "text.secondary", textAlign: "center" }}
          >
            No Workspaces found.
          </Typography>
        </Paper>
      )}
      <Grid container spacing={1}>
        {Object.keys(workspaces)
          .filter((wsName) => search === "" || Boolean(wsName.match(search)))
          .map((wsName) => workspaces[wsName])
          .sort((a, b) =>
            a.ownerName !== b.ownerName ? 1 : a.name < b.name ? -1 : 1
          )
          .map((ws) => (
            <WorkspaceItem
              workspace={ws}
              key={ws.name}
              events={ws.events}
              user={user}
              defaultExpandState={
                isUpSM
                  ? { networkRule: true }
                  : { networkRule: false, event: false }
              }
              expandState={
                expandState.currentState === undefined
                  ? undefined
                  : expandState.currentState
                  ? { networkRule: true, event: false }
                  : { networkRule: false, event: false }
              }
            />
          ))}
      </Grid>
    </>
  );
};

export const WorkspacePage: React.VFC = () => {
  console.log("WorkspacePage");

  return (
    <PageTemplate title="Workspaces">
      <div>
        <WorkspaceContext.Provider>
          <WorkspaceCreateDialogContext.Provider>
            <WorkspaceStartDialogContext.Provider>
              <WorkspaceStopDialogContext.Provider>
                <WorkspaceDeleteDialogContext.Provider>
                  <WorkspaceInfoDialogContext.Provider>
                    <NetworkRuleUpsertDialogContext.Provider>
                      <NetworkRuleDeleteDialogContext.Provider>
                        <EventDetailDialogContext.Provider>
                          <WorkspaceList />
                        </EventDetailDialogContext.Provider>
                      </NetworkRuleDeleteDialogContext.Provider>
                    </NetworkRuleUpsertDialogContext.Provider>
                  </WorkspaceInfoDialogContext.Provider>
                </WorkspaceDeleteDialogContext.Provider>
              </WorkspaceStopDialogContext.Provider>
            </WorkspaceStartDialogContext.Provider>
          </WorkspaceCreateDialogContext.Provider>
        </WorkspaceContext.Provider>
      </div>
    </PageTemplate>
  );
};
