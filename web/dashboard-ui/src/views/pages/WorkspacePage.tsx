import useUrlState from "@ahooksjs/use-url-state";
import {
  AddTwoTone,
  CheckCircleOutlined,
  Clear,
  ContentCopy,
  DeleteTwoTone,
  EditTwoTone,
  Error,
  ErrorOutline,
  ExpandLessTwoTone,
  ExpandMoreTwoTone,
  InfoOutlined,
  KeyboardArrowDownTwoTone,
  KeyboardArrowUpTwoTone,
  LockOutlined,
  MoreVertTwoTone,
  OpenInNewTwoTone,
  PlayCircleFilledWhiteTwoTone,
  PublicOutlined,
  RefreshTwoTone,
  SearchTwoTone,
  StopCircleOutlined,
  StopCircleTwoTone,
  WebTwoTone,
} from "@mui/icons-material";
import {
  Avatar,
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
  styled,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  TextField,
  Tooltip,
  Typography,
  useMediaQuery,
  useTheme,
} from "@mui/material";
import copy from "copy-to-clipboard";
import { useSnackbar } from "notistack";
import React, { useRef, useState } from "react";
import { useLogin } from "../../components/LoginProvider";
import { Event } from "../../proto/gen/dashboard/v1alpha1/event_pb";
import { Workspace } from "../../proto/gen/dashboard/v1alpha1/workspace_pb";
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
import {
  computeStatus,
  useWorkspaceModule,
  WorkspaceContext,
  WorkspaceWrapper,
} from "../organisms/WorkspaceModule";
import { PageTemplate } from "../templates/PageTemplate";

/**
 * view
 */
const RotatingRefreshTwoTone = styled(RefreshTwoTone)({
  animation: "rotatingRefresh 2s linear infinite",
  "@keyframes rotatingRefresh": {
    "to": {
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
          icon={ws.progress > 0
            ? ws.progress > 100 ? <InfoOutlined /> : (
              <CircularProgress
                color="info"
                size={13}
                variant="determinate"
                value={ws.progress}
              />
            )
            : <CircularProgress color="info" size={13} />}
          color="info"
          label={statusLabel}
        />
      );
  }
};

const WorkspaceMenu: React.VFC<{ workspace: Workspace }> = ({ workspace }) => {
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
          disabled={!Boolean(workspace.name)}
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
          avatar={<NameAvatar name={user.displayName} />}
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
              setUser(user);
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

const NetworkRuleList: React.FC<{ workspace: Workspace }> = (
  { workspace },
) => {
  const upsertDialogDispatch = NetworkRuleUpsertDialogContext.useDispatch();
  const deleteDialogDispatch = NetworkRuleDeleteDialogContext.useDispatch();
  const { enqueueSnackbar } = useSnackbar();
  const onCopy = (text: string) => {
    copy(text);
    enqueueSnackbar("Copied!", { variant: "success" });
  };
  const theme = useTheme();

  return (
    <TableContainer
      sx={{
        border: "1px solid",
        borderRadius: "4px",
        borderColor: theme.palette.mode === "light"
          ? "rgba(224,224,224,1)"
          : "rgba(81,81,81,1)",
      }}
    >
      <Table size="small">
        <TableHead sx={{ backgroundColor: theme.palette.background.default }}>
          <TableRow>
            <TableCell align="center">Mode</TableCell>
            <TableCell align="left">URL</TableCell>
            <TableCell align="center">Port #</TableCell>
            <TableCell align="center">
              {
                <IconButton
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
          {workspace.spec?.network.map((networkRule, index) => {
            return (
              <TableRow key={index}>
                <TableCell align="center">
                  {networkRule.public &&
                      (
                        <Tooltip title="No authentication is required for this URL">
                          <PublicOutlined />
                        </Tooltip>
                      ) ||
                    (
                      <Tooltip title="Private URL">
                        <LockOutlined />
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
                <TableCell align="center">{networkRule.portNumber}</TableCell>
                <TableCell align="center">
                  {
                    <>
                      <IconButton
                        onClick={() => {
                          upsertDialogDispatch(true, {
                            workspace: workspace,
                            networkRule: networkRule,
                            defaultOpenHttpOptions: true,
                            index: index,
                            isMain:
                              networkRule.url == workspace.status?.mainUrl,
                          });
                        }}
                      >
                        <EditTwoTone />
                      </IconButton>
                      <IconButton
                        disabled={networkRule.url == workspace.status?.mainUrl}
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

const WorkspaceItem: React.VFC<
  { workspace: WorkspaceWrapper; events: Event[] }
> = (
  { workspace: ws, events },
) => {
  console.log("WorkspaceItem", ws.status?.phase, ws.spec?.replicas);
  const [networkRuleExpanded, setNetworkRuleExpanded] = useState(false);
  const [eventExpanded, setEventExpanded] = useState(false);
  const { clock } = useLogin();

  const theme = useTheme();
  const isUpSM = useMediaQuery(theme.breakpoints.up("sm"), { noSsr: true });

  return (
    <Grid item key={ws.name} xs={12}>
      <Card>
        <CardHeader
          sx={{
            borderBottom: "1px solid",
            borderColor: theme.palette.mode === "light"
              ? theme.palette.grey[300]
              : theme.palette.grey["A700"],
          }}
          avatar={
            <Avatar>
              <WebTwoTone />
            </Avatar>
          }
          title={ws.status && ws.status.mainUrl
            ? (
              <Link
                variant="h6"
                target="_blank"
                href={ws.status.mainUrl}
                onClick={(e: any) => e.stopPropagation()}
              >
                {ws.name}{" "}
                <OpenInNewTwoTone
                  fontSize="inherit"
                  sx={{ position: "relative", top: "0.2em" }}
                />
              </Link>
            )
            : <Typography variant="h6">{ws.name}</Typography>}
          subheader={ws.spec && ws.spec.template}
          action={
            <Stack direction="row" spacing={2} alignItems="center">
              {ws.hasWarningEvents(clock) && (
                <IconButton
                  color="inherit"
                  onClick={() => setEventExpanded(true)}
                >
                  <Error color="error" />
                </IconButton>
              )}
              <StatusChip ws={ws} />
              <Box onClick={(e) => e.stopPropagation()}>
                <WorkspaceMenu workspace={ws} />
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
                    {networkRuleExpanded
                      ? <KeyboardArrowUpTwoTone />
                      : <KeyboardArrowDownTwoTone />}
                  </IconButton>
                  <Typography variant="body2">Network Rules</Typography>
                </Box>
                <Box display="flex" alignItems="center">
                  <IconButton
                    onClick={() => setEventExpanded(!eventExpanded)}
                  >
                    {eventExpanded
                      ? <KeyboardArrowUpTwoTone />
                      : <KeyboardArrowDownTwoTone />}
                  </IconButton>
                  <Typography variant="body2">Events</Typography>
                </Box>
              </Stack>
            </Grid>
            {networkRuleExpanded &&
              (
                <Grid item xs={12} mb={2}>
                  <NetworkRuleList workspace={ws} />
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
  } = useWorkspaceModule();
  const { loginUser } = useLogin();
  const { enqueueSnackbar } = useSnackbar();
  const isPriv = hasPrivilegedRole(loginUser?.roles || []);
  const [urlParam, setUrlParam] = useUrlState({ "search": "" }, {
    stringifyOptions: { skipEmptyString: true },
  });
  const [isSearchFocused, setIsSearchFocused] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [isHoverRefreshIcon, setIsHoverRefreshIcon] = useState(false);
  const createDialogDisptch = WorkspaceCreateDialogContext.useDispatch();

  const theme = useTheme();
  const isUpSM = useMediaQuery(theme.breakpoints.up("sm"), { noSsr: true });

  const isPolling = checkIsPolling();

  return (
    <>
      <Paper sx={{ minWidth: 320, maxWidth: 1200, mb: 1, px: 2, py: 1 }}>
        <Stack direction="row" alignItems="center" spacing={2}>
          <TextField
            InputProps={urlParam.search !== ""
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
              }}
            placeholder="Search"
            size="small"
            value={urlParam.search}
            onChange={(e) => setUrlParam({ search: e.target.value })}
            onFocus={() => setIsSearchFocused(true)}
            onBlur={() => setIsSearchFocused(false)}
            sx={{ flexGrow: 0.5 }}
          />
          <Box sx={{ flexGrow: 1 }} />
          {isPriv && (isUpSM || (!isSearchFocused && urlParam.search === "")) &&
            <UserSelect />}
          <Tooltip
            title={isPolling ? "Cancel polling" : "Refresh"}
            placement="top"
          >
            {(isHoverRefreshIcon && isPolling)
              ? (
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
              )
              : (
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
                    setIsHoverRefreshIcon((prev) => prev ? prev : true)}
                  onMouseLeave={() =>
                    setIsHoverRefreshIcon((prev) => prev ? false : prev)}
                >
                  {isPolling || isLoading
                    ? <RotatingRefreshTwoTone />
                    : <RefreshTwoTone />}
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
      {!Object.keys(workspaces).filter((wsName) =>
        urlParam.search === "" || Boolean(wsName.match(urlParam.search))
      ).length &&
        (
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
        {Object.keys(workspaces).filter((wsName) =>
          urlParam.search === "" || Boolean(wsName.match(urlParam.search))
        ).map((wsName) => workspaces[wsName]).sort((a, b) =>
          (a.name < b.name) ? -1 : 1
        ).map((ws) => (
          <WorkspaceItem
            workspace={ws}
            key={ws.name}
            events={ws.events}
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
                  <NetworkRuleUpsertDialogContext.Provider>
                    <NetworkRuleDeleteDialogContext.Provider>
                      <EventDetailDialogContext.Provider>
                        <WorkspaceList />
                      </EventDetailDialogContext.Provider>
                    </NetworkRuleDeleteDialogContext.Provider>
                  </NetworkRuleUpsertDialogContext.Provider>
                </WorkspaceDeleteDialogContext.Provider>
              </WorkspaceStopDialogContext.Provider>
            </WorkspaceStartDialogContext.Provider>
          </WorkspaceCreateDialogContext.Provider>
        </WorkspaceContext.Provider>
      </div>
    </PageTemplate>
  );
};
