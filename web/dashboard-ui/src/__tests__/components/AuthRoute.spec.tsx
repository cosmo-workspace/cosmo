import { Box } from '@mui/system';
import { render } from '@testing-library/react';
import React from 'react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { AuthRoute } from '../../components/AuthRoute';
import { useLogin } from '../../components/LoginProvider';
import { UserRoleEnum } from '../../api/dashboard/v1alpha1';

//--------------------------------------------------
// mock definition
//--------------------------------------------------
jest.mock('../../components/LoginProvider');

type MockedMemberFunction<T extends (...args: any) => any> = {
  [P in keyof ReturnType<T>]: jest.MockedFunction<ReturnType<T>[P]>;
};

//-----------------------------------------------
// test
//-----------------------------------------------

describe('AuthRoute', () => {

  const useLoginMock = useLogin as jest.MockedFunction<typeof useLogin>;
  const loginMock: MockedMemberFunction<typeof useLogin> = {
    loginUser: undefined as any,
    verifyLogin: jest.fn(),
    login: jest.fn(),
    logout: jest.fn(),
    updataPassword: jest.fn(),
  }

  beforeEach(async () => {
    useLoginMock.mockReturnValue(loginMock);
  });

  afterEach(() => { jest.restoreAllMocks(); });

  const routerTester = (path: string) => {
    return render(
      <MemoryRouter initialEntries={[path]}>
        <Routes>
          <Route path="/signin" element={<div>signin</div>} />
          <Route path="/workspace" element={<AuthRoute><div>workspace</div></AuthRoute>} />
          <Route path="/user" element={<AuthRoute admin><div>user</div></AuthRoute>} />
          <Route path='*' element={<Box>404</Box>} />
        </Routes>
      </MemoryRouter >
    );
  }

  it('normal not login =>/signin', async () => {
    const { asFragment } = routerTester('/workspace');
    expect(asFragment()).toMatchSnapshot();
  });

  it('normal not login admin =>/signin', async () => {
    const { asFragment } = routerTester('/user');
    expect(asFragment()).toMatchSnapshot();
  });

  it('normal login =>/workspace', async () => {
    useLoginMock.mockReturnValue({ loginUser: { id: 'user1' } } as ReturnType<typeof useLogin>);
    const { asFragment } = routerTester('/workspace');
    expect(asFragment()).toMatchSnapshot();
  });

  it('normal login admin => /user', async () => {
    useLoginMock.mockReturnValue({ loginUser: { id: 'user1', role: UserRoleEnum.CosmoAdmin } } as ReturnType<typeof useLogin>);
    const { asFragment } = routerTester('/user');
    expect(asFragment()).toMatchSnapshot();
  });

  it('normal login admin page not admin user page => 404', async () => {
    useLoginMock.mockReturnValue({ loginUser: { id: 'user1' } } as ReturnType<typeof useLogin>);
    const { asFragment } = routerTester('/user');
    expect(asFragment()).toMatchSnapshot();
  });

  it('normal login another page => 404', async () => {
    useLoginMock.mockReturnValue({ loginUser: { id: 'user1' } } as ReturnType<typeof useLogin>);
    const { asFragment } = routerTester('/xxx');
    expect(asFragment()).toMatchSnapshot();
  });
});