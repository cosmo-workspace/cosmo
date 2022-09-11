import { act, renderHook } from '@testing-library/react';
import { AxiosError, AxiosResponse } from 'axios';
import { useSnackbar } from "notistack";
import React from 'react';
import { CreateWorkspaceResponse, DeleteWorkspaceResponse, GetUserResponse, GetWorkspaceResponse, ListTemplatesResponse, NetworkRule, PatchWorkspaceResponse, RemoveNetworkRuleResponse, Template, TemplateApiFactory, UpsertNetworkRuleResponse, User, UserApiFactory, UserRoleEnum, Workspace, WorkspaceApiFactory } from "../../../api/dashboard/v1alpha1";
import { useLogin } from "../../../components/LoginProvider";
import { useProgress } from '../../../components/ProgressProvider';
import { computeStatus, useNetworkRule, useTemplates, useWorkspaceModule, useWorkspaceUsersModule, WorkspaceContext, WorkspaceUsersContext } from '../../../views/organisms/WorkspaceModule';
//--------------------------------------------------
// mock definition
//--------------------------------------------------
jest.mock('notistack');
jest.mock('../../../components/LoginProvider');
jest.mock('../../../api/dashboard/v1alpha1/api');
jest.mock('../../../components/ProgressProvider');
jest.mock('react-router-dom', () => ({
  useNavigate: () => jest.fn(),
}));

//type ReturnMemberType<T extends (...args: any) => any, K extends keyof ReturnType<T>> = ReturnType<T>[K];
//type MockedFunc<T extends (...args: any) => any, K extends keyof ReturnType<T>> = jest.MockedFunction<ReturnType<T>[K]>;
type MockedMemberFunction<T extends (...args: any) => any> = {
  [P in keyof ReturnType<T>]: jest.MockedFunction<ReturnType<T>[P]>;
};

const RestWorkspaceMock = WorkspaceApiFactory as jest.MockedFunction<typeof WorkspaceApiFactory>;
const wsMock: MockedMemberFunction<typeof WorkspaceApiFactory> = {
  getWorkspace: jest.fn(),
  getWorkspaces: jest.fn(),
  postWorkspace: jest.fn(),
  deleteWorkspace: jest.fn(),
  patchWorkspace: jest.fn(),
  deleteNetworkRule: jest.fn(),
  putNetworkRule: jest.fn(),
}
const restTemplateMock = TemplateApiFactory as jest.MockedFunction<typeof TemplateApiFactory>;
const templateMock: MockedMemberFunction<typeof TemplateApiFactory> = {
  getWorkspaceTemplates: jest.fn(),
  getUserAddonTemplates: jest.fn(),
}
const useProgressMock = useProgress as jest.MockedFunction<typeof useProgress>;
const progressMock: MockedMemberFunction<typeof useProgress> = {
  setMask: jest.fn(),
  releaseMask: jest.fn(),
}
const useSnackbarMock = useSnackbar as jest.MockedFunction<typeof useSnackbar>;
const snackbarMock: MockedMemberFunction<typeof useSnackbar> = {
  enqueueSnackbar: jest.fn(),
  closeSnackbar: jest.fn(),
}

//--------------------------------------------------
// mock data definition
//--------------------------------------------------
const newWorkspace = (name: string, user: User, tmpl: Template, replicas = 1, phase = 'Running'): Workspace => {
  return ({
    name: name, ownerID: user.id,
    spec: { template: tmpl.name, replicas: replicas, vars: { xxx: 'XXXX', yyy: 'YYYY' }, additionalNetwork: [] },
    status: { phase: phase, mainUrl: "", urlBase: 'urlbasexxxx' }
  });
}
const wsStat = (ws: Workspace, replicas: number, phase: string): Workspace => {
  return { ...ws, spec: { ...ws.spec!, replicas }, status: { ...ws.status, phase } }
}

const user1: User = { id: 'user1', role: UserRoleEnum.CosmoAdmin, displayName: 'user1 name' };
const user2: User = { id: 'user2', displayName: 'user2 name' };
const user3: User = { id: 'user3', displayName: 'user2 name' };
const tmpl1: Template = { name: 'tmpl1' };
const tmpl2: Template = { name: 'tmpl2', requiredVars: [{ varName: 'var1' }, { varName: 'var2' }] };
const tmpl3: Template = { name: 'tmpl3' };
const ws11 = newWorkspace('ws11', user1, tmpl1);
const ws12 = newWorkspace('ws12', user1, tmpl2);
const ws13 = newWorkspace('ws13', user1, tmpl1);
const ws14 = newWorkspace('ws14', user1, tmpl1); //add 
const ws15 = newWorkspace('ws15', user1, tmpl1);

function axiosNormalResponse<T>(data: T): AxiosResponse<T> {
  return { data: data, status: 200, statusText: 'ok', headers: {}, config: {}, request: {} }
}


//-----------------------------------------------
// test
//-----------------------------------------------
describe('computeStatus', () => {
  const wsStarting: Workspace = { name: '', ownerID: '', spec: { template: '', replicas: 1 }, status: { phase: 'Stopped' } };
  const wsPending: Workspace = { name: '', ownerID: '', spec: { template: '', replicas: -1 }, status: { phase: 'Pending' } };
  const wsStoping: Workspace = { name: '', ownerID: '', spec: { template: '', replicas: 0 }, status: { phase: 'Running' } };
  const wsStopped: Workspace = { name: '', ownerID: '', spec: { template: '', replicas: 0 }, status: { phase: 'Stopped' } };
  const wsRunning: Workspace = { name: '', ownerID: '', spec: { template: '', replicas: 1 }, status: { phase: 'Running' } };

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
    RestWorkspaceMock.mockReturnValue(wsMock);
    wsMock.getWorkspaces.mockResolvedValue(axiosNormalResponse({ message: "", items: [ws11, ws13, ws12] }));
  });

  afterEach(() => {
    jest.restoreAllMocks();
    jest.useRealTimers();
  });

  function renderUseWorkspaceModule() {
    return renderHook(() => useWorkspaceModule(), {
      wrapper: ({ children }) => (<WorkspaceContext.Provider>{children}</WorkspaceContext.Provider>),
    });
  }

  describe('useWorkspace getWorkspaces', () => {

    it('normal', async () => {
      const { result } = renderUseWorkspaceModule();
      await act(async () => { result.current.getWorkspaces(user1.id) });
      expect(result.current.workspaces).toStrictEqual([ws11, ws12, ws13]);
    });

    it('normal empty', async () => {
      const { result } = renderUseWorkspaceModule();
      wsMock.getWorkspaces.mockResolvedValue(axiosNormalResponse({ message: "", items: [] }));
      await act(async () => { result.current.getWorkspaces(user1.id) });
      expect(result.current.workspaces).toStrictEqual([]);
    });


    it('get error', async () => {
      const { result } = renderUseWorkspaceModule();
      const err: AxiosError<GetWorkspaceResponse> = {
        response: { data: { message: 'data.message' }, status: 401 } as any,
      } as any
      wsMock.getWorkspaces.mockRejectedValue(err);
      await expect(result.current.getWorkspaces(user1.id)).rejects.toStrictEqual(err);
      expect(result.current.workspaces).toStrictEqual([]);
      expect(snackbarMock.enqueueSnackbar.mock.calls[0][0]).toEqual('data.message');
    });
  });


  describe('useWorkspace refreshWorkspaces', () => {

    it('normal', async () => {
      const { result } = renderUseWorkspaceModule();
      await act(async () => { result.current.refreshWorkspaces(user1.id) });
      expect(result.current.workspaces).toStrictEqual([ws11, ws12, ws13]);
    });

  });

  describe('useWorkspace refreshWorkspace', () => {

    it('normal creating', async () => {

      const wsCreateing = wsStat(ws12, 1, 'Creating');
      const wsStarting = wsStat(ws12, 1, 'NotRunning');
      const wsPending = wsStat(ws12, 1, 'Pending');
      const wsRunning = wsStat(ws12, 1, 'Running');

      jest.useFakeTimers();
      jest.spyOn(global, 'setTimeout');

      const { result } = renderUseWorkspaceModule();

      // getWorkspaces then setWorkspaces
      wsMock.getWorkspaces.mockResolvedValueOnce(axiosNormalResponse({ message: "", items: [ws11, wsCreateing, ws13] }));
      await act(async () => { result.current.getWorkspaces(user1.id) });
      expect(result.current.workspaces).toStrictEqual([ws11, wsCreateing, ws13]);

      // refReshWorkspace
      await act(async () => { result.current.refreshWorkspace(wsCreateing) });

      wsMock.getWorkspace.mockRejectedValueOnce(new Error());
      await act(async () => { jest.runAllTimers(); });
      expect(result.current.workspaces).toStrictEqual([ws11, wsCreateing, ws13]);
      //　　refReshWorkspace starting
      wsMock.getWorkspace.mockResolvedValueOnce(axiosNormalResponse({ message: "", workspace: wsStarting }));
      await act(async () => { jest.runAllTimers(); });
      expect(result.current.workspaces).toStrictEqual([ws11, wsStarting, ws13]);
      //　　refReshWorkspace starting
      wsMock.getWorkspace.mockResolvedValueOnce(axiosNormalResponse({ message: "", workspace: wsStarting }));
      await act(async () => { jest.runAllTimers(); });
      expect(result.current.workspaces).toStrictEqual([ws11, wsStarting, ws13]);
      //　　refReshWorkspace pending
      wsMock.getWorkspace.mockResolvedValueOnce(axiosNormalResponse({ message: "", workspace: wsPending }));
      await act(async () => { jest.runAllTimers(); });
      expect(result.current.workspaces).toStrictEqual([ws11, wsPending, ws13]);
      //　　refReshWorkspace running
      wsMock.getWorkspace.mockResolvedValueOnce(axiosNormalResponse({ message: "", workspace: wsRunning }));
      await act(async () => { jest.runAllTimers(); });
      expect(result.current.workspaces).toStrictEqual([ws11, wsRunning, ws13]);

      // expect(setTimeout).toHaveBeenCalledTimes(10);
      // await act(async () => { jest.runAllTimers(); });
      // expect(setTimeout).toHaveBeenCalledTimes(10);
    });

    it('normal stopping', async () => {

      const wsStoping = wsStat(ws12, 0, 'Running');
      const wsStopped = wsStat(ws12, 0, 'NotRunning');

      const { result } = renderUseWorkspaceModule();

      // getWorkspaces then setWorkspaces
      wsMock.getWorkspaces.mockResolvedValueOnce(axiosNormalResponse({ message: "", items: [ws11, wsStoping, ws13] }));
      await act(async () => { result.current.getWorkspaces(user1.id) });
      expect(result.current.workspaces).toStrictEqual([ws11, wsStoping, ws13]);

      jest.useFakeTimers();
      jest.spyOn(global, 'setTimeout');

      // refReshWorkspace
      await act(async () => { result.current.refreshWorkspace(wsStoping) });

      wsMock.getWorkspace.mockResolvedValueOnce(axiosNormalResponse({ message: "", workspace: wsStoping }));
      await act(async () => { jest.runOnlyPendingTimers(); });
      expect(result.current.workspaces).toStrictEqual([ws11, wsStoping, ws13]);
      //　　refReshWorkspace stoping
      wsMock.getWorkspace.mockResolvedValueOnce(axiosNormalResponse({ message: "", workspace: wsStoping }));
      await act(async () => { jest.runOnlyPendingTimers(); });
      expect(result.current.workspaces).toStrictEqual([ws11, wsStoping, ws13]);
      //　　refReshWorkspace stopped
      wsMock.getWorkspace.mockResolvedValueOnce(axiosNormalResponse({ message: "", workspace: wsStopped }));
      await act(async () => { jest.runOnlyPendingTimers(); });
      expect(result.current.workspaces).toStrictEqual([ws11, wsStopped, ws13]);
    });

    it('normal network modify', async () => {


      const wsRunning1 = wsStat(ws12, 1, 'Running');
      wsRunning1.spec!.additionalNetwork = [
        { portName: 'portname1', portNumber: 3000, url: 'url1', public: false },
        { portName: 'portname2', portNumber: 3000, public: false },
      ];
      const wsRunning2 = wsStat(ws12, 1, 'Running');
      wsRunning1.spec!.additionalNetwork = [
        { portName: 'portname1', portNumber: 3000, url: 'url1', public: false },
        { portName: 'portname2', portNumber: 3000, url: 'url2', public: false },
      ];

      jest.useFakeTimers();
      jest.spyOn(global, 'setTimeout');

      const { result } = renderUseWorkspaceModule();

      // getWorkspaces then setWorkspaces
      wsMock.getWorkspaces.mockResolvedValueOnce(axiosNormalResponse({ message: "", items: [ws11, wsRunning1, ws13] }));
      await act(async () => { result.current.getWorkspaces(user1.id) });
      expect(result.current.workspaces).toStrictEqual([ws11, wsRunning1, ws13]);

      // refReshWorkspace
      await act(async () => { result.current.refreshWorkspace(wsRunning1) });

      wsMock.getWorkspace.mockResolvedValueOnce(axiosNormalResponse({ message: "", workspace: wsRunning1 }));
      await act(async () => { jest.runAllTimers(); });
      expect(result.current.workspaces).toStrictEqual([ws11, wsRunning1, ws13]);
      //　　refReshWorkspace starting
      wsMock.getWorkspace.mockResolvedValueOnce(axiosNormalResponse({ message: "", workspace: wsRunning2 }));
      await act(async () => { jest.runAllTimers(); });
      expect(result.current.workspaces).toStrictEqual([ws11, wsRunning1, ws13]);

      const timerCount = (setTimeout as any).mock.calls.length;
      await act(async () => { jest.runOnlyPendingTimers(); });
      expect(setTimeout).toHaveBeenCalledTimes(timerCount);
    });

    it('normal timeout', async () => {

      const wsStoping = wsStat(ws12, 0, 'Running');

      const { result } = renderUseWorkspaceModule();

      // getWorkspaces then setWorkspaces
      wsMock.getWorkspaces.mockResolvedValueOnce(axiosNormalResponse({ message: "", items: [ws11, wsStoping, ws13] }));
      await act(async () => { result.current.getWorkspaces(user1.id) });
      expect(result.current.workspaces).toStrictEqual([ws11, wsStoping, ws13]);

      jest.useFakeTimers();
      jest.spyOn(global, 'setTimeout');

      // refReshWorkspace
      await act(async () => { result.current.refreshWorkspace(wsStoping) });

      wsMock.getWorkspace.mockResolvedValue(axiosNormalResponse({ message: "", workspace: wsStoping }));

      let i;
      for (i = 120000; 0 < i; i -= 1000) {
        await act(async () => { jest.runOnlyPendingTimers(); });
        expect(result.current.workspaces).toStrictEqual([ws11, wsStoping, ws13]);
      }

      const timerCount = (setTimeout as any).mock.calls.length;
      await act(async () => { jest.runOnlyPendingTimers(); });
      expect(setTimeout).toHaveBeenCalledTimes(timerCount);
    });

  });

  it('error', async () => {

    const wsStoping = wsStat(ws12, 0, 'Running');

    const err: AxiosError<GetWorkspaceResponse> = {
      response: { data: { message: 'data.message' }, status: 401 } as any,
    } as any

    jest.useFakeTimers();
    jest.spyOn(global, 'setTimeout');

    const { result } = renderUseWorkspaceModule();
    await act(async () => { result.current.refreshWorkspace(wsStoping) });

    wsMock.getWorkspace.mockRejectedValueOnce(err);
    await act(async () => { jest.runOnlyPendingTimers(); });
    expect(setTimeout).toHaveBeenCalledTimes(1);
  });


  describe('useWorkspace createWorkspace', () => {

    it('normal', async () => {
      const { result } = renderUseWorkspaceModule();
      wsMock.getWorkspaces.mockResolvedValueOnce(axiosNormalResponse({ message: "", items: [ws11, ws12, ws13, ws15] }));
      await act(async () => { result.current.getWorkspaces(user1.id) });
      expect(result.current.workspaces).toStrictEqual([ws11, ws12, ws13, ws15]);

      wsMock.postWorkspace.mockResolvedValueOnce(axiosNormalResponse({ message: "ok", workspace: ws14 }));
      await act(async () => { result.current.createWorkspace(ws14.ownerID!, ws14.name, ws14.spec!.template, {}) });
      expect(result.current.workspaces).toStrictEqual([ws11, ws12, ws13, wsStat(ws14, 1, 'Creating'), ws15]);
      expect(snackbarMock.enqueueSnackbar.mock.calls[0][0]).toEqual('ok');
    });

    it('get error', async () => {
      const { result } = renderUseWorkspaceModule();
      const err: AxiosError<CreateWorkspaceResponse> = {
        response: { data: { message: 'data.message' }, status: 401 } as any,
      } as any
      wsMock.postWorkspace.mockRejectedValue(err);
      await expect(result.current.createWorkspace(ws14.ownerID!, ws14.name, ws14.spec!.template, {}))
        .rejects.toStrictEqual(err);
      expect(result.current.workspaces).toStrictEqual([]);
      expect(snackbarMock.enqueueSnackbar.mock.calls[0][0]).toEqual('data.message');
    });
  });


  describe('useWorkspace runWorkspace', () => {

    it('normal', async () => {
      const { result } = renderUseWorkspaceModule();
      wsMock.getWorkspaces.mockResolvedValueOnce(axiosNormalResponse({ message: "", items: [ws11, ws12, ws13] }));
      await act(async () => { result.current.getWorkspaces(user1.id) });
      expect(result.current.workspaces).toStrictEqual([ws11, ws12, ws13]);

      wsMock.patchWorkspace.mockResolvedValueOnce(axiosNormalResponse({ message: "ok", workspace: ws11 }));
      wsMock.getWorkspace.mockResolvedValueOnce(axiosNormalResponse({ message: "ok", workspace: ws11 }));
      await act(async () => { result.current.runWorkspace(ws11) });
      expect(result.current.workspaces).toStrictEqual([ws11, ws12, ws13]);
      expect(snackbarMock.enqueueSnackbar.mock.calls[0][0]).toEqual('Successfully run workspace');
    });

    it('error', async () => {
      const { result } = renderUseWorkspaceModule();
      const err: AxiosError<PatchWorkspaceResponse> = {
        response: { data: { message: 'data.message' }, status: 401 } as any,
      } as any
      wsMock.patchWorkspace.mockRejectedValue(err);
      await expect(result.current.runWorkspace(ws11)).rejects.toStrictEqual(err);
      expect(result.current.workspaces).toStrictEqual([]);
      expect(snackbarMock.enqueueSnackbar.mock.calls[0][0]).toEqual('data.message');
    });
  });


  describe('useWorkspace stopWorkspace', () => {

    it('normal', async () => {
      const { result } = renderUseWorkspaceModule();
      wsMock.getWorkspaces.mockResolvedValueOnce(axiosNormalResponse({ message: "", items: [ws11, ws12, ws13] }));
      await act(async () => { result.current.getWorkspaces(user1.id) });
      expect(result.current.workspaces).toStrictEqual([ws11, ws12, ws13]);

      wsMock.patchWorkspace.mockResolvedValueOnce(axiosNormalResponse({ message: "ok", workspace: ws11 }));
      wsMock.getWorkspace.mockResolvedValueOnce(axiosNormalResponse({ message: "ok", workspace: ws11 }));
      await act(async () => { result.current.stopWorkspace(ws11) });
      expect(result.current.workspaces).toStrictEqual([ws11, ws12, ws13]);
      expect(snackbarMock.enqueueSnackbar.mock.calls[0][0]).toEqual('Successfully stopped workspace');
    });

    it('error', async () => {
      const { result } = renderUseWorkspaceModule();
      const err: AxiosError<PatchWorkspaceResponse> = {
        response: { data: { message: 'data.message' }, status: 401 } as any,
      } as any
      wsMock.patchWorkspace.mockRejectedValue(err);
      await expect(result.current.stopWorkspace(ws11)).rejects.toStrictEqual(err);
      expect(result.current.workspaces).toStrictEqual([]);
      expect(snackbarMock.enqueueSnackbar.mock.calls[0][0]).toEqual('data.message');
    });
  });

  describe('useWorkspace deleteWorkspace', () => {

    it('normal', async () => {
      const { result } = renderUseWorkspaceModule();
      wsMock.getWorkspaces.mockResolvedValueOnce(axiosNormalResponse({ message: "", items: [ws11, ws12, ws13] }));
      await act(async () => { result.current.getWorkspaces(user1.id) });
      expect(result.current.workspaces).toStrictEqual([ws11, ws12, ws13]);

      wsMock.deleteWorkspace.mockResolvedValueOnce(axiosNormalResponse({ message: "ok", workspace: ws11 }));
      wsMock.getWorkspace.mockResolvedValueOnce(axiosNormalResponse({ message: "ok", workspace: ws11 }));
      await act(async () => { result.current.deleteWorkspace(ws11) });
      expect(result.current.workspaces).toStrictEqual([ws11, ws12, ws13]);
      expect(snackbarMock.enqueueSnackbar.mock.calls[0][0]).toEqual('ok');
    });

    it('error', async () => {
      const { result } = renderUseWorkspaceModule();
      const err: AxiosError<DeleteWorkspaceResponse> = {
        response: { data: { message: 'data.message' }, status: 401 } as any,
      } as any
      wsMock.deleteWorkspace.mockRejectedValue(err);
      await expect(result.current.deleteWorkspace(ws11)).rejects.toStrictEqual(err);
      expect(result.current.workspaces).toStrictEqual([]);
      expect(snackbarMock.enqueueSnackbar.mock.calls[0][0]).toEqual('data.message');
    });
  });

});


describe('useTemplates', () => {

  beforeEach(async () => {
    useSnackbarMock.mockReturnValue(snackbarMock);
    useProgressMock.mockReturnValue(progressMock);
    restTemplateMock.mockReturnValue(templateMock);
  });

  afterEach(() => { jest.restoreAllMocks(); });

  it('normal', async () => {
    const { result } = renderHook(() => useTemplates());
    templateMock.getWorkspaceTemplates.mockResolvedValue(axiosNormalResponse({ message: "", items: [tmpl1, tmpl3, tmpl2] }));
    await act(async () => { result.current.getTemplates() });
    expect(result.current.templates).toStrictEqual([tmpl1, tmpl2, tmpl3]);
  });

  it('normal empty', async () => {
    const { result } = renderHook(() => useTemplates());
    templateMock.getWorkspaceTemplates.mockResolvedValue(axiosNormalResponse({ message: "", items: [] }));
    await act(async () => { result.current.getTemplates() });
    expect(result.current.templates).toStrictEqual([]);
  });

  it('get error', async () => {
    const { result } = renderHook(() => useTemplates());
    const err: AxiosError<ListTemplatesResponse> = {
      response: { data: { message: 'data.message' }, status: 401 } as any,
    } as any
    templateMock.getWorkspaceTemplates.mockRejectedValue(err);
    await expect(result.current.getTemplates()).rejects.toStrictEqual(err);
    expect(result.current.templates).toStrictEqual([]);
    expect(snackbarMock.enqueueSnackbar.mock.calls[0][0]).toEqual('data.message');
  });
});


describe('useNetworkRule', () => {

  const nw111: NetworkRule = { portName: 'nw1', httpPath: '/path1', portNumber: 1111, public: false };

  beforeEach(async () => {
    useSnackbarMock.mockReturnValue(snackbarMock);
    useProgressMock.mockReturnValue(progressMock);
    RestWorkspaceMock.mockReturnValue(wsMock);
  });

  afterEach(() => { jest.restoreAllMocks(); });

  function renderUseNetworkRule() {
    return renderHook(() => useNetworkRule(), {
      wrapper: ({ children }) => (<WorkspaceContext.Provider>{children}</WorkspaceContext.Provider>),
    });
  }

  describe('useNetworkRule upsertNetwork', () => {

    it('normal', async () => {
      const { result } = renderUseNetworkRule();
      wsMock.putNetworkRule.mockResolvedValue(axiosNormalResponse({ message: "ok", networkRule: nw111 }));
      await act(async () => { result.current.upsertNetwork(ws11, nw111) });
      expect(snackbarMock.enqueueSnackbar.mock.calls[0][0]).toEqual('ok');
    });

    it('error', async () => {
      const { result } = renderUseNetworkRule();
      const err: AxiosError<UpsertNetworkRuleResponse> = {
        response: { data: { message: 'data.message' }, status: 401 } as any,
      } as any
      wsMock.putNetworkRule.mockRejectedValue(err);
      await expect(result.current.upsertNetwork(ws11, nw111)).rejects.toStrictEqual(err);
      expect(snackbarMock.enqueueSnackbar.mock.calls[0][0]).toEqual('data.message');
    });

  });

  describe('useNetworkRule removeNetwork', () => {

    it('normal', async () => {
      const { result } = renderUseNetworkRule();
      wsMock.deleteNetworkRule.mockResolvedValue(axiosNormalResponse({ message: "ok", networkRule: nw111 }));
      await act(async () => { result.current.removeNetwork(ws11, nw111.portName) });
      expect(snackbarMock.enqueueSnackbar.mock.calls[0][0]).toEqual('ok');
    });

    it('error', async () => {
      const { result } = renderUseNetworkRule();
      const err: AxiosError<RemoveNetworkRuleResponse> = {
        response: { data: { message: 'data.message' }, status: 401 } as any,
      } as any
      wsMock.deleteNetworkRule.mockRejectedValue(err);
      await expect(result.current.removeNetwork(ws11, nw111.portName)).rejects.toStrictEqual(err);
      expect(snackbarMock.enqueueSnackbar.mock.calls[0][0]).toEqual('data.message');
    });

  });

});


describe('useWorkspaceUsers', () => {

  const useLoginMock = useLogin as jest.MockedFunction<typeof useLogin>;
  const loginMock: MockedMemberFunction<typeof useLogin> = {
    loginUser: {} as any,
    verifyLogin: jest.fn(),
    login: jest.fn(),
    logout: jest.fn(),
    updataPassword: jest.fn(),
    refreshUserInfo: jest.fn(),
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

  beforeEach(async () => {
    useSnackbarMock.mockReturnValue(snackbarMock);
    useProgressMock.mockReturnValue(progressMock);
  });

  afterEach(() => { jest.restoreAllMocks(); });

  describe('useWorkspaceUsers not login', () => {

    it('normal', async () => {
      useLoginMock.mockReturnValue({
        loginUser: undefined,
        verifyLogin: jest.fn(),
        login: jest.fn(),
        logout: jest.fn(),
        updataPassword: jest.fn(),
        refreshUserInfo: jest.fn(),
      });
      const result = renderHook(() => useWorkspaceUsersModule(), {
        wrapper: ({ children }) => (<WorkspaceUsersContext.Provider>{children}</WorkspaceUsersContext.Provider>),
      }).result;

      expect(result.current.user).toStrictEqual({ id: "" });
      expect(result.current.users).toStrictEqual([{ id: "" }]);
      expect(snackbarMock.enqueueSnackbar.mock.calls.length).toEqual(0);
    });
  });


  describe('useWorkspaceUsers getUsers', () => {

    beforeEach(async () => {
      useLoginMock.mockReturnValue(loginMock);
      RestUserMock.mockReturnValue(userMock);
    });

    function renderUseWorkspaceUsersModule() {
      return renderHook(() => useWorkspaceUsersModule(), {
        wrapper: ({ children }) => (<WorkspaceUsersContext.Provider>{children}</WorkspaceUsersContext.Provider>),
      });
    }

    it('normal', async () => {
      const { result } = renderUseWorkspaceUsersModule();
      userMock.getUsers.mockResolvedValue(axiosNormalResponse({ message: "ok", items: [user2, user1, user3] }));
      await act(async () => { result.current.getUsers() });
      expect(result.current.users).toStrictEqual([user1, user2, user3]);
      expect(snackbarMock.enqueueSnackbar.mock.calls.length).toEqual(0);
    });

    it('normal empty', async () => {
      const { result } = renderUseWorkspaceUsersModule();
      userMock.getUsers.mockResolvedValue(axiosNormalResponse({ message: "ok", items: [] }));
      await act(async () => { result.current.getUsers() });
      expect(result.current.users).toStrictEqual([]);
      expect(snackbarMock.enqueueSnackbar.mock.calls.length).toEqual(0);
    });

    it('normal no change', async () => {
      const { result } = renderUseWorkspaceUsersModule();
      userMock.getUsers.mockResolvedValue(axiosNormalResponse({ message: "ok", items: [user2, user1, user3] }));
      await act(async () => { result.current.getUsers() });
      await act(async () => { result.current.getUsers() });
      expect(result.current.users).toStrictEqual([user1, user2, user3]);
      expect(snackbarMock.enqueueSnackbar.mock.calls.length).toEqual(0);
    });

    it('error', async () => {
      const { result } = renderUseWorkspaceUsersModule();
      const err: AxiosError<GetUserResponse> = {
        response: { data: { message: undefined }, status: 402 } as any,
      } as any
      userMock.getUsers.mockRejectedValue(err);
      await expect(result.current.getUsers()).rejects.toStrictEqual(err);
      expect(snackbarMock.enqueueSnackbar.mock.calls.length).toEqual(0);
    });

  });

});