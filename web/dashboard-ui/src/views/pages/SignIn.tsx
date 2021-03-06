import LockOutlinedIcon from '@mui/icons-material/LockOutlined';
import { Avatar, Button, Container, CssBaseline, Stack, TextField, Typography } from '@mui/material';
import React from 'react';
import { useForm, UseFormRegisterReturn } from 'react-hook-form';
import { useLocation, useNavigate } from 'react-router-dom';
import { useLogin } from '../../components/LoginProvider';
import { PasswordTextField } from '../atoms/PasswordTextField';
import { PasswordChangeDialogContext } from '../organisms/PasswordChangeDialog';
import { PageTemplate } from '../templates/PageTemplate';

const registerMui = ({ ref, ...rest }: UseFormRegisterReturn) => ({
  inputRef: ref, ...rest
});

interface Inputs {
  userid: string,
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
  const { login } = useLogin();
  const navigate = useNavigate();
  const location = useLocation();

  const passwordChangeDialogDispach = PasswordChangeDialogContext.useDispatch();

  /**
   * submit
   */
  const onSignIn = async (data: Inputs) => {
    console.log('onSignIn');
    const { requirePasswordUpdate } = await login(data.userid, data.password);
    if (requirePasswordUpdate) {
      passwordChangeDialogDispach(true);
    }
    let _location = location;
    let _route = '/workspace'
    if (_location.state && (_location.state as any).from) {
      _route = (_location.state as any).from.pathname;
    }
    console.log(_route);
    navigate(_route);
  }

  return (
    <Stack sx={{ marginTop: 3, alignItems: 'center' }}>
      <Avatar sx={{ m: 1, bgcolor: 'secondary.main' }}>
        <LockOutlinedIcon />
      </Avatar>
      <Typography component="h1" variant="h5">Sign in</Typography>
      <Typography color="textPrimary" variant="body1">cosmo-dashboard</Typography>
      <form noValidate onSubmit={handleSubmit(onSignIn)}>
        <TextField label="User ID" margin="normal" fullWidth autoComplete="userid" autoFocus
          {...registerMui(register("userid", {
            required: { value: true, message: "Required" },
            pattern: {
              value: /^[a-z0-9]([-a-z0-9]*[a-z0-9])?$/,
              message: 'Only lowercase alphanumeric characters or in "-" are allowed'
            },
            maxLength: { value: 128, message: "Max 128 characters" },
          }))}
          error={Boolean(errors.userid)}
          helperText={(errors.userid && errors.userid.message)}
        />
        <PasswordTextField label="Password" margin="normal" fullWidth autoComplete="current-password"
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
