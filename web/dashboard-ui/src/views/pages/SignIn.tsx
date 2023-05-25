import LockOutlinedIcon from '@mui/icons-material/LockOutlined';
import { Avatar, Button, Container, CssBaseline, Stack, TextField, Typography } from '@mui/material';
import React from 'react';
import { UseFormRegisterReturn, useForm } from 'react-hook-form';
import { useLocation, useNavigate } from 'react-router-dom';
import { useLogin } from '../../components/LoginProvider';
import { PasswordTextField } from '../atoms/PasswordTextField';
import { PasswordChangeDialogContext } from '../organisms/PasswordChangeDialog';
import { PageTemplate } from '../templates/PageTemplate';

const registerMui = ({ ref, ...rest }: UseFormRegisterReturn) => ({
  inputRef: ref, ...rest
});

export const extractDomainFromHostname = (hostname: string) => {
  return hostname.slice(hostname.indexOf('.') + 1)
}

export const isValidRedirectURLDomain = (url: string) => {
  // check redirect domain is the same as current domain
  const redirect_hostname = new URL(url).hostname;
  const current_hostname = window.location.hostname;
  return extractDomainFromHostname(redirect_hostname) === extractDomainFromHostname(current_hostname);
}

interface Inputs {
  username: string,
  password: string,
};

/**
 * view
 */
export function SignIn() {
  console.log('SignInContent');

  return (
    <PageTemplate title="">
      <Container component="main" maxWidth="xs">
        <CssBaseline />
        <SignInContent />
      </Container>
    </PageTemplate>
  );
};

const SignInContent: React.VFC = () => {
  console.log('SignIn');
  const { register, handleSubmit, formState: { errors } } = useForm<Inputs>();
  const { login, loginUser } = useLogin();
  const navigate = useNavigate();
  const location = useLocation();

  const passwordChangeDialogDispach = PasswordChangeDialogContext.useDispatch();

  const redirect = () => {
    const query = new URLSearchParams(location.search);
    let redirect_to = query.get('redirect_to')
    if (redirect_to) {
      console.log('redirect_to=' + redirect_to);
      if (isValidRedirectURLDomain(redirect_to)) {
        window.location.href = redirect_to;
        return;
      }
      throw new Error("invalid redirect URL");
    }

    let _location = location;
    let _route = '/workspace';
    if (_location.state && (_location.state as any).from) {
      _route = (_location.state as any).from.pathname;
    }
    console.log(_route);
    navigate(_route);
  }
  React.useEffect(() => { loginUser && redirect() }, []);

  /**
   * submit
   */
  const onSignIn = async (data: Inputs) => {
    console.log('onSignIn');
    const { requirePasswordUpdate } = await login(data.username, data.password);
    if (requirePasswordUpdate) {
      passwordChangeDialogDispach(true);
    }
    redirect();
  }

  return (
    <Stack sx={{ marginTop: 3, alignItems: 'center' }}>
      <Avatar sx={{ m: 1, bgcolor: 'secondary.main' }}>
        <LockOutlinedIcon />
      </Avatar>
      <Typography component="h1" variant="h5">Sign in</Typography>
      <Typography color="textPrimary" variant="body1">cosmo-dashboard</Typography>
      <form noValidate onSubmit={handleSubmit(onSignIn)}>
        <TextField label="User ID" margin="normal" fullWidth autoComplete="username" autoFocus defaultValue={loginUser?.name}
          {...registerMui(register("username", {
            required: { value: true, message: "Required" },
            pattern: {
              value: /^[a-z0-9]([-a-z0-9]*[a-z0-9])?$/,
              message: 'Only lowercase alphanumeric characters or in "-" are allowed'
            },
            maxLength: { value: 128, message: "Max 128 characters" },
          }))}
          error={Boolean(errors.username)}
          helperText={(errors.username && errors.username.message)}
        />
        <PasswordTextField label="Password" margin="normal" fullWidth autoComplete="current-password" defaultValue={loginUser && "you-are-logined"}
          {...registerMui(register("password", {
            required: { value: true, message: "Required" },
            maxLength: { value: 128, message: "Max 128 characters" },
          }))}
          error={Boolean(errors.password)}
          helperText={(errors.password && errors.password.message)}
        />
        <Button type='submit' fullWidth variant="contained" sx={{ mt: 3, mb: 2 }}>Sign In</Button>
      </form>
    </Stack>
  );
}
