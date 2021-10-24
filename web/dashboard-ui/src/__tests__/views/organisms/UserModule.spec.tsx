import { act, renderHook, RenderResult } from '@testing-library/react-hooks';
import { AxiosError, AxiosResponse } from 'axios';
import { useSnackbar } from "notistack";
import { User, UserApiFactory, UserResponse } from "../../../api/dashboard/v1alpha1";
import { useProgress } from '../../../components/ProgressProvider';
import { UserContext, useUserModule } from '../../../views/organisms/UserModule';

//--------------------------------------------------
// mock definition
//--------------------------------------------------
jest.mock('notistack');
jest.mock('../../../components/LoginProvider');
jest.mock('.../../../api/dashboard/v1alpha1/api');
jest.mock('../../../components/ProgressProvider');
jest.mock('react-router-dom', () => ({
  useHistory: () => ({
    push: jest.fn(),
  }),
}));

type MockedMemberFunction<T extends (...args: any) => any> = {
  [P in keyof ReturnType<T>]: jest.MockedFunction<ReturnType<T>[P]>;
};

const RestUserMock = UserApiFactory as jest.MockedFunction<typeof UserApiFactory>;
const userMock: MockedMemberFunction<typeof UserApiFactory> = {
  postUser: jest.fn(),
  putUserRole: jest.fn(),
  putUserPassword: jest.fn(),
  getUser: jest.fn(),
  getUsers: jest.fn(),
  deleteUser: jest.fn(),
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
const user1: User = { id: 'user1', role: 'cosmo-admin', displayName: 'user1 name' };
const user2: User = { id: 'user2', displayName: 'user2 name' };
const user3: User = { id: 'user3', displayName: 'user3 name' };

function axiosNormalResponse<T>(data: T): AxiosResponse<T> {
  return { data: data, status: 200, statusText: 'ok', headers: {}, config: {}, request: {} }
}


//-----------------------------------------------
// test
//-----------------------------------------------
describe('useUserModule', () => {

  let result: RenderResult<ReturnType<typeof useUserModule>>;

  beforeEach(async () => {
    useSnackbarMock.mockReturnValue(snackbarMock);
    useProgressMock.mockReturnValue(progressMock);
    RestUserMock.mockReturnValue(userMock);
    result = renderHook(() => useUserModule(), {
      wrapper: ({ children }) => (<UserContext.Provider>{children}</UserContext.Provider>),
    }).result;
  });

  afterEach(() => {
    jest.restoreAllMocks();
  });

  describe('useUserModule getUsers', () => {

    it('normal', async () => {
      userMock.getUsers.mockResolvedValue(axiosNormalResponse({ message: "ok", items: [user2, user1, user3] }));
      await act(async () => { result.current.getUsers() });
      expect(result.current.users).toStrictEqual([user1, user2, user3]);
      expect(snackbarMock.enqueueSnackbar.mock.calls.length).toEqual(0);
    });


    it('normal', async () => {
      userMock.getUsers.mockResolvedValue(axiosNormalResponse({ message: "ok", items: undefined }));
      await act(async () => { result.current.getUsers() });
      expect(result.current.users).toStrictEqual([]);
      expect(snackbarMock.enqueueSnackbar.mock.calls.length).toEqual(0);
    });

    it('error', async () => {
      const err: AxiosError<UserResponse> = {
        response: { data: { message: undefined }, status: 402 } as any,
      } as any
      userMock.getUsers.mockRejectedValue(err);
      await expect(result.current.getUsers()).rejects.toStrictEqual(err);
      expect(snackbarMock.enqueueSnackbar.mock.calls.length).toEqual(0);
    });

  });


  describe('useUserModule createUser', () => {
    beforeEach(async () => {
      userMock.getUsers.mockResolvedValue(axiosNormalResponse({ message: "ok", items: [user1, user2, user3] }));
      await act(async () => { result.current.getUsers() });
    });

    it('nomal', async () => {
      userMock.postUser.mockResolvedValue(axiosNormalResponse({ message: "ok", user: user2 }));
      await act(async () => { result.current.createUser('user2', 'user2 name') });
      expect(result.current.users).toStrictEqual([user1, user2, user3]);
      expect(snackbarMock.enqueueSnackbar.mock.calls[0][0]).toEqual('ok');
    });

    it('error', async () => {
      const err: AxiosError<UserResponse> = {
        response: { data: { message: 'data.message' }, status: 401 } as any,
      } as any
      userMock.postUser.mockRejectedValue(err);
      await expect(result.current.createUser('user2', 'user2 name')).rejects.toStrictEqual(err);
      expect(snackbarMock.enqueueSnackbar.mock.calls[0][0]).toEqual('data.message');
    });
  });

  describe('useUserModule updateRole', () => {
    beforeEach(async () => {
      userMock.getUsers.mockResolvedValue(axiosNormalResponse({ message: "ok", items: [user1, user2, user3] }));
      await act(async () => { result.current.getUsers() });
    });

    it('nomal', async () => {
      userMock.putUserRole.mockResolvedValue(axiosNormalResponse({ message: "ok", user: user2 }));
      await act(async () => { result.current.updateRole('user2', 'user2 name') });
      expect(result.current.users).toStrictEqual([user1, user2, user3]);
      expect(snackbarMock.enqueueSnackbar.mock.calls[0][0]).toEqual('ok');
    });

    it('error', async () => {
      const err: AxiosError<UserResponse> = {
        response: { data: { message: 'data.message' }, status: 402 } as any,
      } as any
      userMock.putUserRole.mockRejectedValue(err);
      await expect(result.current.updateRole('user2', 'user2 name')).rejects.toStrictEqual(err);
      expect(snackbarMock.enqueueSnackbar.mock.calls[0][0]).toEqual('data.message');
    });
  });

  describe('useUserModule deleteUser', () => {
    beforeEach(async () => {
      userMock.getUsers.mockResolvedValue(axiosNormalResponse({ message: "ok", items: [user1, user2, user3] }));
      await act(async () => { result.current.getUsers() });
    });

    it('nomal', async () => {
      userMock.deleteUser.mockResolvedValue(axiosNormalResponse({ message: "ok", user: user2 }));
      await act(async () => { result.current.deleteUser('user2') });
      expect(result.current.users).toStrictEqual([user1, user3]);
      expect(snackbarMock.enqueueSnackbar.mock.calls[0][0]).toEqual('ok');
    });

    it('error', async () => {
      const err: AxiosError<UserResponse> = {
        response: { data: { message: 'data.message' }, status: 402 } as any,
      } as any
      userMock.deleteUser.mockRejectedValue(err);
      await expect(result.current.deleteUser('user2')).rejects.toStrictEqual(err);
      expect(snackbarMock.enqueueSnackbar.mock.calls[0][0]).toEqual('data.message');
    });
  });

});
