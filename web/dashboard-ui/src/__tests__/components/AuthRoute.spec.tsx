import { Box } from '@mui/system';
import { render } from '@testing-library/react';
import React from 'react';
import { MemoryRouter, Route, Switch } from 'react-router-dom';
import { AuthRoute } from '../../components/AuthRoute';
import { useLogin } from '../../components/LoginProvider';

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
  // todo: compile error
  const loginMock: MockedMemberFunction<typeof useLogin> = {
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
        <Switch>
          <Route path="/signin" component={() => (<div>signin</div>)} exact />
          <AuthRoute path="/workspace" component={() => (<div>workspace</div>)} exact />
          <AuthRoute path="/user" component={() => (<div>user</div>)} admin exact />
          <Route component={() => (<Box>404</Box>)} />
        </Switch>
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
    useLoginMock.mockReturnValue({ loginUser: { id: 'user1', role: 'cosmo-admin' } } as ReturnType<typeof useLogin>);
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