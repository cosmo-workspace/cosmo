import { renderHook, waitFor } from "@testing-library/react";
import { AxiosResponse } from "axios";
import { useSnackbar } from "notistack";
import React from 'react';
import { act } from "react-dom/test-utils";
import { AuthApiFactory, UserApiFactory, UserRoleEnum } from "../../api/dashboard/v1alpha1";
import { LoginProvider, useLogin } from "../../components/LoginProvider";
import { useProgress } from "../../components/ProgressProvider";

//--------------------------------------------------
// mock definition
//--------------------------------------------------

jest.mock('notistack');
jest.mock('../../api/dashboard/v1alpha1/api');
jest.mock('../../components/ProgressProvider');
jest.mock('react-router-dom', () => ({
  //useHistory: jest.fn(),
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
  putUserName: jest.fn(),
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

  async function renderUseLogin() {
    const utils = renderHook(() => useLogin(), {
      wrapper: ({ children }) => (<LoginProvider>{children}</LoginProvider >),
    });
    await waitFor(async () => { expect(utils.result.current).not.toBeNull(); });
    return utils;
  }

  describe('verify', () => {

    it('ok', async () => {
      authMock.verify.mockResolvedValueOnce(axiosNormalResponse({
        id: 'user1',
        expireAt: 'xxxx',
        requirePasswordUpdate: false,
      }));
      userMock.getUser.mockResolvedValue(axiosNormalResponse({
        message: "ok",
        user: { id: 'user1', role: UserRoleEnum.CosmoAdmin, displayName: 'user1 name' },
      }));

      const { result } = await renderUseLogin();
      expect(result.current.loginUser).toMatchSnapshot();
    });


    it('ng / id is undefined', async () => {
      authMock.verify.mockResolvedValueOnce(axiosNormalResponse({
        id: undefined as any,
        expireAt: 'xxxx',
        requirePasswordUpdate: false,
      }));

      const { result } = await renderUseLogin();
      expect(result.current.loginUser).toMatchSnapshot();
    });


    it('ng', async () => {

      authMock.verify.mockRejectedValue({ response: { data: { message: 'data.message' }, status: 401 } as any } as any);
      const { result } = await renderUseLogin();
      expect(result.current.loginUser).toMatchSnapshot();
    });

    it('getUser error', async () => {

      authMock.verify.mockResolvedValueOnce(axiosNormalResponse({
        id: 'user1',
        expireAt: 'xxxx',
        requirePasswordUpdate: false,
      }));

      userMock.getUser.mockRejectedValue({ response: { data: { message: 'getUser error' }, status: 401 } as any } as any);

      const { result } = await renderUseLogin();
      expect(result.current.loginUser).toMatchSnapshot();
    });

  });


  describe('login', () => {

    it('ok', async () => {
      const { result } = await renderUseLogin();

      authMock.login.mockResolvedValue(axiosNormalResponse({
        id: 'user1',
        expireAt: 'xxxx',
        requirePasswordUpdate: true,
      }));
      userMock.getUser.mockResolvedValue(axiosNormalResponse({
        message: "ok",
        user: { id: 'user1', role: UserRoleEnum.CosmoAdmin, displayName: 'user1 name' },
      }));
      await act(async () => {
        await expect(result.current.login('user1', 'password1')).resolves.toMatchSnapshot();
      });
      await waitFor(async () => { expect(result.current.loginUser).not.toBeUndefined(); });
      expect(result.current.loginUser).toMatchSnapshot();
      await act(async () => {
        await expect(result.current.login('user1', 'password1')).resolves.toMatchSnapshot();
      });
      expect(result.current.loginUser).toMatchSnapshot();
    });


    it('ng', async () => {
      const { result } = await renderUseLogin();

      authMock.login.mockResolvedValue(axiosNormalResponse({
        id: 'user1',
        expireAt: 'xxxx',
        requirePasswordUpdate: true,
      }));
      userMock.getUser.mockResolvedValue(axiosNormalResponse({
        message: "ok",
        user: { id: 'user1', role: UserRoleEnum.CosmoAdmin, displayName: 'user1 name' },
      }));

      authMock.login.mockRejectedValue({ response: { data: { message: 'login error' }, status: 401 } as any } as any);
      await act(async () => {
        await expect(result.current.login('user1', 'password1')).rejects.toMatchSnapshot();
      });
      await waitFor(async () => { expect(result.current).not.toBeNull(); });
      expect(result.current.loginUser).toMatchSnapshot();
    });


    it('getUser error', async () => {
      const { result } = await renderUseLogin();

      authMock.login.mockResolvedValue(axiosNormalResponse({
        id: 'user1',
        expireAt: 'xxxx',
        requirePasswordUpdate: true,
      }));

      userMock.getUser.mockRejectedValue({ response: { data: { message: 'getUser error' }, status: 401 } as any } as any);

      await act(async () => {
        await expect(result.current.login('user1', 'password1')).rejects.toMatchSnapshot();
      });
      expect(result.current.loginUser).toMatchSnapshot();
    });

  });


  describe('refreshUserInfo', () => {

    describe('not login', () => {

      it('ok', async () => {
        const { result } = await renderUseLogin();
        await expect(result.current.refreshUserInfo()).resolves.toMatchSnapshot();
        expect(result.current.loginUser).toMatchSnapshot();
      });

    });

    describe('logined', () => {

      beforeEach(async () => {
        authMock.verify.mockResolvedValueOnce(axiosNormalResponse({
          id: 'user1',
          expireAt: 'xxxx',
          requirePasswordUpdate: false,
        }));
        userMock.getUser.mockResolvedValue(axiosNormalResponse({
          message: "ok",
          user: { id: 'user1', role: UserRoleEnum.CosmoAdmin, displayName: 'user1 name' },
        }));
      });

      it('ok', async () => {
        const { result } = await renderUseLogin();
        await act(async () => {
          await expect(result.current.refreshUserInfo()).resolves.toMatchSnapshot();
        });
        expect(result.current.loginUser).toMatchSnapshot();
      });

    });
  });


  describe('logout', () => {

    beforeEach(async () => {
      authMock.verify.mockResolvedValueOnce(axiosNormalResponse({
        id: 'user1',
        expireAt: 'xxxx',
        requirePasswordUpdate: false,
      }));
      userMock.getUser.mockResolvedValue(axiosNormalResponse({
        message: "ok",
        user: { id: 'user1', role: UserRoleEnum.CosmoAdmin, displayName: 'user1 name' },
      }));
    });

    it('ok', async () => {
      const { result } = await renderUseLogin();
      await act(async () => {
        await expect(result.current.logout()).resolves.toMatchSnapshot();
      });
      await waitFor(async () => { expect(result.current.loginUser).toBeUndefined(); });
      expect(authMock.logout.mock.calls).toMatchSnapshot('logout calls');
    });

    it('error', async () => {
      const { result } = await renderUseLogin();
      authMock.logout.mockRejectedValue({ response: { data: { message: 'logout error' }, status: 500 } as any } as any);
      await act(async () => {
        await expect(result.current.logout()).rejects.toMatchSnapshot();
      });
      await waitFor(async () => { expect(result.current.loginUser).toBeUndefined(); });
      expect(authMock.logout.mock.calls).toMatchSnapshot('logout calls');
    });

  });


  describe('updatePassword', () => {

    beforeEach(async () => {
      authMock.verify.mockResolvedValueOnce(axiosNormalResponse({
        id: 'user1',
        expireAt: 'xxxx',
        requirePasswordUpdate: false,
      }));
      userMock.getUser.mockResolvedValue(axiosNormalResponse({
        message: "ok",
        user: { id: 'user1', role: UserRoleEnum.CosmoAdmin, displayName: 'user1 name' },
      }));
    });

    it('ok', async () => {
      const { result } = await renderUseLogin();
      userMock.putUserPassword.mockResolvedValue(axiosNormalResponse({
        message: "ok",
        user: { id: 'user1', role: 'cosmo-admin', displayName: 'user1 name' },
      }));
      await expect(result.current.updataPassword('oldpw', 'newpw')).resolves.toMatchSnapshot();
      expect(userMock.putUserPassword.mock.calls).toMatchSnapshot('putUserPassword calls');
    });

    it('error', async () => {
      const { result } = await renderUseLogin();
      userMock.putUserPassword.mockRejectedValue({ message: 'updataPassword error' } as any);
      await expect(result.current.updataPassword('oldpw', 'newpw')).rejects.toMatchSnapshot();
      expect(userMock.putUserPassword.mock.calls).toMatchSnapshot('putUserPassword calls');
    });

  });

});
