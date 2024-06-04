import { Code, ConnectError } from "@bufbuild/connect";
import {
  Box,
  CircularProgress,
  CssBaseline,
  Stack,
  Typography,
} from "@mui/material";
import { useSnackbar } from "notistack";
import React, { createContext, useContext, useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { Event } from "../proto/gen/dashboard/v1alpha1/event_pb";
import { User } from "../proto/gen/dashboard/v1alpha1/user_pb";
import {
  useAuthService,
  useStreamService,
  useUserService,
  useWebAuthnService,
} from "../services/DashboardServices";
import { latestTime } from "../views/organisms/EventModule";
import { base64url } from "./Base64";
import { useProgress } from "./ProgressProvider";

/**
 * context
 * ref: https://github.com/DefinitelyTyped/DefinitelyTyped/pull/24509#issuecomment-774430643
 */
const Context = createContext<ReturnType<typeof useLoginModule>>(
  undefined as any,
);

/**
 * hooks
 */
export function useLogin() {
  return useContext(Context);
}

const useLoginModule = () => {
  console.log("useLoginModule");
  const [loginUser, setLoginUser] = useState<User>();
  const isSignIn = Boolean(loginUser);

  const { enqueueSnackbar } = useSnackbar();
  const { setMask, releaseMask } = useProgress();
  const userService = useUserService();
  const authService = useAuthService();
  const webauthnService = useWebAuthnService();

  const [myEvents, setMyEvents] = useState<Event[]>([]);
  const [newEventsCount, setNewEventsCount] = useState(0);
  const [clock, setClock] = useState(new Date());

  const streamService = useStreamService();
  const [isWatching, setIsWatching] = React.useState(false);

  const updateClock = () => setClock(new Date());

  const handleMyEvents = (events: Event[]) => {
    for (const event of events) {
      const index = myEvents.findIndex((e) => (e.id === event.id));
      if (index >= 0) {
        // replace event
        console.log("!!! replace", event.id, index);
        myEvents[index] = event;
        myEvents.sort((a, b) => latestTime(b) - latestTime(a));
        setMyEvents(myEvents);
      } else {
        // put event on the top of event list
        console.log("!!! new", event.id);
        myEvents.push(event);
        myEvents.sort((a, b) => latestTime(b) - latestTime(a));
        setMyEvents(myEvents);
      }
    }
  };

  const watchMyEvents = async () => {
    if (isSignIn) {
      const watchEvents = async (retryCount: number) => {
        console.log("Start watching events...", loginUser?.name, retryCount);
        try {
          const result = await streamService.streamingEvents({
            userName: loginUser?.name,
          }, {});
          for await (const event of result) {
            updateClock();
            setNewEventsCount((v) => v + 1);
            handleMyEvents(event.items);
            retryCount = 0;
          }
        } catch (error) {
          console.log("watch failed", error, "retryCount", retryCount);
          if (retryCount < 10) {
            retryCount += 1;
            const waitTime = Math.floor(Math.random() * 5 + 1) * 1000;
            console.log("watch re-conntect after milisec", waitTime);
            setTimeout(() => watchEvents(retryCount), waitTime);
          } else {
            console.log("Reached retry limit for watching events");
          }
        }
      };
      if (!isWatching) {
        setIsWatching(true);
        watchEvents(0);
        setIsWatching(false);
      }
    }
  };

  /**
   * loginWithWebAuthn
   */
  const loginWithWebAuthn = async (userName: string) => {
    console.log("loginWithWebAuthn start");
    try {
      const credId = localStorage.getItem(`credId`);
      if (credId === null) {
        throw Error("credId is null");
      }

      const beginResp = await webauthnService.beginLogin({
        userName: userName,
      });
      const options = JSON.parse(beginResp.credentialRequestOptions);

      const opt: CredentialRequestOptions = JSON.parse(JSON.stringify(options));

      if (options.publicKey?.challenge) {
        opt.publicKey!.challenge = base64url.decode(
          options.publicKey?.challenge,
        );
      }

      let allowed = false;
      for (
        let index = 0;
        index < options.publicKey?.allowCredentials.length;
        index++
      ) {
        if (options.publicKey?.allowCredentials[index].id === credId) {
          allowed = true;
        }
        if (options.publicKey?.allowCredentials) {
          opt.publicKey!.allowCredentials![index].id = base64url.decode(
            options.publicKey?.allowCredentials[index].id,
          );
        }
      }
      if (!allowed) throw Error("invalid credentials");

      // Credential is allowed to access only id and type so use any.
      const cred: any = await navigator.credentials.get(opt);
      if (cred === null) {
        console.log("cred is null");
        throw Error("credential is null");
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

      const finResp = await webauthnService.finishLogin({
        userName: userName,
        credentialRequestResult: JSON.stringify(credential),
      });
      await getMyUserInfo(userName);
      console.log("loginWithWebAuthn end", finResp);
      return;
    } catch (error) {
      setLoginUser(undefined);
      error instanceof DOMException || handleError(error);
      throw error;
    }
  };

  /**
   * verifyLogin
   */
  const verifyLogin = async () => {
    console.log("verify start");
    try {
      const resp = await authService.verify({});
      if (resp.userName) {
        await getMyUserInfo(resp.userName);
      }
    } catch (error) {
      setLoginUser(undefined);
      //handleError(error);
    } finally {
      console.log("verify end");
    }
  };

  /**
   * login: SignIn
   */
  const login = async (userName: string, password: string) => {
    console.log("login");
    try {
      const res = await authService.login({
        userName: userName,
        password: password,
      });
      await getMyUserInfo(userName);
      console.log("login end");
      return res;
    } catch (error) {
      setLoginUser(undefined);
      console.log("login error");
      handleError(error);
      throw error;
    }
  };

  /**
   * login: MyUserInfo
   */
  const getMyUserInfo = async (userName: string) => {
    console.log("getMyUserInfo", userName);
    // if (loginUser || !username) {
    //   console.log('getMyUserInfo cancel', loginUser, username);
    //   return;
    // }
    try {
      const responseUser = await userService.getUser({ userName: userName });
      setLoginUser((prev) => {
        console.log("setLoginUser", prev, responseUser.user);
        return responseUser.user!;
      });
    } catch (error) {
      console.log("getMyUserInfo error");
      handleError(error);
      throw error;
    }
  };

  /**
   * login: MyUserInfo
   */
  const refreshUserInfo = async () => {
    console.log("refreshUserInfo");
    if (loginUser) {
      getMyUserInfo(loginUser.name);
    }
  };

  /**
   * logout: Pagetemplate
   */
  const logout = async () => {
    console.log("logout");
    try {
      await authService.logout({});
    } catch (error) {
      console.log("logout error");
      handleError(error);
      throw error;
    } finally {
      console.log("logout end");
      setLoginUser((prev) => {
        console.log("setLoginUser", `${prev} -> undefined`);
        return undefined;
      });
    }
  };

  /**
   * updataPassword
   */
  const updataPassword = async (
    currentPassword: string,
    newPassword: string,
  ) => {
    console.log("updataPassword", loginUser?.name);
    setMask();
    try {
      try {
        return await userService.updateUserPassword({
          userName: loginUser!.name,
          currentPassword,
          newPassword,
        });
      } catch (error) {
        handleError(error);
        throw error;
      }
    } finally {
      console.log("updataPassword end");
      releaseMask();
    }
  };

  /**
   * clearLoginUser
   */
  const clearLoginUser = async () => {
    setLoginUser(undefined);
  };

  const getMyEvents = async () => {
    if (loginUser) {
      try {
        const result = await userService.getEvents({
          userName: loginUser?.name,
        });
        handleMyEvents(result.items);
      } catch (error) {
        handleError(error);
      }
    } else {
      console.error("getEvents called before login");
    }
  };

  /**
   * error handler
   */
  const handleError = (error: any) => {
    console.log(error);
    const msg = error?.message;
    msg && enqueueSnackbar(msg, { variant: "error" });
  };

  return {
    loginUser,
    verifyLogin,
    login,
    loginWithWebAuthn,
    logout,
    refreshUserInfo,
    updataPassword,
    clearLoginUser,
    myEvents,
    getMyEvents,
    watchMyEvents,
    newEventsCount,
    setNewEventsCount,
    clock,
    updateClock,
  };
};

/**
 * Provider
 */
export const LoginProvider: React.FC<React.PropsWithChildren<unknown>> = (
  { children },
) => {
  console.log("LoginProvider");
  const loginModule = useLoginModule();
  const [isVerified, setIsVerified] = useState(false);

  useEffect(() => {
    loginModule.verifyLogin()
      .then(() => setIsVerified(true));
  }, []); // eslint-disable-line

  return (
    <Context.Provider value={loginModule}>
      {isVerified ? children : (
        <div>
          <CssBaseline />
          <Stack sx={{ m: 10 }} alignItems="center" spacing={2}>
            <Box sx={{ display: "flex" }}>
              <CircularProgress />
            </Box>
            <Typography>verifying session...</Typography>
          </Stack>
        </div>
      )}
    </Context.Provider>
  );
};

/**
 * error handler
 */
export function useHandleError() {
  const { enqueueSnackbar } = useSnackbar();
  const navigate = useNavigate();
  const { clearLoginUser } = useLogin();

  const handleError = (error: any) => {
    console.log("handleError", error, "metadata", error?.metadata);

    if (
      error instanceof ConnectError &&
      (error.code === Code.Unauthenticated ||
        (error.message.includes("302") && error.code === Code.Unknown))
    ) {
      clearLoginUser();
      navigate("/signin");
      const msg = error.message.includes("302")
        ? "session expired"
        : error?.message;
      msg && enqueueSnackbar(msg, { variant: "error" });
    } else {
      const msg = error?.message;
      msg && enqueueSnackbar(msg, { variant: "error" });
    }
    throw error;
  };
  return { handleError };
}
