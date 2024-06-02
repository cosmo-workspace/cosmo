import {
  AccountCircle,
  Badge as BadgeIcon,
  ExitToApp,
  FingerprintTwoTone,
  Info,
  LockOutlined,
  Menu as MenuIcon,
  Notifications,
  ReportProblem,
  SupervisorAccountTwoTone,
  VpnKey,
  Warning,
  WebTwoTone,
} from "@mui/icons-material";
import {
  Alert,
  Badge,
  Box,
  Button,
  Chip,
  colors,
  Container,
  CssBaseline,
  Divider,
  Grid,
  IconButton,
  Link,
  List,
  ListItem,
  ListItemIcon,
  ListItemText,
  Menu,
  MenuItem,
  Stack,
  Toolbar,
  Typography,
} from "@mui/material";
import MuiAppBar, { AppBarProps } from "@mui/material/AppBar";
import { experimentalStyled as styled } from "@mui/material/styles";
import React, { useEffect } from "react";
import { ErrorBoundary } from "react-error-boundary";
import { Link as RouterLink } from "react-router-dom";
import { useLogin } from "../../components/LoginProvider";
import logo from "../../logo-with-name-small.png";
import { formatTime } from "../atoms/EventsDataGrid";
import { NameAvatar } from "../atoms/NameAvatar";
import { AuthenticatorManageDialogContext } from "../organisms/AuthenticatorManageDialog";
import { EventDetailDialogContext } from "../organisms/EventDetailDialog";
import { latestTime } from "../organisms/EventModule";
import { PasswordChangeDialogContext } from "../organisms/PasswordChangeDialog";
import { UserInfoDialogContext } from "../organisms/UserActionDialog";
import { UserAddonChangeDialogContext } from "../organisms/UserAddonsChangeDialog";
import {
  isAdminRole,
  isAdminUser,
  isPrivilegedRole,
} from "../organisms/UserModule";
import { UserNameChangeDialogContext } from "../organisms/UserNameChangeDialog";

const AppBar = styled(MuiAppBar)<AppBarProps>(({ theme }) => ({
  zIndex: theme.zIndex.drawer + 1,
  transition: theme.transitions.create(["width", "margin"], {
    easing: theme.transitions.easing.sharp,
    duration: theme.transitions.duration.leavingScreen,
  }),
}));

const Copyright = () => {
  return (
    <Typography variant="body2" color="textSecondary" align="center">
      {"Copyright © "}
      <Link href="https://github.com/cosmo-workspace">cosmo-workspace</Link>
      {` ${new Date().getFullYear()}.`}
    </Typography>
  );
};

interface PageTemplateProps {
  children: React.ReactNode;
  title: string;
}

export const PageTemplate: React.FC<
  React.PropsWithChildren<PageTemplateProps>
> = ({ children, title }) => {
  const {
    loginUser,
    logout,
    myEvents,
    getMyEvents,
    watchMyEvents,
    newEventsCount,
    setNewEventsCount,
    clock,
    updateClock,
  } = useLogin();
  const authenticatorManagerDialogDispatch = AuthenticatorManageDialogContext
    .useDispatch();
  const passwordChangeDialogDispach = PasswordChangeDialogContext.useDispatch();
  const userNameChangeDialogDispach = UserNameChangeDialogContext.useDispatch();
  const userAddonChangeDialogDispatch = UserAddonChangeDialogContext
    .useDispatch();
  const userInfoDialogDispatch = UserInfoDialogContext.useDispatch();
  const isAdmin = isAdminUser(loginUser);
  const isSignIn = Boolean(loginUser);
  const canChangePassword = Boolean(loginUser?.authType === "password-secret");

  const eventDetailDialogDispatch = EventDetailDialogContext.useDispatch();

  useEffect(() => {
    watchMyEvents();
  }, [isSignIn]);

  const manageAuthenticators = () => {
    console.log("manageAuthenticators");
    authenticatorManagerDialogDispatch(true, { user: loginUser! });
    setAnchorEl(null);
  };

  const changeUserName = () => {
    console.log("changeUserName");
    userNameChangeDialogDispach(true, { user: loginUser! });
    setAnchorEl(null);
  };

  const changePassword = () => {
    console.log("changePassword");
    passwordChangeDialogDispach(true);
    setAnchorEl(null);
  };

  const changeAddons = () => {
    console.log("changeAddons");
    userAddonChangeDialogDispatch(true, { user: loginUser! });
    setAnchorEl(null);
  };

  const openUserInfoDialog = () => {
    console.log("openUserInfoDialog");
    userInfoDialogDispatch(true, {
      user: loginUser!,
      defaultOpenUserAddon: true,
    });
    setAnchorEl(null);
  };

  const [menuAnchorEl, setMenuAnchorEl] = React.useState<null | HTMLElement>(
    null,
  );
  const [anchorEl, setAnchorEl] = React.useState<null | HTMLElement>(null);
  const [eventAnchorEl, setEventAnchorEl] = React.useState<null | HTMLElement>(
    null,
  );

  return (
    <Box sx={{ display: "flex" }}>
      <CssBaseline />
      <AppBar position="absolute">
        <Toolbar sx={{ pr: 3 }}>
          <IconButton
            edge="start"
            color="inherit"
            onClick={(e) => setMenuAnchorEl(e.currentTarget)}
            sx={{ mr: 2 }}
          >
            <MenuIcon />
          </IconButton>
          <Menu
            id="basic-menu"
            anchorEl={menuAnchorEl}
            open={Boolean(menuAnchorEl)}
            onClose={() => setMenuAnchorEl(null)}
          >
            {loginUser && (
              <RouterLink
                to="/workspace"
                style={{ textDecoration: "none", color: "inherit" }}
              >
                <MenuItem>
                  <ListItemIcon>
                    <WebTwoTone />
                  </ListItemIcon>
                  <ListItemText primary="Workspaces" />
                </MenuItem>
              </RouterLink>
            )}
            {loginUser && isAdmin && (
              <RouterLink
                to="/user"
                style={{ textDecoration: "none", color: "inherit" }}
              >
                <MenuItem>
                  <ListItemIcon>
                    <SupervisorAccountTwoTone />
                  </ListItemIcon>
                  <ListItemText primary="Users" />
                </MenuItem>
              </RouterLink>
            )}
            {loginUser && (
              <RouterLink
                to="/event"
                style={{ textDecoration: "none", color: "inherit" }}
              >
                <MenuItem>
                  <ListItemIcon>
                    <Notifications />
                  </ListItemIcon>
                  <ListItemText primary="Events" />
                </MenuItem>
              </RouterLink>
            )}
            {!loginUser && (
              <RouterLink
                to="/signin"
                style={{ textDecoration: "none", color: "inherit" }}
              >
                <MenuItem>
                  <ListItemIcon>
                    <LockOutlined />
                  </ListItemIcon>
                  <ListItemText primary="sign in" />
                </MenuItem>
              </RouterLink>
            )}
          </Menu>
          <RouterLink
            to="/workspace"
            style={{ textDecoration: "none", color: "inherit" }}
          >
            <img alt="cosmo" src={logo} height={40} />
          </RouterLink>
          <Box sx={{ flexGrow: 1 }} />
          <Box>
            <IconButton
              color="inherit"
              onClick={(e) => {
                updateClock();
                setEventAnchorEl(e.currentTarget);
              }}
              disabled={!isSignIn}
            >
              <Badge
                invisible={newEventsCount === 0}
                variant="dot"
                color="error"
              >
                <Notifications />
              </Badge>
            </IconButton>
            <Menu
              id="notification-menu"
              anchorEl={eventAnchorEl}
              open={Boolean(eventAnchorEl)}
              onClose={() => {
                setEventAnchorEl(null);
                setNewEventsCount(0);
              }}
              PaperProps={{
                style: {
                  maxHeight: 700,
                  width: 400,
                },
              }}
            >
              <List>
                {myEvents.length > 0
                  ? (
                    myEvents.slice(0, 10).map((event, index) => (
                      <React.Fragment key={event.id}>
                        <MenuItem
                          onClick={() =>
                            eventDetailDialogDispatch(true, { event: event })}
                        >
                          <ListItemIcon>
                            {event.type == "Normal"
                              ? <Info color="success" />
                              : <Warning color="warning" />}
                          </ListItemIcon>
                          <ListItemText
                            primary={
                              <Stack>
                                <Stack direction="row" alignContent="center">
                                  <Box display="flex" alignItems="center">
                                    <Chip
                                      variant="outlined"
                                      color="primary"
                                      size="small"
                                      label={event.regarding?.kind}
                                      sx={{ mr: 1 }}
                                    />
                                    <Typography variant="body1">
                                      {event.reason}
                                    </Typography>
                                  </Box>
                                  <Box sx={{ flex: "1 1 auto" }} />
                                  <Typography variant="body2">
                                    {formatTime(
                                      clock.getTime() - latestTime(event),
                                    )}
                                  </Typography>
                                </Stack>
                                <Stack direction="row" alignContent="center">
                                  <Typography variant="body2">
                                    {event.regarding?.name}
                                  </Typography>
                                </Stack>
                              </Stack>
                            }
                            secondary={
                              <Typography
                                variant="body2"
                                color="textSecondary"
                                style={{
                                  wordWrap: "break-word",
                                  whiteSpace: "normal",
                                }}
                              >
                                {event.note}
                              </Typography>
                            }
                          />
                        </MenuItem>
                        {index < myEvents.length - 1
                          ? <Divider component="li" />
                          : null}
                      </React.Fragment>
                    ))
                  )
                  : (
                    <ListItem>
                      <ListItemText>
                        <Typography
                          align="center"
                          variant="body2"
                          color="textSecondary"
                        >
                          No new events recieved
                        </Typography>
                      </ListItemText>
                    </ListItem>
                  )}
              </List>
              <Divider />
              <Stack direction="row" sx={{ mt: 1, mr: 2 }}>
                <Box sx={{ flexGrow: 1 }} />
                <Link
                  variant="body2"
                  href="#/event"
                  onClick={() => setEventAnchorEl(null)}
                >
                  View all events...
                </Link>
              </Stack>
            </Menu>
            <IconButton
              color="inherit"
              onClick={(e) => setAnchorEl(e.currentTarget)}
              disabled={!isSignIn}
            >
              <AccountCircle fontSize="large" />
            </IconButton>
            <Menu
              id="basic-menu"
              anchorEl={anchorEl}
              open={Boolean(anchorEl)}
              onClose={() => setAnchorEl(null)}
            >
              <Stack alignItems="center" spacing={1} sx={{ mt: 1, mb: 2 }}>
                <NameAvatar
                  name={loginUser?.displayName}
                  sx={{ width: 50, height: 50 }}
                  onClick={() => openUserInfoDialog()}
                />
                <Typography>{loginUser?.displayName}</Typography>
                <Typography color={colors.grey[700]} fontSize="small">
                  {loginUser?.name}
                </Typography>
                <Grid container justifyContent="center" sx={{ width: 200 }}>
                  {loginUser?.roles &&
                    loginUser.roles.map((v, i) => {
                      return (
                        <Grid item key={i}>
                          <Chip
                            color={isPrivilegedRole(v)
                              ? "error"
                              : isAdminRole(v)
                              ? "warning"
                              : "default"}
                            variant="outlined"
                            size="small"
                            key={i}
                            label={v}
                          />
                        </Grid>
                      );
                    })}
                </Grid>
              </Stack>
              <Divider sx={{ mb: 1 }} />
              {isSignIn && (
                <MenuItem onClick={() => manageAuthenticators()}>
                  <ListItemIcon>
                    <FingerprintTwoTone fontSize="small" />
                  </ListItemIcon>
                  <ListItemText>Manage WebAuthn Credentials...</ListItemText>
                </MenuItem>
              )}
              {isSignIn && canChangePassword && (
                <MenuItem onClick={() => changePassword()}>
                  <ListItemIcon>
                    <VpnKey fontSize="small" />
                  </ListItemIcon>
                  <ListItemText>Change Password...</ListItemText>
                </MenuItem>
              )}
              {isSignIn && (
                <MenuItem onClick={() => changeUserName()}>
                  <ListItemIcon>
                    <BadgeIcon fontSize="small" />
                  </ListItemIcon>
                  <ListItemText>Change Name...</ListItemText>
                </MenuItem>
              )}
              {isSignIn && isAdmin && (
                <MenuItem onClick={() => changeAddons()}>
                  <ListItemIcon>
                    <BadgeIcon fontSize="small" />
                  </ListItemIcon>
                  <ListItemText>Change Addons...</ListItemText>
                </MenuItem>
              )}
              <Divider />
              {isSignIn && (
                <MenuItem onClick={() => logout()}>
                  <ListItemIcon>
                    <ExitToApp fontSize="small" />
                  </ListItemIcon>
                  <ListItemText>Sign out</ListItemText>
                </MenuItem>
              )}
            </Menu>
          </Box>
        </Toolbar>
      </AppBar>

      <Box
        component="main"
        sx={{
          backgroundColor: (theme) =>
            theme.palette.mode === "light"
              ? theme.palette.grey[100]
              : theme.palette.grey[900],
          flexGrow: 1,
          height: "100vh",
          overflow: "auto",
        }}
      >
        <Toolbar />
        <Container maxWidth="lg" sx={{ mt: 2, mb: 2 }}>
          <Typography
            component="h2"
            variant="h5"
            color="inherit"
            noWrap
            sx={{ mb: 1 }}
          >
            {title}
          </Typography>
          <ErrorBoundary
            FallbackComponent={({ error, resetErrorBoundary }) => {
              return (
                <Stack sx={{ marginTop: 3, alignItems: "center" }}>
                  <ReportProblem fontSize="large" />
                  <Typography variant="h5">Something went wrong...</Typography>
                  <p>{error.toString()}</p>
                  <Alert severity="info" variant="outlined" sx={{ margin: 2 }}>
                    Please contact administrators to report the issue.
                  </Alert>
                  <Stack direction="row" spacing={1}>
                    <Button
                      size="small"
                      onClick={() => {
                        window.location.href = "/";
                        resetErrorBoundary();
                      }}
                      variant="contained"
                      color="primary"
                    >
                      Go to Top
                    </Button>
                  </Stack>
                </Stack>
              );
            }}
          >
            {children}
          </ErrorBoundary>
          <Box pt={4}>
            <Copyright />
          </Box>
        </Container>
      </Box>
    </Box>
  );
};
