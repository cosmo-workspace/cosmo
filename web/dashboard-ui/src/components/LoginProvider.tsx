import { useSnackbar } from 'notistack';
import React, { useContext, useEffect, useState } from 'react';
import { AuthApiFactory, User, UserApiFactory } from '../api/dashboard/v1alpha1';
import { useProgress } from './ProgressProvider';

/**
 * context
 * ref: https://github.com/DefinitelyTyped/DefinitelyTyped/pull/24509#issuecomment-774430643
 */
const Context = React.createContext<ReturnType<typeof useLoginModule>>(undefined as any);

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
  const { getUser, putUserPassword } = UserApiFactory(undefined, "");

  /**
   * verifyLogin
   */
  const verifyLogin = async () => {
    console.log('verify start');
    try {
      const restAuth = AuthApiFactory(undefined, "");
      const resp = await restAuth.verify();
      if (resp.data.id) {
        await getMyUserInfo(resp.data.id);
      }
    }
    catch (error) {
      //handleError(error);
    }
    finally {
      console.log('verify end');
    }
  }


  /**
   * login: SignIn 
   */
  const login = async (id: string, password: string) => {
    console.log('login');
    try {
      const restAuth = AuthApiFactory(undefined, "");
      const res = await restAuth.login({ id, password });
      await getMyUserInfo(id);
      return res.data;
    }
    catch (error) {
      setLoginUser(undefined);
      console.log('login error');
      handleError(error);
      throw error;
    }
    finally {
      console.log('login end');
    }

  }


  /**
   * login: MyUserInfo
   */
  const getMyUserInfo = async (userId: string) => {
    console.log('getMyUserInfo', userId);
    if (loginUser || !userId) {
      console.log('getMyUserInfo cancel', loginUser, userId);
      return;
    }
    try {
      const responseUser = await getUser(userId);
      setLoginUser(prev => { console.log('setLoginUser', prev, responseUser.data.user); return responseUser.data.user! });
    }
    catch (error) {
      console.log('getMyUserInfo error');
      handleError(error);
      throw error;
    }
  }


  /**
   * logout: Pagetemplate 
   */
  const logout = async () => {
    console.log('logout');
    try {
      const restAuth = AuthApiFactory(undefined, "");
      await restAuth.logout();
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
    console.log('updataPassword', loginUser?.id);
    setMask();
    try {
      return await putUserPassword(loginUser!.id, { currentPassword, newPassword });
    }
    catch (error) {
      handleError(error);
      throw error;
    }
    finally {
      console.log('updataPassword end');
      releaseMask();
    }
  }


  /**
   * error handler
   */
  const handleError = (error: any) => {
    console.log(error);
    const msg = error?.response?.data?.message || error?.message;
    msg && enqueueSnackbar(msg, { variant: 'error' });
  }

  return {
    loginUser,
    verifyLogin,
    login,
    logout,
    updataPassword,
  };
}

/**
 * Provider
 */
export const LoginProvider: React.FC = ({ children }) => {
  console.log('LoginProvider');
  const loginModule = useLoginModule();
  const [isVerified, setIsVerified] = useState(false);

  useEffect(() => {
    loginModule.verifyLogin()
      .then(() => setIsVerified(true));
  }, []); // eslint-disable-line

  return (
    <Context.Provider value={loginModule}>
      {isVerified ? children : <div>...verify login...</div>}
    </Context.Provider>
  )
}

/**
 * export private member. (for test) 
 */
export const __local__ = {
  useLoginModule,
};
