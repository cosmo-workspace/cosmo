import { Box, CircularProgress, CssBaseline, Stack, Typography } from '@mui/material';
import { useSnackbar } from 'notistack';
import React, { createContext, useContext, useEffect, useState } from 'react';
import { User } from '../proto/gen/dashboard/v1alpha1/user_pb';
import { useAuthService, useUserService, useWebAuthnService } from '../services/DashboardServices';
import { base64url } from './Base64';
import { useProgress } from './ProgressProvider';


/**
 * context
 * ref: https://github.com/DefinitelyTyped/DefinitelyTyped/pull/24509#issuecomment-774430643
 */
const Context = createContext<ReturnType<typeof useLoginModule>>(undefined as any);

/**
 * hooks
 */
export function useLogin() {
  return useContext(Context);
}

const useLoginModule = () => {
  console.log('useLoginModule');
  const [loginUser, setLoginUser] = useState<User>();
  const { enqueueSnackbar } = useSnackbar();
  const { setMask, releaseMask } = useProgress();
  const userService = useUserService();
  const authService = useAuthService();
  const webauthnService = useWebAuthnService();

  /**
   * loginWithWebAuthn
   */
  const loginWithWebAuthn = async (userName: string) => {
    console.log('loginWithWebAuthn start');
    try {
      const credId = localStorage.getItem(`credId`);
      if (credId === null) {
        throw Error('credId is null');
      }

      const beginResp = await webauthnService.beginLogin({ userName: userName });
      const options = JSON.parse(beginResp.credentialRequestOptions);

      const opt: CredentialRequestOptions = JSON.parse(JSON.stringify(options));

      if (options.publicKey?.challenge) {
        opt.publicKey!.challenge = base64url.decode(options.publicKey?.challenge);
      }

      let allowed = false;
      for (let index = 0; index < options.publicKey?.allowCredentials.length; index++) {
        if (options.publicKey?.allowCredentials[index].id === credId) { allowed = true };
        if (options.publicKey?.allowCredentials) {
          opt.publicKey!.allowCredentials![index].id = base64url.decode(options.publicKey?.allowCredentials[index].id);
        }
      }
      if (!allowed) { throw Error('invalid credentials'); }

      // Credential is allowed to access only id and type so use any.
      const cred: any = await navigator.credentials.get(opt);
      if (cred === null) {
        console.log("cred is null");
        throw Error('credential is null');
      }

      const credential = {
        id: cred.id,
        type: cred.type,
        rawId: base64url.encode(cred.rawId),
        response: {
          clientDataJSON: base64url.encode(cred.response.clientDataJSON),
          authenticatorData: base64url.encode(cred.response.authenticatorData),
          signature: base64url.encode(cred.response.signature),
          userHandle: base64url.encode(cred.response.userHandle),
        },
      };

      const finResp = await webauthnService.finishLogin({ userName: userName, credentialRequestResult: JSON.stringify(credential) });
      await getMyUserInfo(userName);
      console.log('loginWithWebAuthn end', finResp);
      return;
    }
    catch (error) {
      setLoginUser(undefined);
      error instanceof DOMException || handleError(error);
      throw error;
    }
  }


  /**
   * verifyLogin
   */
  const verifyLogin = async () => {
    console.log('verify start');
    try {
      const resp = await authService.verify({});
      if (resp.userName) {
        await getMyUserInfo(resp.userName);
      }
    }
    catch (error) {
      setLoginUser(undefined);
      //handleError(error);
    }
    finally {
      console.log('verify end');
    }
  }


  /**
   * login: SignIn 
   */
  const login = async (userName: string, password: string) => {
    console.log('login');
    try {
      const res = await authService.login({ userName: userName, password: password });
      await getMyUserInfo(userName);
      console.log('login end');
      return res;
    }
    catch (error) {
      setLoginUser(undefined);
      console.log('login error');
      handleError(error);
      throw error;
    }
  }


  /**
   * login: MyUserInfo
   */
  const getMyUserInfo = async (userName: string) => {
    console.log('getMyUserInfo', userName);
    // if (loginUser || !username) {
    //   console.log('getMyUserInfo cancel', loginUser, username);
    //   return;
    // }
    try {
      const responseUser = await userService.getUser({ userName: userName });
      setLoginUser(prev => { console.log('setLoginUser', prev, responseUser.user); return responseUser.user! });
    }
    catch (error) {
      console.log('getMyUserInfo error');
      handleError(error);
      throw error;
    }
  }

  /**
   * login: MyUserInfo
   */
  const refreshUserInfo = async () => {
    console.log('refreshUserInfo');
    if (loginUser) {
      getMyUserInfo(loginUser.name);
    }
  }

  /**
   * logout: Pagetemplate 
   */
  const logout = async () => {
    console.log('logout');
    try {
      await authService.logout({});
    }
    catch (error) {
      console.log('logout error');
      handleError(error);
      throw error;
    }
    finally {
      console.log('logout end');
      setLoginUser(prev => { console.log('setLoginUser', `${prev} -> undefined`); return undefined });
    }
  }

  /**
   * updataPassword
   */
  const updataPassword = async (currentPassword: string, newPassword: string) => {
    console.log('updataPassword', loginUser?.name);
    setMask();
    try {
      try {
        return await userService.updateUserPassword({ userName: loginUser!.name, currentPassword, newPassword });
      }
      catch (error) {
        handleError(error);
        throw error;
      }
    }
    finally {
      console.log('updataPassword end');
      releaseMask();
    }
  }

  /**
   * clearLoginUser 
   */
  const clearLoginUser = async () => {
    setLoginUser(undefined);
  }

  /**
   * error handler
   */
  const handleError = (error: any) => {
    console.log(error);
    const msg = error?.message;
    msg && enqueueSnackbar(msg, { variant: 'error' });
  }

  return {
    loginUser,
    verifyLogin,
    login,
    loginWithWebAuthn,
    logout,
    refreshUserInfo,
    updataPassword,
    clearLoginUser,
  };
}

/**
 * Provider
 */
export const LoginProvider: React.FC<React.PropsWithChildren<unknown>> = ({ children }) => {
  console.log('LoginProvider');
  const loginModule = useLoginModule();
  const [isVerified, setIsVerified] = useState(false);

  useEffect(() => {
    loginModule.verifyLogin()
      .then(() => setIsVerified(true));
  }, []); // eslint-disable-line

  return (
    <Context.Provider value={loginModule}>
      {isVerified ? children :
        <div>
          <CssBaseline />
          <Stack sx={{ m: 10 }} alignItems='center' spacing={2}>
            <Box sx={{ display: 'flex' }}>
              <CircularProgress />
            </Box>
            <Typography >verifying session...</Typography>
          </Stack>
        </div>}
    </Context.Provider >
  )
}

/**
 * export private member. (for test)
 */
// export const __local__ = {
//   useLoginModule,
// };
