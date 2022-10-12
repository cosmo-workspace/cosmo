import { createConnectTransport, createPromiseClient } from '@bufbuild/connect-web';
import LockOutlinedIcon from '@mui/icons-material/LockOutlined';
import { Alert, AppBar, Avatar, Backdrop, Box, Button, CircularProgress, Container, Link, Snackbar, Stack, TextField, Toolbar, Typography } from '@mui/material';
import React, { useMemo, useState } from 'react';
import { ErrorBoundary } from 'react-error-boundary';
import { AuthProxyService } from '../../auth-proxy/v1alpha1/authproxy_connectweb';
import logo from "../../logo-with-name-small.png";
import { PasswordTextField } from '../atoms/PasswordTextField';

/**
 * view
 */

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

interface Inputs {
  userid: string,
  password: string,
};

const transport = createConnectTransport({
  baseUrl: import.meta.env.BASE_URL,
});

function useClient() {
  return useMemo(() => createPromiseClient(AuthProxyService, transport), [AuthProxyService]);
}

export const SignIn: React.VFC = () => {
  const [values, _setValues] = useState<Inputs>({ userid: '', password: '' });
  const [errors, setErrors] = useState<Inputs>({ userid: '', password: '' });
  const [signinResult, setSigninResult] = useState("");
  const [loading, setLoading] = useState(false);

  const client = useClient();

  const validateInput = (inp: Partial<Inputs>) => {
    const errs = Object.entries(inp).reduce((acc, [k, v]) => ({ ...acc, [k]: v ? '' : 'Required' }), {});
    setErrors({ ...errors, ...errs });
    return !Object.values(errs).some(errorMessage => errorMessage !== '');
  }

  const setValue = (inp: Partial<Inputs>) => {
    _setValues({ ...values, ...inp });
    validateInput(inp);
  }

  const login = async () => {
    try {
      if (!validateInput(values)) {
        return;
      }
      setLoading(true);
      // login
      await client.login({
        id: values.userid,
        password: values.password,
      })
      // redirect
      const urlParams = new URLSearchParams(window.location.search);
      const redirectURL = urlParams.get('redirect_to') || '/';
      window.location.href = redirectURL;
    }
    catch (error) {
      console.error(error);
      if (error instanceof Error) {
        setSigninResult(error.message);
      }
    }
    finally {
      setLoading(false);
    }
  }

  return (
    <Box sx={{
      display: 'flex',
      bgcolor: (theme) => theme.palette.mode === 'light' ? theme.palette.grey[100] : theme.palette.grey[900],
    }}>
      <AppBar position="absolute">
        <Toolbar ><img alt="cosmo" src={logo} height={40} /></Toolbar>
      </AppBar>
      <ErrorBoundary
        FallbackComponent={
          ({ error, resetErrorBoundary }) => {
            return (
              <div>
                <p>Something went wrong:</p>
                <pre>{error.message}</pre>
              </div>
            )
          }
        }
      >
        <Stack
          component="main"
          sx={{
            flexGrow: 1,
            height: '100vh',
            overflow: 'auto',
            pt: 15, pb: 4,
          }}
        >
          <Container maxWidth="xs">
            <Stack sx={{ alignItems: 'center' }}>
              <Avatar sx={{ m: 1, bgcolor: 'secondary.main' }}>
                <LockOutlinedIcon />
              </Avatar>
              <Typography color="textPrimary" variant="h5">Sign In</Typography>
              <Typography color="textPrimary" variant="body1">cosmo-auth-proxy</Typography>
              <form noValidate onKeyDown={(e) => { if (e.key === 'Enter') login() }}>
                <TextField label="User ID" margin="normal" fullWidth autoComplete="userid" autoFocus
                  error={Boolean(errors.userid)} helperText={errors.userid}
                  value={values.userid} onChange={e => setValue({ userid: e.target.value })}
                />
                <PasswordTextField label="Password" margin="normal" fullWidth autoComplete="current-password"
                  error={Boolean(errors.password)} helperText={errors.password}
                  value={values.password} onChange={e => setValue({ password: e.target.value })}
                />
                <Button fullWidth variant="contained" sx={{ mt: 3 }}
                  onClick={() => { login() }
                  }>Authenticate</Button>
              </form>
            </Stack>
            <Box pt={6}>
              <Copyright />
            </Box>
          </Container>
        </Stack>
        <Snackbar anchorOrigin={{ vertical: 'bottom', horizontal: 'center' }}
          open={Boolean(signinResult)} autoHideDuration={3000} onClose={() => { setSigninResult("") }} >
          <Alert elevation={1} severity="error" variant="filled">{signinResult}</Alert>
        </Snackbar>
        <Backdrop sx={{ zIndex: (theme) => theme.zIndex.drawer + 1000 }} open={loading}>
          <CircularProgress />
        </Backdrop>
      </ErrorBoundary>
    </Box>
  );
};
