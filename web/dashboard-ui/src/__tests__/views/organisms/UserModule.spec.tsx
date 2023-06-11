import { Code, ConnectError } from '@bufbuild/connect';
import '@testing-library/jest-dom';
import { act, cleanup, renderHook } from '@testing-library/react';
import { useSnackbar } from "notistack";
import React from 'react';
import { afterEach, beforeEach, describe, expect, it, MockedFunction, vi } from "vitest";
import { useLogin } from '../../../components/LoginProvider';
import { useProgress } from '../../../components/ProgressProvider';
import { Template } from '../../../proto/gen/dashboard/v1alpha1/template_pb';
import { GetUserAddonTemplatesResponse } from '../../../proto/gen/dashboard/v1alpha1/template_service_pb';
import { User } from '../../../proto/gen/dashboard/v1alpha1/user_pb';
import { CreateUserResponse, DeleteUserResponse, GetUsersResponse, UpdateUserDisplayNameResponse, UpdateUserRoleResponse } from '../../../proto/gen/dashboard/v1alpha1/user_service_pb';
import { useTemplateService, useUserService } from "../../../services/DashboardServices";
import { UserContext, useTemplates, useUserModule } from '../../../views/organisms/UserModule';

//--------------------------------------------------
// mock definition
//--------------------------------------------------
vi.mock('notistack');
vi.mock('../../../components/LoginProvider');
vi.mock('../../../services/DashboardServices');
vi.mock('../../../components/ProgressProvider');
vi.mock('react-router-dom', () => ({
  useNavigate: () => vi.fn(),
}));

type MockedMemberFunction<T extends (...args: any) => any> = {
  [P in keyof ReturnType<T>]: MockedFunction<ReturnType<T>[P]>;
};

const useUserServiceMock = useUserService as MockedFunction<typeof useUserService>;
const userMock: MockedMemberFunction<typeof useUserService> = {
  getUser: vi.fn(),
  getUsers: vi.fn(),
  deleteUser: vi.fn(),
  createUser: vi.fn(),
  updateUserDisplayName: vi.fn(),
  updateUserPassword: vi.fn(),
  updateUserRole: vi.fn(),
}
const useLoginMock = useLogin as MockedFunction<typeof useLogin>;
const loginMock: MockedMemberFunction<typeof useLogin> = {
  loginUser: {} as any,
  verifyLogin: vi.fn(),
  login: vi.fn(),
  logout: vi.fn(),
  updataPassword: vi.fn(),
  refreshUserInfo: vi.fn(),
  clearLoginUser: vi.fn(),
};
const useTemplateServiceMock = useTemplateService as MockedFunction<typeof useTemplateService>;
const templateMock: MockedMemberFunction<typeof useTemplateService> = {
  getUserAddonTemplates: vi.fn(),
  getWorkspaceTemplates: vi.fn(),
}

const useProgressMock = useProgress as MockedFunction<typeof useProgress>;
const progressMock: MockedMemberFunction<typeof useProgress> = {
  setMask: vi.fn(),
  releaseMask: vi.fn(),
}
const useSnackbarMock = useSnackbar as MockedFunction<typeof useSnackbar>;
const snackbarMock: MockedMemberFunction<typeof useSnackbar> = {
  enqueueSnackbar: vi.fn(),
  closeSnackbar: vi.fn(),
}

//--------------------------------------------------
// mock data definition
//--------------------------------------------------
const user1 = new User({ name: 'user1', roles: ['cosmoAdmin'], displayName: 'user1 name' });
const user2 = new User({ name: 'user2', displayName: 'user2 name' });
const user3 = new User({ name: 'user3', displayName: 'user3 name' });

//-----------------------------------------------
// test
//-----------------------------------------------
describe('useUserModule', () => {

  beforeEach(async () => {
    useSnackbarMock.mockReturnValue(snackbarMock);
    useProgressMock.mockReturnValue(progressMock);
    useUserServiceMock.mockReturnValue(userMock);
    useLoginMock.mockReturnValue(loginMock);
  });

  afterEach(() => {
    expect(snackbarMock.enqueueSnackbar.mock.calls).toMatchSnapshot();
    vi.restoreAllMocks();
    cleanup();
  });

  async function renderUseUserModule() {
    return renderHook(() => useUserModule(), {
      wrapper: ({ children }) => (<UserContext.Provider>{children}</UserContext.Provider>),
    });
  }

  describe('useUserModule getUsers', () => {

    it('normal', async () => {
      const { result } = await renderUseUserModule();
      userMock.getUsers.mockResolvedValue(new GetUsersResponse({ items: [user2, user1, user3] }));
      await act(async () => { result.current.getUsers() });
      expect(result.current.users).toMatchSnapshot();
    });

    it('normal2', async () => {
      const { result } = await renderUseUserModule();
      userMock.getUsers.mockResolvedValue(new GetUsersResponse({ items: undefined as any }));
      await act(async () => { result.current.getUsers() });
      expect(result.current.users).toMatchSnapshot();
    });

    it('normal3', async () => {
      const { result } = await renderUseUserModule();
      userMock.getUsers.mockResolvedValue(new GetUsersResponse({ items: [] }));
      await act(async () => { result.current.getUsers() });
      expect(result.current.users).toMatchSnapshot();
    });

    it('error', async () => {
      const { result } = await renderUseUserModule();
      //userMock.getUsers.mockRejectedValue(new Error('[mock] getUsers error'));
      userMock.getUsers.mockRejectedValue(new ConnectError('[mock] getUsers error', Code.Unauthenticated));
      await expect(result.current.getUsers()).rejects.toMatchSnapshot();
    });

  });


  describe('useUserModule createUser', () => {
    beforeEach(async () => {
      userMock.getUsers.mockResolvedValue(new GetUsersResponse({ items: [user1, user2, user3] }));
    });

    it('nomal', async () => {
      const { result } = await renderUseUserModule();
      await act(async () => { result.current.getUsers() });
      userMock.createUser.mockResolvedValue(new CreateUserResponse({ message: "ok", user: user2 }));
      await act(async () => { result.current.createUser('user2', 'user2 name') });
      expect(result.current.users).toMatchSnapshot();
    });

    it('error', async () => {
      const { result } = await renderUseUserModule();
      await act(async () => { result.current.getUsers() });
      userMock.createUser.mockRejectedValue(new Error('[mock] createUser error'));
      await expect(result.current.createUser('user2', 'user2 name')).rejects.toMatchSnapshot();
    });
  });

  describe('useUserModule updateUserName', () => {

    it('nomal before getUsers', async () => {
      const { result } = await renderUseUserModule();
      const user2x = { ...user2, displayName: 'displayNameChange' }
      userMock.updateUserDisplayName.mockResolvedValue(new UpdateUserDisplayNameResponse({ message: "ok", user: user2x }));
      await act(async () => { result.current.updateName('user2', 'displayNameChange') });
      expect(result.current.users).toMatchSnapshot();
    });

    it('nomal after getUsers', async () => {
      const { result } = await renderUseUserModule();
      userMock.getUsers.mockResolvedValue(new GetUsersResponse({ message: "ok", items: [user1, user2, user3] }));
      await act(async () => { result.current.getUsers() });
      const user2x = { ...user2, displayName: 'displayNameChange' }
      userMock.updateUserDisplayName.mockResolvedValue(new UpdateUserDisplayNameResponse({ message: "ok", user: user2x }));
      await act(async () => { result.current.updateName('user2', 'displayNameChange') });
      expect(result.current.users).toMatchSnapshot();
    });

    it('nomal after getUsers. return value nothing', async () => {
      const { result } = await renderUseUserModule();
      userMock.getUsers.mockResolvedValue(new GetUsersResponse({ message: "ok", items: [user1, user2, user3] }));
      await act(async () => { result.current.getUsers() });
      userMock.updateUserDisplayName.mockResolvedValue(new UpdateUserDisplayNameResponse({ message: "ok", user: undefined as any }));
      await act(async () => { result.current.updateName('user2', 'displayNameChange') });
      expect(result.current.users).toMatchSnapshot();
    });

    it('error', async () => {
      const { result } = await renderUseUserModule();
      userMock.updateUserDisplayName.mockRejectedValue(new Error('[mock] putUserName error'));
      await expect(result.current.updateName('user2', 'displayNameChange')).rejects.toMatchSnapshot();
    });
  });

  describe('useUserModule updateRole', () => {

    it('nomal before getUsers', async () => {
      const { result } = await renderUseUserModule();
      userMock.updateUserRole.mockResolvedValue(new UpdateUserRoleResponse({ message: "ok", user: user2 }));
      await act(async () => { result.current.updateRole('user2', ['user2 name']) });
      expect(result.current.users).toMatchSnapshot();
    });

    it('nomal after getUsers', async () => {
      const { result } = await renderUseUserModule();
      userMock.getUsers.mockResolvedValue(new GetUsersResponse({ message: "ok", items: [user1, user2, user3] }));
      await act(async () => { result.current.getUsers() });

      const user2x = { ...user2, roles: ['Role2'] }
      userMock.updateUserRole.mockResolvedValue(new UpdateUserRoleResponse({ message: "ok", user: user2x }));

      await act(async () => { result.current.updateRole('user2', ['Role2']) });
      expect(result.current.users).toMatchSnapshot();
    });

    it('nomal after getUsers. return value nothing', async () => {
      const { result } = await renderUseUserModule();
      userMock.getUsers.mockResolvedValue(new GetUsersResponse({ message: "ok", items: [user1, user2, user3] }));
      await act(async () => { result.current.getUsers() });

      userMock.updateUserRole.mockResolvedValue(new UpdateUserRoleResponse({ message: "ok", user: undefined as any }));

      await act(async () => { result.current.updateRole('user2', ['Role2']) });
      expect(result.current.users).toMatchSnapshot();
    });

    it('error', async () => {
      const { result } = await renderUseUserModule();
      userMock.updateUserRole.mockRejectedValue(new Error('[mock] updateUserRole error'));
      await expect(result.current.updateRole('user2', ['user2 name'])).rejects.toMatchSnapshot();
    });
  });

  describe('useUserModule deleteUser', () => {
    beforeEach(async () => {
      userMock.getUsers.mockResolvedValue(new GetUsersResponse({ message: "ok", items: [user1, user2, user3] }));
    });

    it('nomal', async () => {
      const { result } = await renderUseUserModule();
      await act(async () => { result.current.getUsers() });
      userMock.deleteUser.mockResolvedValue(new DeleteUserResponse({ message: "ok", user: user2 }));
      await act(async () => { result.current.deleteUser('user2') });
      expect(result.current.users).toMatchSnapshot();
    });

    it('error', async () => {
      const { result } = await renderUseUserModule();
      await act(async () => { result.current.getUsers() });
      userMock.deleteUser.mockRejectedValue(new Error('[mock] deleteUser error'));
      await expect(result.current.deleteUser('user2')).rejects.toMatchSnapshot();
    });
  });

});

describe('useTemplates', () => {

  beforeEach(async () => {
    useSnackbarMock.mockReturnValue(snackbarMock);
    useProgressMock.mockReturnValue(progressMock);
    useTemplateServiceMock.mockReturnValue(templateMock);
    useLoginMock.mockReturnValue(loginMock);
  });

  afterEach(() => {
    vi.restoreAllMocks();
    cleanup();
  });

  async function renderUseTemplates() {
    const utils = renderHook(() => useTemplates(), {
      wrapper: ({ children }) => (<UserContext.Provider>{children}</UserContext.Provider>),
    });
    return utils;
  }

  describe('useTemplates getUserAddonTemplates', () => {

    const tmpl1 = new Template({ name: 'tmpl1' });
    const tmpl2 = new Template({
      name: 'tmpl2',
      description: "hoge",
      requiredVars: [{ varName: 'var1', defaultValue: 'var1Value' }, { varName: 'var2' }],
      isDefaultUserAddon: true,
    });
    const tmpl3 = new Template({ name: 'tmpl3' });

    it('normal', async () => {
      const { result } = await renderUseTemplates();
      templateMock.getUserAddonTemplates.mockResolvedValue(new GetUserAddonTemplatesResponse({
        message: "ok", items: [tmpl1, tmpl3, tmpl2]
      }));
      await act(async () => { result.current.getUserAddonTemplates() });
      expect(result.current.templates).toMatchSnapshot();
    });

    it('normal template empty', async () => {
      const { result } = await renderUseTemplates();
      templateMock.getUserAddonTemplates.mockResolvedValue(new GetUserAddonTemplatesResponse({ items: [] }));
      await act(async () => { result.current.getUserAddonTemplates() });
      expect(result.current.templates).toMatchSnapshot();
    });

    it('error', async () => {
      const { result } = await renderUseTemplates();
      templateMock.getUserAddonTemplates.mockRejectedValue(new Error('[mock] getUser error'));
      await expect(result.current.getUserAddonTemplates()).rejects.toMatchSnapshot();
      expect(result.current.templates).toMatchSnapshot();
    });

  });

});
