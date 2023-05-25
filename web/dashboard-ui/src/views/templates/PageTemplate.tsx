import { AccountCircle, Badge, ExitToApp, LockOutlined, Menu as MenuIcon, ReportProblem, SupervisorAccountTwoTone, VpnKey, WebTwoTone } from "@mui/icons-material";
import {
  Alert,
  Box, Button, Chip,
  Container, CssBaseline, Divider, Grid, IconButton,
  Link, ListItemIcon, ListItemText,
  Menu, MenuItem, Stack, Toolbar, Typography,
  colors
} from "@mui/material";
import MuiAppBar, { AppBarProps } from '@mui/material/AppBar';
import { experimentalStyled as styled } from "@mui/material/styles";
import React from "react";
import { ErrorBoundary } from 'react-error-boundary';
import { Link as RouterLink } from "react-router-dom";
import { useLogin } from "../../components/LoginProvider";
import logo from "../../logo-with-name-small.png";
import { NameAvatar } from "../atoms/NameAvatar";
import { PasswordChangeDialogContext } from "../organisms/PasswordChangeDialog";
import { isAdminRole, isAdminUser, isPrivilegedRole } from "../organisms/UserModule";
import { UserNameChangeDialogContext } from "../organisms/UserNameChangeDialog";


const AppBar = styled(MuiAppBar)<AppBarProps>(({ theme }) => ({
  zIndex: theme.zIndex.drawer + 1,
  transition: theme.transitions.create(['width', 'margin'], {
    easing: theme.transitions.easing.sharp,
    duration: theme.transitions.duration.leavingScreen,
  }),
}));

const Copyright = () => {
  return (
    <Typography variant="body2" color="textSecondary" align="center">
      {"Copyright © "}
      <Link href="https://github.com/cosmo-workspace">
        cosmo-workspace
      </Link>{` ${new Date().getFullYear()}.`}
    </Typography>
  );
};

interface PageTemplateProps {
  children: React.ReactNode;
  title: string;
}

export const PageTemplate: React.FC<React.PropsWithChildren<PageTemplateProps>> = ({ children, title, }) => {

  const { loginUser, logout } = useLogin();
  const passwordChangeDialogDispach = PasswordChangeDialogContext.useDispatch();
  const userNameChangeDialogDispach = UserNameChangeDialogContext.useDispatch();
  const isAdmin = isAdminUser(loginUser);
  const isSignIn = Boolean(loginUser);

  const changeUserName = () => {
    console.log('changeUserName');
    userNameChangeDialogDispach(true, { user: loginUser! });
    setAnchorEl(null);
  }

  const changePassword = () => {
    console.log('changePassword');
    passwordChangeDialogDispach(true);
    setAnchorEl(null);
  }

  const [menuAnchorEl, setMenuAnchorEl] = React.useState<null | HTMLElement>(null);
  const [anchorEl, setAnchorEl] = React.useState<null | HTMLElement>(null);


  return (
    <Box sx={{ display: 'flex' }}>
      <CssBaseline />
      <AppBar position="absolute">
        <Toolbar sx={{ pr: 3 }} >
          <IconButton
            edge="start"
            color="inherit"
            onClick={(e) => setMenuAnchorEl(e.currentTarget)}
            sx={{ mr: 2 }}>
            <MenuIcon />
          </IconButton>
          <Menu
            id="basic-menu"
            anchorEl={menuAnchorEl}
            open={Boolean(menuAnchorEl)}
            onClose={() => setMenuAnchorEl(null)}
          >
            {loginUser && <RouterLink to="/workspace" style={{ textDecoration: "none", color: "inherit" }}>
              <MenuItem>
                <ListItemIcon>
                  <WebTwoTone />
                </ListItemIcon>
                <ListItemText primary="Workspaces" />
              </MenuItem>
            </RouterLink>}
            {loginUser && isAdmin && <RouterLink to="/user" style={{ textDecoration: "none", color: "inherit" }}>
              <MenuItem>
                <ListItemIcon>
                  <SupervisorAccountTwoTone />
                </ListItemIcon>
                <ListItemText primary="Users" />
              </MenuItem>
            </RouterLink>}
            {!loginUser && <RouterLink to="/signin" style={{ textDecoration: "none", color: "inherit" }}>
              <MenuItem>
                <ListItemIcon>
                  <LockOutlined />
                </ListItemIcon>
                <ListItemText primary="sign in" />
              </MenuItem>
            </RouterLink>}
          </Menu>
          <RouterLink to="/workspace" style={{ textDecoration: "none", color: "inherit" }}>
            <img alt="cosmo" src={logo} height={40} />
          </RouterLink>
          <Box sx={{ flexGrow: 1 }} />
          <Box>
            <IconButton
              color="inherit"
              onClick={(e) => setAnchorEl(e.currentTarget)}
              disabled={!isSignIn}>
              <AccountCircle fontSize="large" />
            </IconButton>
            <Menu
              id="basic-menu"
              anchorEl={anchorEl}
              open={Boolean(anchorEl)}
              onClose={() => setAnchorEl(null)}
            >
              <Stack alignItems="center" spacing={1} sx={{ mt: 1, mb: 2 }}>
                <NameAvatar name={loginUser?.displayName} sx={{ width: 50, height: 50 }} />
                <Typography>{loginUser?.displayName}</Typography>
                <Typography color={colors.grey[700]} fontSize="small">{loginUser?.name}</Typography>
                <Grid container justifyContent="center" sx={{ width: 200 }}>
                  {loginUser?.roles && loginUser.roles.map((v, i) => {
                    return (
                      <Grid item key={i} >
                        <Chip color={isPrivilegedRole(v) ? "error" : isAdminRole(v) ? "warning" : "default"} variant="outlined" size="small" key={i} label={v} />
                      </Grid>)
                  })}
                </Grid>
              </Stack>
              <Divider sx={{ mb: 1 }} />
              {isSignIn && <MenuItem onClick={() => changeUserName()}>
                <ListItemIcon><Badge fontSize="small" /></ListItemIcon>
                <ListItemText>Change user name...</ListItemText>
              </MenuItem>}
              {isSignIn && <MenuItem onClick={() => changePassword()}>
                <ListItemIcon><VpnKey fontSize="small" /></ListItemIcon>
                <ListItemText>Change password...</ListItemText>
              </MenuItem>}
              <Divider />
              {isSignIn && <MenuItem onClick={() => logout()}>
                <ListItemIcon><ExitToApp fontSize="small" /></ListItemIcon>
                <ListItemText>Sign out</ListItemText>
              </MenuItem>}
            </Menu>
          </Box>

        </Toolbar>
      </AppBar>

      <Box
        component="main"
        sx={{
          backgroundColor: (theme) =>
            theme.palette.mode === 'light'
              ? theme.palette.grey[100]
              : theme.palette.grey[900],
          flexGrow: 1,
          height: '100vh',
          overflow: 'auto',
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
            FallbackComponent={
              ({ error, resetErrorBoundary }) => {
                return (
                  <Stack sx={{ marginTop: 3, alignItems: 'center' }}>
                    <ReportProblem fontSize="large" />
                    <Typography variant="h5">Something went wrong...</Typography>
                    <p>{error.toString()}</p>
                    <Alert severity="info" variant="outlined" sx={{ margin: 2 }}>Please contact administrators to report the issue.</Alert>
                    <Stack direction="row" spacing={1}>
                      <Button size="small" onClick={() => { window.location.href = '/'; resetErrorBoundary() }} variant="contained" color="primary">Go to Top</Button>
                    </Stack>
                  </Stack>
                )
              }
            }
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
