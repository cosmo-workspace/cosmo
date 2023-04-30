import { Timestamp } from "@bufbuild/protobuf";
import { cleanup, renderHook, waitFor } from "@testing-library/react";
import { useSnackbar } from "notistack";
import React from 'react';
import { act } from "react-dom/test-utils";
import { afterEach, beforeEach, describe, expect, it, MockedFunction, vi } from "vitest";
import { LoginProvider, useLogin } from "../../components/LoginProvider";
import { useProgress } from "../../components/ProgressProvider";
import { LoginResponse, VerifyResponse } from "../../proto/gen/dashboard/v1alpha1/auth_service_pb";
import { User } from "../../proto/gen/dashboard/v1alpha1/user_pb";
import { GetUserResponse, UpdateUserPasswordResponse } from "../../proto/gen/dashboard/v1alpha1/user_service_pb";
import { useAuthService, useUserService } from "../../services/DashboardServices";

//--------------------------------------------------
// mock definition
//--------------------------------------------------

vi.mock('notistack');
vi.mock('../../services/DashboardServices');
vi.mock('../../components/ProgressProvider');
vi.mock('react-router-dom', () => ({
  //useHistory: vi.fn(),
}));


type MockedMemberFunction<T extends (...args: any) => any> = {
  [P in keyof ReturnType<T>]: MockedFunction<ReturnType<T>[P]>;
};

const useSnackbarMock = useSnackbar as MockedFunction<typeof useSnackbar>;
const snackbarMock: MockedMemberFunction<typeof useSnackbar> = {
  enqueueSnackbar: vi.fn(),
  closeSnackbar: vi.fn(),
}

const useProgressMock = useProgress as MockedFunction<typeof useProgress>;
const progressMock: MockedMemberFunction<typeof useProgress> = {
  setMask: vi.fn(),
  releaseMask: vi.fn(),
}

const authService = useAuthService as MockedFunction<typeof useAuthService>;
const authMock: MockedMemberFunction<typeof useAuthService> = {
  verify: vi.fn(),
  login: vi.fn(),
  logout: vi.fn(),
}

const userService = useUserService as MockedFunction<typeof useUserService>;
const userMock: MockedMemberFunction<typeof useUserService> = {
  getUser: vi.fn(),
  getUsers: vi.fn(),
  deleteUser: vi.fn(),
  createUser: vi.fn(),
  updateUserDisplayName: vi.fn(),
  updateUserPassword: vi.fn(),
  updateUserRole: vi.fn()
}

//-----------------------------------------------
// test
//-----------------------------------------------

describe('LoginProvider', () => {

  beforeEach(async () => {
    useSnackbarMock.mockReturnValue(snackbarMock);
    useProgressMock.mockReturnValue(progressMock);
    authService.mockReturnValue(authMock);
    userService.mockReturnValue(userMock);
  });

  afterEach(() => {
    // expect(snackbarMock.enqueueSnackbar.mock.calls).toMatchSnapshot('enqueueSnackbar calls');
    expect(snackbarMock.enqueueSnackbar.mock.calls).toMatchSnapshot();
    vi.restoreAllMocks();
    cleanup();
  });

  async function renderUseLogin() {
    const utils = renderHook(() => useLogin(), {
      wrapper: ({ children }) => (<LoginProvider>{children}</LoginProvider >),
    });
    await waitFor(async () => { expect(utils.result.current).not.toBeNull(); });
    return utils;
  }

  describe('verify', () => {

    it('✅ ok', async () => {
      authMock.verify.mockResolvedValueOnce(new VerifyResponse({
        userName: 'user1',
        expireAt: Timestamp.fromDate(new Date("2022/11/4")),
        requirePasswordUpdate: false,
      }));
      userMock.getUser.mockResolvedValue(new GetUserResponse({
        user: new User({ name: 'user1', roles: ["CosmoAdmin"], displayName: 'user1 name' }),
      }));

      const { result } = await renderUseLogin();
      expect(result.current.loginUser).toMatchSnapshot();
    });


    it('❌ ng / id is undefined', async () => {
      authMock.verify.mockResolvedValueOnce(new VerifyResponse({
        userName: undefined as any,
        expireAt: Timestamp.fromDate(new Date("2022/11/4")),
        requirePasswordUpdate: false,
      }));

      const { result } = await renderUseLogin();
      expect(result.current.loginUser).toMatchSnapshot();
    });


    it('❌ ng', async () => {

      authMock.verify.mockRejectedValue(new Error('[mock] verify error'));
      const { result } = await renderUseLogin();
      expect(result.current.loginUser).toMatchSnapshot();
    });

    it('❌ getUser error', async () => {

      authMock.verify.mockResolvedValueOnce(new VerifyResponse({
        userName: 'user1',
        expireAt: Timestamp.fromDate(new Date("2022/11/4")),
        requirePasswordUpdate: false,
      }));

      userMock.getUser.mockRejectedValue(new Error('[mock] getUser error'));

      const { result } = await renderUseLogin();
      expect(result.current.loginUser).toMatchSnapshot();
    });

  });


  describe('login', () => {

    it('✅ ok', async () => {
      authMock.login.mockResolvedValue(new LoginResponse({
        userName: 'user1',
        expireAt: Timestamp.fromDate(new Date("2022/11/4")),
        requirePasswordUpdate: true,
      }));
      userMock.getUser.mockResolvedValue(new GetUserResponse({
        user: new User({ name: 'user1', roles: ["CosmoAdmin"], displayName: 'user1 name' }),
      }));
      const { result } = await renderUseLogin();
      // await act(async () => {
      await expect(result.current.login('user1', 'password1')).resolves.toMatchSnapshot();
      // });
      await waitFor(async () => { expect(result.current.loginUser).not.toBeUndefined(); });
      expect(result.current.loginUser).toMatchSnapshot();
    });


    it('❌ ng', async () => {
      authMock.login.mockResolvedValue(new LoginResponse({
        userName: 'user1',
        expireAt: Timestamp.fromDate(new Date("2022/11/4")),
        requirePasswordUpdate: true,
      }));
      userMock.getUser.mockResolvedValue(new GetUserResponse({
        user: { name: 'user1', roles: ["CosmoAdmin"], displayName: 'user1 name' },
      }));
      authMock.login.mockRejectedValue(new Error('[mock] login error'));
      const { result } = await renderUseLogin();
      await act(async () => {
        await expect(result.current.login('user1', 'password1')).rejects.toMatchSnapshot();
      });
      await waitFor(async () => { expect(result.current).not.toBeNull(); });
      expect(result.current.loginUser).toMatchSnapshot();
    });


    it('❌ getUser error', async () => {
      const { result } = await renderUseLogin();

      authMock.login.mockResolvedValue(new LoginResponse({
        userName: 'user1',
        expireAt: Timestamp.fromDate(new Date("2022/11/4")),
        requirePasswordUpdate: true,
      }));

      userMock.getUser.mockRejectedValue(new Error('[mock] getUser error'));

      await act(async () => {
        await expect(result.current.login('user1', 'password1')).rejects.toMatchSnapshot();
      });
      expect(result.current.loginUser).toMatchSnapshot();
    });

  });


  describe('refreshUserInfo', () => {

    describe('not login', () => {

      it('✅ ok', async () => {
        const { result } = await renderUseLogin();
        await expect(result.current.refreshUserInfo()).resolves.toMatchSnapshot();
        expect(result.current.loginUser).toMatchSnapshot();
      });

    });

    describe('logined', () => {

      beforeEach(async () => {
        authMock.verify.mockResolvedValueOnce(new VerifyResponse({
          userName: 'user1',
          expireAt: Timestamp.fromDate(new Date("2022/11/4")),
          requirePasswordUpdate: false,
        }));
        userMock.getUser.mockResolvedValue(new GetUserResponse({
          user: { name: 'user1', roles: ["CosmoAdmin"], displayName: 'user1 name' },
        }));
      });

      it('✅ ok', async () => {
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
      authMock.verify.mockResolvedValueOnce(new VerifyResponse({
        userName: 'user1',
        expireAt: Timestamp.fromDate(new Date("2022/11/4")),
        requirePasswordUpdate: false,
      }));
      userMock.getUser.mockResolvedValue(new GetUserResponse({
        user: { name: 'user1', roles: ["CosmoAdmin"], displayName: 'user1 name' },
      }));
    });

    it('✅ ok', async () => {
      const { result } = await renderUseLogin();
      await act(async () => {
        await expect(result.current.logout()).resolves.toMatchSnapshot();
      });
      await waitFor(async () => { expect(result.current.loginUser).toBeUndefined(); });
      expect(authMock.logout.mock.calls).toMatchSnapshot('logout calls');
    });

    it('❌ error', async () => {
      const { result } = await renderUseLogin();
      authMock.logout.mockRejectedValue(new Error('[mock] logout error'));
      await act(async () => {
        await expect(result.current.logout()).rejects.toMatchSnapshot();
      });
      await waitFor(async () => { expect(result.current.loginUser).toBeUndefined(); });
      expect(authMock.logout.mock.calls).toMatchSnapshot('logout calls');
    });

  });


  describe('updatePassword', () => {

    beforeEach(async () => {
      authMock.verify.mockResolvedValueOnce(new VerifyResponse({
        userName: 'user1',
        expireAt: Timestamp.fromDate(new Date("2022/11/4")),
        requirePasswordUpdate: false,
      }));
      userMock.getUser.mockResolvedValue(new GetUserResponse({
        user: { name: 'user1', roles: ["CosmoAdmin"], displayName: 'user1 name' },
      }));
    });

    it('✅ ok', async () => {
      const { result } = await renderUseLogin();
      userMock.updateUserPassword.mockResolvedValue(new UpdateUserPasswordResponse({
        message: "ok",
      }));
      await expect(result.current.updataPassword('oldpw', 'newpw')).resolves.toMatchSnapshot();
      expect(userMock.updateUserPassword.mock.calls).toMatchSnapshot('putUserPassword calls');
    });

    it('❌ error', async () => {
      const { result } = await renderUseLogin();
      userMock.updateUserPassword.mockRejectedValue(new Error('[mock] updateUserPassword error'));
      await expect(result.current.updataPassword('oldpw', 'newpw')).rejects.toMatchSnapshot();
      expect(userMock.updateUserPassword.mock.calls).toMatchSnapshot('putUserPassword calls');
    });

  });

});
