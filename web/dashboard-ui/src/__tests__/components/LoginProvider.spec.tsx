import { act, renderHook, RenderHookResult } from "@testing-library/react-hooks";
import { AxiosResponse } from "axios";
import { useSnackbar } from "notistack";
import React from 'react';
import { AuthApiFactory, UserApiFactory } from "../../api/dashboard/v1alpha1";
import { LoginProvider, useLogin } from "../../components/LoginProvider";
import { useProgress } from "../../components/ProgressProvider";

//--------------------------------------------------
// mock definition
//--------------------------------------------------

jest.mock('notistack');
jest.mock('../../api/dashboard/v1alpha1/api');
jest.mock('../../components/ProgressProvider');
jest.mock('react-router-dom', () => ({
  useHistory: jest.fn(),
}),
);


type MockedMemberFunction<T extends (...args: any) => any> = {
  [P in keyof ReturnType<T>]: jest.MockedFunction<ReturnType<T>[P]>;
};

const useSnackbarMock = useSnackbar as jest.MockedFunction<typeof useSnackbar>;
const snackbarMock: MockedMemberFunction<typeof useSnackbar> = {
  enqueueSnackbar: jest.fn(),
  closeSnackbar: jest.fn(),
}

const useProgressMock = useProgress as jest.MockedFunction<typeof useProgress>;
const progressMock: MockedMemberFunction<typeof useProgress> = {
  setMask: jest.fn(),
  releaseMask: jest.fn(),
}

const restAuthMock = AuthApiFactory as jest.MockedFunction<typeof AuthApiFactory>;
const authMock: MockedMemberFunction<typeof AuthApiFactory> = {
  verify: jest.fn(),
  login: jest.fn(),
  logout: jest.fn(),
}

const RestUserMock = UserApiFactory as jest.MockedFunction<typeof UserApiFactory>;
const userMock: MockedMemberFunction<typeof UserApiFactory> = {
  postUser: jest.fn(),
  putUserRole: jest.fn(),
  putUserPassword: jest.fn(),
  getUser: jest.fn(),
  getUsers: jest.fn(),
  deleteUser: jest.fn(),
}

function axiosNormalResponse<T>(data: T): AxiosResponse<T> {
  return { data: data, status: 200, statusText: 'ok', headers: {}, config: {}, request: {} }
}

//-----------------------------------------------
// test
//-----------------------------------------------

describe('LoginProvider', () => {

  beforeEach(async () => {
    useSnackbarMock.mockReturnValue(snackbarMock);
    useProgressMock.mockReturnValue(progressMock);
    restAuthMock.mockReturnValue(authMock);
    RestUserMock.mockReturnValue(userMock);
  });

  afterEach(() => {
    expect(snackbarMock.enqueueSnackbar.mock.calls).toMatchSnapshot('enqueueSnackbar calls');
    jest.restoreAllMocks();
  });


  describe('verify', () => {

    it('ok', async () => {
      authMock.verify.mockResolvedValueOnce(axiosNormalResponse({
        id: 'user1',
        expireAt: 'xxxx',
        requirePasswordUpdate: false,
      }));
      userMock.getUser.mockResolvedValue(axiosNormalResponse({
        message: "ok",
        user: { id: 'user1', role: 'cosmo-admin', displayName: 'user1 name' },
      }));

      const { result, waitForNextUpdate } = renderHook(() => useLogin(), {
        wrapper: ({ children }) => (<LoginProvider>{children}</LoginProvider >),
      });
      await waitForNextUpdate();
      expect(result.current.loginUser).toMatchSnapshot();
    });


    it('ng / id is undefined', async () => {
      authMock.verify.mockResolvedValueOnce(axiosNormalResponse({
        id: undefined,
        expireAt: 'xxxx',
        requirePasswordUpdate: false,
      }));

      const { result, waitForNextUpdate } = renderHook(() => useLogin(), {
        wrapper: ({ children }) => (<LoginProvider>{children}</LoginProvider >),
      });
      await waitForNextUpdate();
      expect(result.current.loginUser).toMatchSnapshot();
    });


    it('ng', async () => {

      authMock.verify.mockRejectedValue({ response: { data: { message: 'data.message' }, status: 401 } as any } as any);
      const { result, waitForNextUpdate } = renderHook(() => useLogin(), {
        wrapper: ({ children }) => (<LoginProvider>{children}</LoginProvider >),
      });
      await waitForNextUpdate();
      expect(result.current.loginUser).toMatchSnapshot();
    });

    it('getUser error', async () => {

      authMock.verify.mockResolvedValueOnce(axiosNormalResponse({
        id: 'user1',
        expireAt: 'xxxx',
        requirePasswordUpdate: false,
      }));

      userMock.getUser.mockRejectedValue({ response: { data: { message: 'getUser error' }, status: 401 } as any } as any);

      const { result, waitForNextUpdate } = renderHook(() => useLogin(), {
        wrapper: ({ children }) => (<LoginProvider>{children}</LoginProvider >),
      });
      await waitForNextUpdate();
      expect(result.current.loginUser).toMatchSnapshot();
    });

  });


  describe('login', () => {

    let hookResult: RenderHookResult<unknown, ReturnType<typeof useLogin>>;

    beforeEach(async () => {
      authMock.verify.mockRejectedValue({ response: { data: { message: 'verify error' }, status: 401 } as any } as any);
      hookResult = renderHook(() => useLogin(), {
        wrapper: ({ children }) => (<LoginProvider>{children}</LoginProvider >),
      });
      await hookResult.waitForNextUpdate();
      // verify error
    });


    it('ok', async () => {
      const { result } = hookResult;

      authMock.login.mockResolvedValue(axiosNormalResponse({
        id: 'user1',
        expireAt: 'xxxx',
        requirePasswordUpdate: true,
      }));
      userMock.getUser.mockResolvedValue(axiosNormalResponse({
        message: "ok",
        user: { id: 'user1', role: 'cosmo-admin', displayName: 'user1 name' },
      }));

      await act(async () => {
        await expect(result.current.login('user1', 'password1')).resolves.toMatchSnapshot();
        expect(result.current.loginUser).toMatchSnapshot();
      });

      await act(async () => {
        await expect(result.current.login('user1', 'password1')).resolves.toMatchSnapshot();
        expect(result.current.loginUser).toMatchSnapshot();
      });
    });


    it('ng', async () => {
      const { result } = hookResult;

      authMock.login.mockResolvedValue(axiosNormalResponse({
        id: 'user1',
        expireAt: 'xxxx',
        requirePasswordUpdate: true,
      }));
      userMock.getUser.mockResolvedValue(axiosNormalResponse({
        message: "ok",
        user: { id: 'user1', role: 'cosmo-admin', displayName: 'user1 name' },
      }));

      authMock.login.mockRejectedValue({ response: { data: { message: 'login error' }, status: 401 } as any } as any);

      await act(async () => {
        await expect(result.current.login('user1', 'password1')).rejects.toMatchSnapshot();
        expect(result.current.loginUser).toMatchSnapshot();
      });
    });


    it('getUser error', async () => {
      const { result } = hookResult;

      authMock.login.mockResolvedValue(axiosNormalResponse({
        id: 'user1',
        expireAt: 'xxxx',
        requirePasswordUpdate: true,
      }));

      userMock.getUser.mockRejectedValue({ response: { data: { message: 'getUser error' }, status: 401 } as any } as any);

      await act(async () => {
        await expect(result.current.login('user1', 'password1')).rejects.toMatchSnapshot();
        expect(result.current.loginUser).toMatchSnapshot();
      });
    });

  });


  describe('logout', () => {

    let hookResult: RenderHookResult<unknown, ReturnType<typeof useLogin>>;

    beforeEach(async () => {
      authMock.verify.mockResolvedValueOnce(axiosNormalResponse({
        id: 'user1',
        expireAt: 'xxxx',
        requirePasswordUpdate: false,
      }));
      userMock.getUser.mockResolvedValue(axiosNormalResponse({
        message: "ok",
        user: { id: 'user1', role: 'cosmo-admin', displayName: 'user1 name' },
      }));
      hookResult = renderHook(() => useLogin(), {
        wrapper: ({ children }) => (<LoginProvider>{children}</LoginProvider >),
      });
      await hookResult.waitForNextUpdate();
    });


    it('ok', async () => {
      const { result } = hookResult;
      await act(async () => {
        await expect(result.current.logout()).resolves.toMatchSnapshot();
        expect(result.current.loginUser).toMatchSnapshot();
        expect(authMock.logout.mock.calls).toMatchSnapshot('logout calls');
      });
    });

    it('error', async () => {
      const { result } = hookResult;
      authMock.logout.mockRejectedValue({ response: { data: { message: 'logout error' }, status: 500 } as any } as any);
      await act(async () => {
        await expect(result.current.logout()).rejects.toMatchSnapshot();
        expect(result.current.loginUser).toMatchSnapshot();
        expect(authMock.logout.mock.calls).toMatchSnapshot('logout calls');
      });
    });

  });


  describe('updataPassword', () => {

    let hookResult: RenderHookResult<unknown, ReturnType<typeof useLogin>>;

    beforeEach(async () => {
      authMock.verify.mockResolvedValueOnce(axiosNormalResponse({
        id: 'user1',
        expireAt: 'xxxx',
        requirePasswordUpdate: false,
      }));
      userMock.getUser.mockResolvedValue(axiosNormalResponse({
        message: "ok",
        user: { id: 'user1', role: 'cosmo-admin', displayName: 'user1 name' },
      }));
      hookResult = renderHook(() => useLogin(), {
        wrapper: ({ children }) => (<LoginProvider>{children}</LoginProvider >),
      });
      await hookResult.waitForNextUpdate();
    });

    it('ok', async () => {
      const { result } = hookResult;
      userMock.putUserPassword.mockResolvedValue(axiosNormalResponse({
        message: "ok",
        user: { id: 'user1', role: 'cosmo-admin', displayName: 'user1 name' },
      }));
      await act(async () => {
        await expect(result.current.updataPassword('oldpw', 'newpw')).resolves.toMatchSnapshot();
        expect(userMock.putUserPassword.mock.calls).toMatchSnapshot('putUserPassword calls');
      });
    });

    it('error', async () => {
      const { result } = hookResult;
      userMock.putUserPassword.mockRejectedValue({ message: 'updataPassword error' } as any);
      await act(async () => {
        await expect(result.current.updataPassword('oldpw', 'newpw')).rejects.toMatchSnapshot();
        expect(userMock.putUserPassword.mock.calls).toMatchSnapshot('putUserPassword calls');
      });
    });

  });

});
