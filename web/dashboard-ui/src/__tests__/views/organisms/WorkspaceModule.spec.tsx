import { Code, ConnectError } from '@bufbuild/connect';
import { protoInt64 } from '@bufbuild/protobuf';
import { act, cleanup, renderHook } from '@testing-library/react';
import { useSnackbar } from "notistack";
import React from 'react';
import { afterEach, beforeEach, describe, expect, it, MockedFunction, vi } from "vitest";
import { useLogin } from "../../../components/LoginProvider";
import { useProgress } from '../../../components/ProgressProvider';
import { Template } from '../../../proto/gen/dashboard/v1alpha1/template_pb';
import { GetWorkspaceTemplatesResponse } from '../../../proto/gen/dashboard/v1alpha1/template_service_pb';
import { User } from '../../../proto/gen/dashboard/v1alpha1/user_pb';
import { GetUsersResponse } from '../../../proto/gen/dashboard/v1alpha1/user_service_pb';
import { NetworkRule, Workspace } from '../../../proto/gen/dashboard/v1alpha1/workspace_pb';
import { CreateWorkspaceResponse, DeleteNetworkRuleResponse, DeleteWorkspaceResponse, GetWorkspaceResponse, GetWorkspacesResponse, UpdateWorkspaceResponse, UpsertNetworkRuleResponse } from '../../../proto/gen/dashboard/v1alpha1/workspace_service_pb';
import { useTemplateService, useUserService, useWorkspaceService } from '../../../services/DashboardServices';
import { computeStatus, useNetworkRule, useTemplates, useWorkspaceModule, useWorkspaceUsersModule, WorkspaceContext, WorkspaceUsersContext } from '../../../views/organisms/WorkspaceModule';

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

//type ReturnMemberType<T extends (...args: any) => any, K extends keyof ReturnType<T>> = ReturnType<T>[K];
//type MockedFunc<T extends (...args: any) => any, K extends keyof ReturnType<T>> = MockedFunction<ReturnType<T>[K]>;
type MockedMemberFunction<T extends (...args: any) => any> = {
  [P in keyof ReturnType<T>]: MockedFunction<ReturnType<T>[P]>;
};

const useWorkspaceServiceMock = useWorkspaceService as MockedFunction<typeof useWorkspaceService>;
const wsMock: MockedMemberFunction<typeof useWorkspaceService> = {
  getWorkspace: vi.fn(),
  getWorkspaces: vi.fn(),
  deleteWorkspace: vi.fn(),
  deleteNetworkRule: vi.fn(),
  createWorkspace: vi.fn(),
  updateWorkspace: vi.fn(),
  upsertNetworkRule: vi.fn()
}
const useTemplateServiceMock = useTemplateService as MockedFunction<typeof useTemplateService>;
const templateMock: MockedMemberFunction<typeof useTemplateService> = {
  getWorkspaceTemplates: vi.fn(),
  getUserAddonTemplates: vi.fn(),
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
function newWorkspace(name: string, user: User, tmpl: Template, replicas = 1, phase = 'Running'): Workspace {
  return (new Workspace({
    name: name,
    ownerName: user.name,
    spec: { template: tmpl.name, replicas: protoInt64.parse(1), vars: { xxx: 'XXXX', yyy: 'YYYY' }, additionalNetwork: [] },
    status: { phase: phase, mainUrl: "", urlBase: 'urlbasexxxx' }
  }));
}
function wsStat(ws: Workspace, replicas: number, phase: string): Workspace {
  return new Workspace({ ...ws, spec: { ...ws.spec!, replicas: protoInt64.parse(replicas) }, status: { ...ws.status, phase } });
}

const user1: User = new User({ name: 'user1', roles: ["cosmo-admin"], displayName: 'user1 name' });
const user2: User = new User({ name: 'user2', displayName: 'user2 name' });
const user3: User = new User({ name: 'user3', displayName: 'user2 name' });
const tmpl1: Template = new Template({ name: 'tmpl1' });
const tmpl2: Template = new Template({ name: 'tmpl2', requiredVars: [{ varName: 'var1' }, { varName: 'var2' }] });
const tmpl3: Template = new Template({ name: 'tmpl3' });
const ws11 = newWorkspace('ws11', user1, tmpl1);
const ws12 = newWorkspace('ws12', user1, tmpl2);
const ws13 = newWorkspace('ws13', user1, tmpl1);
const ws14 = newWorkspace('ws14', user1, tmpl1); //add 
const ws15 = newWorkspace('ws15', user1, tmpl1);

//-----------------------------------------------
// test
//-----------------------------------------------
describe('computeStatus', () => {
  const wsStarting = new Workspace({ name: '', ownerName: '', spec: { template: '', replicas: protoInt64.parse(1) }, status: { phase: 'Stopped' } });
  const wsPending = new Workspace({ name: '', ownerName: '', spec: { template: '', replicas: protoInt64.parse(-1) }, status: { phase: 'Pending' } });
  const wsStoping = new Workspace({ name: '', ownerName: '', spec: { template: '', replicas: protoInt64.parse(0) }, status: { phase: 'Running' } });
  const wsStopped = new Workspace({ name: '', ownerName: '', spec: { template: '', replicas: protoInt64.parse(0) }, status: { phase: 'Stopped' } });
  const wsRunning = new Workspace({ name: '', ownerName: '', spec: { template: '', replicas: protoInt64.parse(1) }, status: { phase: 'Running' } });

  it('Stopping', () => { expect(computeStatus(wsStoping)).toEqual('Stopping') });
  it('Stopped', () => { expect(computeStatus(wsStopped)).toEqual('Stopped') });
  it('Starting', () => { expect(computeStatus(wsStarting)).toEqual('Starting'); });
  it('Running', () => { expect(computeStatus(wsRunning)).toEqual('Running') });
  it('Other', async () => { expect(computeStatus(wsPending)).toEqual('Pending') });
});


describe('useWorkspace', () => {

  beforeEach(async () => {
    useSnackbarMock.mockReturnValue(snackbarMock);
    useProgressMock.mockReturnValue(progressMock);
    useWorkspaceServiceMock.mockReturnValue(wsMock);
    useLoginMock.mockReturnValue(loginMock);
    wsMock.getWorkspaces.mockResolvedValue(new GetWorkspacesResponse({ message: "", items: [ws11, ws13, ws12] }));
  });

  afterEach(() => {
    expect(snackbarMock.enqueueSnackbar.mock.calls).toMatchSnapshot();
    vi.restoreAllMocks();
    vi.useRealTimers();
    cleanup();
  });

  function renderUseWorkspaceModule() {
    return renderHook(() => useWorkspaceModule(), {
      wrapper: ({ children }) => (<WorkspaceContext.Provider>{children}</WorkspaceContext.Provider>),
    });
  }

  describe('useWorkspace getWorkspaces', () => {

    it('normal', async () => {
      const { result } = renderUseWorkspaceModule();
      await act(async () => { result.current.getWorkspaces(user1.name) });
      expect(result.current.workspaces).toMatchSnapshot();
    });

    it('normal empty', async () => {
      const { result } = renderUseWorkspaceModule();
      wsMock.getWorkspaces.mockResolvedValue(new GetWorkspacesResponse({ message: "", items: [] }));
      await act(async () => { result.current.getWorkspaces(user1.name) });
      expect(result.current.workspaces).toMatchSnapshot();
    });


    it('get error', async () => {
      const { result } = renderUseWorkspaceModule();
      // wsMock.getWorkspaces.mockRejectedValue(new Error('[mock] getWorkspaces error'));
      wsMock.getWorkspaces.mockRejectedValue(new ConnectError('[mock] getWorkspaces error', Code.Unauthenticated));
      await expect(result.current.getWorkspaces(user1.name)).rejects.toMatchSnapshot();
      expect(result.current.workspaces).toMatchSnapshot();
    });
  });


  describe('useWorkspace refreshWorkspaces', () => {

    it('normal', async () => {
      const { result } = renderUseWorkspaceModule();
      await act(async () => { result.current.refreshWorkspaces(user1.name) });
      expect(result.current.workspaces).toMatchSnapshot();
    });

  });

  describe('useWorkspace refreshWorkspace', () => {

    it('normal creating', async () => {

      const wsCreateing = wsStat(ws12, 1, 'Creating');
      const wsStarting = wsStat(ws12, 1, 'NotRunning');
      const wsPending = wsStat(ws12, 1, 'Pending');
      const wsRunning = wsStat(ws12, 1, 'Running');

      vi.useFakeTimers();
      vi.spyOn(global, 'setTimeout');

      const { result } = renderUseWorkspaceModule();

      // getWorkspaces then setWorkspaces
      wsMock.getWorkspaces.mockResolvedValueOnce(new GetWorkspacesResponse({ message: "", items: [ws11, wsCreateing, ws13] }));
      await act(async () => { result.current.getWorkspaces(user1.name) });
      expect(result.current.workspaces).toMatchSnapshot();

      // refReshWorkspace
      await act(async () => { result.current.refreshWorkspace(wsCreateing) });

      wsMock.getWorkspace.mockRejectedValueOnce(new Error());
      await act(async () => { vi.runAllTimers(); });
      expect(result.current.workspaces).toMatchSnapshot();
      //　　refReshWorkspace starting
      wsMock.getWorkspace.mockResolvedValueOnce(new GetWorkspaceResponse({ workspace: wsStarting }));
      await act(async () => { vi.runAllTimers(); });
      expect(result.current.workspaces).toMatchSnapshot();
      //　　refReshWorkspace starting
      wsMock.getWorkspace.mockResolvedValueOnce(new GetWorkspaceResponse({ workspace: wsStarting }));
      await act(async () => { vi.runAllTimers(); });
      expect(result.current.workspaces).toMatchSnapshot();
      //　　refReshWorkspace pending
      wsMock.getWorkspace.mockResolvedValueOnce(new GetWorkspaceResponse({ workspace: wsPending }));
      await act(async () => { vi.runAllTimers(); });
      expect(result.current.workspaces).toMatchSnapshot();
      //　　refReshWorkspace running
      wsMock.getWorkspace.mockResolvedValueOnce(new GetWorkspaceResponse({ workspace: wsRunning }));
      await act(async () => { vi.runAllTimers(); });
      expect(result.current.workspaces).toMatchSnapshot();

      // expect(setTimeout).toHaveBeenCalledTimes(10);
      // await act(async () => { jest.runAllTimers(); });
      // expect(setTimeout).toHaveBeenCalledTimes(10);
    });

    it('normal stopping', async () => {

      const wsStoping = wsStat(ws12, 0, 'Running');
      const wsStopped = wsStat(ws12, 0, 'NotRunning');

      const { result } = renderUseWorkspaceModule();

      // getWorkspaces then setWorkspaces
      wsMock.getWorkspaces.mockResolvedValueOnce(new GetWorkspacesResponse({ message: "", items: [ws11, wsStoping, ws13] }));
      await act(async () => { result.current.getWorkspaces(user1.name) });
      expect(result.current.workspaces).toMatchSnapshot();

      vi.useFakeTimers();
      vi.spyOn(global, 'setTimeout');

      // refReshWorkspace
      await act(async () => { result.current.refreshWorkspace(wsStoping) });

      wsMock.getWorkspace.mockResolvedValueOnce(new GetWorkspaceResponse({ workspace: wsStoping }));
      await act(async () => { vi.runOnlyPendingTimers(); });
      expect(result.current.workspaces).toMatchSnapshot();
      //　　refReshWorkspace stoping
      wsMock.getWorkspace.mockResolvedValueOnce(new GetWorkspaceResponse({ workspace: wsStoping }));
      await act(async () => { vi.runOnlyPendingTimers(); });
      expect(result.current.workspaces).toMatchSnapshot();
      //　　refReshWorkspace stopped
      wsMock.getWorkspace.mockResolvedValueOnce(new GetWorkspaceResponse({ workspace: wsStopped }));
      await act(async () => { vi.runOnlyPendingTimers(); });
      expect(result.current.workspaces).toMatchSnapshot();
    });

    it('normal network modify', async () => {


      const wsRunning1 = wsStat(ws12, 1, 'Running');
      wsRunning1.spec!.additionalNetwork = [
        new NetworkRule({ name: 'portname1', portNumber: 3000, url: 'url1', public: false }),
        new NetworkRule({ name: 'portname2', portNumber: 3000, public: false }),
      ];
      const wsRunning2 = wsStat(ws12, 1, 'Running');
      wsRunning1.spec!.additionalNetwork = [
        new NetworkRule({ name: 'portname1', portNumber: 3000, url: 'url1', public: false }),
        new NetworkRule({ name: 'portname2', portNumber: 3000, url: 'url2', public: false }),
      ];

      vi.useFakeTimers();
      vi.spyOn(global, 'setTimeout');

      const { result } = renderUseWorkspaceModule();

      // getWorkspaces then setWorkspaces
      wsMock.getWorkspaces.mockResolvedValueOnce(new GetWorkspacesResponse({ message: "", items: [ws11, wsRunning1, ws13] }));
      await act(async () => { result.current.getWorkspaces(user1.name) });
      expect(result.current.workspaces).toMatchSnapshot();

      // refReshWorkspace
      await act(async () => { result.current.refreshWorkspace(wsRunning1) });

      wsMock.getWorkspace.mockResolvedValueOnce(new GetWorkspaceResponse({ workspace: wsRunning1 }));
      await act(async () => { vi.runAllTimers(); });
      expect(result.current.workspaces).toMatchSnapshot();
      //　　refReshWorkspace starting
      wsMock.getWorkspace.mockResolvedValueOnce(new GetWorkspaceResponse({ workspace: wsRunning2 }));
      await act(async () => { vi.runAllTimers(); });
      expect(result.current.workspaces).toMatchSnapshot();

      const timerCount = (setTimeout as any).mock.calls.length;
      await act(async () => { vi.runOnlyPendingTimers(); });
      expect(setTimeout).toHaveBeenCalledTimes(timerCount);
    });

    it('normal timeout', async () => {

      const wsStoping = wsStat(ws12, 0, 'Running');

      const { result } = renderUseWorkspaceModule();

      // getWorkspaces then setWorkspaces
      wsMock.getWorkspaces.mockResolvedValueOnce(new GetWorkspacesResponse({ message: "", items: [ws11, wsStoping, ws13] }));
      await act(async () => { result.current.getWorkspaces(user1.name) });
      expect(result.current.workspaces).toMatchSnapshot();

      vi.useFakeTimers();
      vi.spyOn(global, 'setTimeout');

      // refReshWorkspace
      await act(async () => { result.current.refreshWorkspace(wsStoping) });

      wsMock.getWorkspace.mockResolvedValue(new GetWorkspaceResponse({ workspace: wsStoping }));

      let i;
      for (i = 120000; 0 < i; i -= 1000) {
        await act(async () => { vi.runOnlyPendingTimers(); });
        expect(result.current.workspaces).toMatchSnapshot();
      }

      const timerCount = (setTimeout as any).mock.calls.length;
      await act(async () => { vi.runOnlyPendingTimers(); });
      expect(setTimeout).toHaveBeenCalledTimes(timerCount);
    });

  });

  it('error', async () => {

    const wsStoping = wsStat(ws12, 0, 'Running');

    vi.useFakeTimers();
    vi.spyOn(global, 'setTimeout');

    const { result } = renderUseWorkspaceModule();
    await act(async () => { result.current.refreshWorkspace(wsStoping) });

    wsMock.getWorkspace.mockRejectedValueOnce(new Error('[mock] getWorkspace error'));
    await act(async () => { vi.runOnlyPendingTimers(); });
    expect(setTimeout).toHaveBeenCalledTimes(1);
  });


  describe('useWorkspace createWorkspace', () => {

    it('normal', async () => {
      vi.useFakeTimers();
      const { result } = renderUseWorkspaceModule();
      wsMock.getWorkspaces.mockResolvedValueOnce(new GetWorkspacesResponse({ message: "", items: [ws11, ws12, ws13, ws15] }));
      await act(async () => { result.current.getWorkspaces(user1.name) });
      expect(result.current.workspaces).toMatchSnapshot();

      wsMock.createWorkspace.mockResolvedValueOnce(new CreateWorkspaceResponse({ message: "ok", workspace: ws14 }));
      await act(async () => { result.current.createWorkspace(ws14.ownerName, ws14.name, ws14.spec!.template, {}) });
      expect(result.current.workspaces).toMatchSnapshot();
    });

    it('get error', async () => {
      const { result } = renderUseWorkspaceModule();
      wsMock.createWorkspace.mockRejectedValue(new Error('[mock] createWorkspace error'));
      await expect(result.current.createWorkspace(ws14.ownerName, ws14.name, ws14.spec!.template, {}))
        .rejects.toMatchSnapshot();
      expect(result.current.workspaces).toMatchSnapshot();
    });
  });


  describe('useWorkspace runWorkspace', () => {

    it('normal', async () => {
      vi.useFakeTimers();
      const { result } = renderUseWorkspaceModule();
      wsMock.getWorkspaces.mockResolvedValueOnce(new GetWorkspacesResponse({ message: "", items: [ws11, ws12, ws13] }));
      await act(async () => { result.current.getWorkspaces(user1.name) });
      expect(result.current.workspaces).toMatchSnapshot();

      wsMock.updateWorkspace.mockResolvedValueOnce(new UpdateWorkspaceResponse({ message: "ok", workspace: ws11 }));
      wsMock.getWorkspace.mockResolvedValueOnce(new GetWorkspaceResponse({ workspace: ws11 }));
      await act(async () => { result.current.runWorkspace(ws11) });
      expect(result.current.workspaces).toMatchSnapshot();
    });

    it('error', async () => {
      const { result } = renderUseWorkspaceModule();
      wsMock.updateWorkspace.mockRejectedValue(new Error('[mock] updateWorkspace error'));
      await expect(result.current.runWorkspace(ws11)).rejects.toMatchSnapshot();
      expect(result.current.workspaces).toMatchSnapshot();
    });
  });


  describe('useWorkspace stopWorkspace', () => {

    it('normal', async () => {
      vi.useFakeTimers();
      const { result } = renderUseWorkspaceModule();
      wsMock.getWorkspaces.mockResolvedValueOnce(new GetWorkspacesResponse({ message: "", items: [ws11, ws12, ws13] }));
      await act(async () => { result.current.getWorkspaces(user1.name) });
      expect(result.current.workspaces).toMatchSnapshot();

      wsMock.updateWorkspace.mockResolvedValueOnce(new UpdateWorkspaceResponse({ message: "ok", workspace: ws11 }));
      wsMock.getWorkspace.mockResolvedValueOnce(new GetWorkspaceResponse({ workspace: ws11 }));
      await act(async () => { result.current.stopWorkspace(ws11) });
      expect(result.current.workspaces).toMatchSnapshot();
    });

    it('error', async () => {
      const { result } = renderUseWorkspaceModule();
      wsMock.updateWorkspace.mockRejectedValue(new Error('[mock] patchWorkspace error'));
      await expect(result.current.stopWorkspace(ws11)).rejects.toMatchSnapshot();
      expect(result.current.workspaces).toMatchSnapshot();
    });
  });

  describe('useWorkspace deleteWorkspace', () => {

    it('normal', async () => {
      vi.useFakeTimers();
      const { result } = renderUseWorkspaceModule();
      wsMock.getWorkspaces.mockResolvedValueOnce(new GetWorkspacesResponse({ message: "", items: [ws11, ws12, ws13] }));
      await act(async () => { result.current.getWorkspaces(user1.name) });
      expect(result.current.workspaces).toMatchSnapshot();

      wsMock.deleteWorkspace.mockResolvedValueOnce(new DeleteWorkspaceResponse({ message: "ok", workspace: ws11 }));
      wsMock.getWorkspace.mockResolvedValueOnce(new GetWorkspaceResponse({ workspace: ws11 }));
      await act(async () => { result.current.deleteWorkspace(ws11) });
      expect(result.current.workspaces).toMatchSnapshot();
    });

    it('error', async () => {
      const { result } = renderUseWorkspaceModule();
      wsMock.deleteWorkspace.mockRejectedValue(new Error('[mock] updateUserRole error'));
      await expect(result.current.deleteWorkspace(ws11)).rejects.toMatchSnapshot();
      expect(result.current.workspaces).toMatchSnapshot();
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
    expect(snackbarMock.enqueueSnackbar.mock.calls).toMatchSnapshot();
    vi.restoreAllMocks();
    cleanup();
  });

  it('normal', async () => {
    const { result } = renderHook(() => useTemplates());
    templateMock.getWorkspaceTemplates.mockResolvedValue(new GetWorkspaceTemplatesResponse({ message: "", items: [tmpl1, tmpl3, tmpl2] }));
    await act(async () => { result.current.getTemplates() });
    expect(result.current.templates).toMatchSnapshot();
  });

  it('normal empty', async () => {
    const { result } = renderHook(() => useTemplates());
    templateMock.getWorkspaceTemplates.mockResolvedValue(new GetWorkspaceTemplatesResponse({ message: "", items: [] }));
    await act(async () => { result.current.getTemplates() });
    expect(result.current.templates).toMatchSnapshot();
  });

  it('get error', async () => {
    const { result } = renderHook(() => useTemplates());
    templateMock.getWorkspaceTemplates.mockRejectedValue(new Error('[mock] getWorkspaceTemplates error'));
    await expect(result.current.getTemplates()).rejects.toMatchSnapshot();
    expect(result.current.templates).toMatchSnapshot();
  });
});


describe('useNetworkRule', () => {

  const nw111 = new NetworkRule({ name: 'nw1', httpPath: '/path1', portNumber: 1111, public: false });

  beforeEach(async () => {
    vi.useFakeTimers();
    useSnackbarMock.mockReturnValue(snackbarMock);
    useProgressMock.mockReturnValue(progressMock);
    useWorkspaceServiceMock.mockReturnValue(wsMock);
    useLoginMock.mockReturnValue(loginMock);
  });

  afterEach(() => {
    expect(snackbarMock.enqueueSnackbar.mock.calls).toMatchSnapshot();
    vi.restoreAllMocks();
    cleanup();
  });

  function renderUseNetworkRule() {
    return renderHook(() => useNetworkRule(), {
      wrapper: ({ children }) => (<WorkspaceContext.Provider>{children}</WorkspaceContext.Provider>),
    });
  }

  describe('useNetworkRule upsertNetwork', () => {

    it('normal', async () => {
      const { result } = renderUseNetworkRule();
      wsMock.upsertNetworkRule.mockResolvedValue(new UpsertNetworkRuleResponse({ message: "ok", networkRule: nw111 }));
      await act(async () => { result.current.upsertNetwork(ws11, nw111) });
    });

    it('error', async () => {
      const { result } = renderUseNetworkRule();
      wsMock.upsertNetworkRule.mockRejectedValue(new Error('[mock] upsertNetworkRule error'));
      await expect(result.current.upsertNetwork(ws11, nw111)).rejects.toMatchSnapshot();
    });

  });

  describe('useNetworkRule removeNetwork', () => {

    it('normal', async () => {
      const { result } = renderUseNetworkRule();
      wsMock.deleteNetworkRule.mockResolvedValue(new DeleteNetworkRuleResponse({ message: "ok", networkRule: nw111 }));
      await act(async () => { result.current.removeNetwork(ws11, nw111.name) });
    });

    it('error', async () => {
      const { result } = renderUseNetworkRule();
      wsMock.deleteNetworkRule.mockRejectedValue(new Error('[mock] deleteNetworkRule error'));
      await expect(result.current.removeNetwork(ws11, nw111.name)).rejects.toMatchSnapshot();
    });

  });

});


describe('useWorkspaceUsers', () => {

  const useLoginMock = useLogin as MockedFunction<typeof useLogin>;
  const loginMock: MockedMemberFunction<typeof useLogin> = {
    loginUser: {} as any,
    verifyLogin: vi.fn(),
    login: vi.fn(),
    logout: vi.fn(),
    updataPassword: vi.fn(),
    refreshUserInfo: vi.fn(),
    clearLoginUser: vi.fn(),
  }
  const useUserServiceMock = useUserService as MockedFunction<typeof useUserService>;
  const userMock: MockedMemberFunction<typeof useUserService> = {
    getUser: vi.fn(),
    getUsers: vi.fn(),
    deleteUser: vi.fn(),
    createUser: vi.fn(),
    updateUserDisplayName: vi.fn(),
    updateUserPassword: vi.fn(),
    updateUserRole: vi.fn()
  }

  beforeEach(async () => {
    useSnackbarMock.mockReturnValue(snackbarMock);
    useProgressMock.mockReturnValue(progressMock);
  });

  afterEach(() => {
    expect(snackbarMock.enqueueSnackbar.mock.calls).toMatchSnapshot();
    vi.restoreAllMocks();
    cleanup();
  });

  describe('useWorkspaceUsers not login', () => {

    it('normal', async () => {
      useLoginMock.mockReturnValue({
        loginUser: undefined as any,
        verifyLogin: vi.fn(),
        login: vi.fn(),
        logout: vi.fn(),
        updataPassword: vi.fn(),
        refreshUserInfo: vi.fn(),
        clearLoginUser: vi.fn(),
      });
      const result = renderHook(() => useWorkspaceUsersModule(), {
        wrapper: ({ children }) => (<WorkspaceUsersContext.Provider>{children}</WorkspaceUsersContext.Provider>),
      }).result;

      expect(result.current.user).toMatchSnapshot();
      expect(result.current.users).toMatchSnapshot();
    });
  });


  describe('useWorkspaceUsers getUsers', () => {

    beforeEach(async () => {
      useLoginMock.mockReturnValue(loginMock);
      useUserServiceMock.mockReturnValue(userMock);
    });

    function renderUseWorkspaceUsersModule() {
      return renderHook(() => useWorkspaceUsersModule(), {
        wrapper: ({ children }) => (<WorkspaceUsersContext.Provider>{children}</WorkspaceUsersContext.Provider>),
      });
    }

    it('normal', async () => {
      const { result } = renderUseWorkspaceUsersModule();
      userMock.getUsers.mockResolvedValue(new GetUsersResponse({ message: "ok", items: [user2, user1, user3] }));
      await act(async () => { result.current.getUsers() });
      expect(result.current.users).toMatchSnapshot();
    });

    it('normal empty', async () => {
      const { result } = renderUseWorkspaceUsersModule();
      userMock.getUsers.mockResolvedValue(new GetUsersResponse({ message: "ok", items: [] }));
      await act(async () => { result.current.getUsers() });
      expect(result.current.users).toMatchSnapshot();
    });

    it('normal no change', async () => {
      const { result } = renderUseWorkspaceUsersModule();
      userMock.getUsers.mockResolvedValue(new GetUsersResponse({ message: "ok", items: [user2, user1, user3] }));
      await act(async () => { result.current.getUsers() });
      await act(async () => { result.current.getUsers() });
      expect(result.current.users).toMatchSnapshot();
    });

    it('error', async () => {
      const { result } = renderUseWorkspaceUsersModule();
      userMock.getUsers.mockRejectedValue(new Error('[mock] getUsers error'));
      await expect(result.current.getUsers()).rejects.toMatchSnapshot();
    });

  });

});