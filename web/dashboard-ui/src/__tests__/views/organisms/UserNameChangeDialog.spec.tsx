import { Button } from "@mui/material";
import { act, cleanup, render, RenderResult, screen } from "@testing-library/react";
import userEvent from '@testing-library/user-event';
import { useSnackbar } from "notistack";
import { User, UserRoleEnum } from "../../../api/dashboard/v1alpha1";
import { useLogin } from "../../../components/LoginProvider";
import { useUserModule } from "../../../views/organisms/UserModule";
import { UserNameChangeDialog, UserNameChangeDialogContext } from "../../../views/organisms/UserNameChangeDialog";

//--------------------------------------------------
// mock definition
//--------------------------------------------------
jest.mock("notistack");
jest.mock("../../../components/LoginProvider");
jest.mock("../../../views/organisms/UserModule");

type MockedMemberFunction<T extends (...args: any) => any> = {
  [P in keyof ReturnType<T>]: jest.MockedFunction<ReturnType<T>[P]>;
};

const useUserModuleMock = useUserModule as jest.MockedFunction<typeof useUserModule>;
const userModuleMock: MockedMemberFunction<typeof useUserModule> = {
  users: [] as any,
  getUsers: jest.fn(),
  createUser: jest.fn(),
  updateName: jest.fn(),
  updateRole: jest.fn(),
  deleteUser: jest.fn(),
}
const useLoginMock = useLogin as jest.MockedFunction<typeof useLogin>;
const loginMock: MockedMemberFunction<typeof useLogin> = {
  loginUser: {} as any,
  verifyLogin: jest.fn(),
  login: jest.fn(),
  logout: jest.fn(),
  updataPassword: jest.fn(),
  refreshUserInfo: jest.fn(),
};

const useSnackbarMock = useSnackbar as jest.MockedFunction<typeof useSnackbar>;
const snackbarMock: MockedMemberFunction<typeof useSnackbar> = {
  enqueueSnackbar: jest.fn(),
  closeSnackbar: jest.fn(),
};

const closeHandlerMock = jest.fn();

//--------------------------------------------------
// test
//--------------------------------------------------
describe("UserNameChangeDialog", () => {

  beforeEach(async () => {
    useSnackbarMock.mockReturnValue(snackbarMock);
    useUserModuleMock.mockReturnValue(userModuleMock);
    useLoginMock.mockReturnValue(loginMock);
  });

  afterEach(() => {
    jest.restoreAllMocks();
    cleanup();
  });


  describe("render", () => {

    it("render", async () => {
      const user1: User = { id: 'user1', role: UserRoleEnum.CosmoAdmin, displayName: 'user1 name' };
      const target = render(
        <UserNameChangeDialog onClose={() => closeHandlerMock()} user={user1} />
      );
      const { baseElement } = target;
      expect(baseElement).toMatchSnapshot();
    });

  });


  describe("behavior", () => {

    let target: RenderResult;

    beforeEach(async () => {
      const user1: User = { id: 'user1', role: UserRoleEnum.CosmoAdmin, displayName: 'user1 name' };
      target = render(
        <UserNameChangeDialog onClose={() => closeHandlerMock()} user={user1} />
      );
    });

    it("ok", async () => {
      const baseElement = target.baseElement;
      const nameElement = baseElement.querySelector('[name="name"]')!;
      (nameElement as HTMLInputElement).setSelectionRange(0, 99);
      userEvent.type(baseElement.querySelector('[name="name"]')!, 'New Name');

      await act(async () => {
        userModuleMock.updateName.mockResolvedValue({} as any);
        loginMock.refreshUserInfo.mockResolvedValue({} as any);
        userEvent.click(screen.getByText('Update'));
      });
      expect(userModuleMock.updateName.mock.calls).toMatchObject([["user1", "New Name"]]);
      expect(loginMock.refreshUserInfo.mock.calls).toMatchObject([[]]);
      expect(closeHandlerMock.mock.calls.length).toEqual(1);
    });

    it("ng required", async () => {
      expect(target.getAllByText('User Name')[0].parentElement!
        .getElementsByClassName('MuiFormHelperText-root')[0]).toBeUndefined();

      const baseElement = target.baseElement;
      const nameElement = baseElement.querySelector('[name="name"]')!;
      (nameElement as HTMLInputElement).setSelectionRange(0, 99);
      userEvent.type(nameElement, '{backspace}');

      await act(async () => {
        userEvent.click(screen.getByText('Update'));
      });
      expect(target.getAllByText('User Name')[0].parentElement!
        .getElementsByClassName('MuiFormHelperText-root')[0]).toHaveTextContent('Required');
      expect(userModuleMock.updateName.mock.calls.length).toEqual(0);
      expect(loginMock.refreshUserInfo.mock.calls.length).toEqual(0);
      expect(closeHandlerMock.mock.calls.length).toEqual(0);
    });

    it("ng over 32", async () => {
      expect(target.getAllByText('User Name')[0].parentElement!
        .getElementsByClassName('MuiFormHelperText-root')[0]).toBeUndefined();

      const baseElement = target.baseElement;
      const nameElement = baseElement.querySelector('[name="name"]')!;
      (nameElement as HTMLInputElement).setSelectionRange(0, 99);
      userEvent.type(nameElement, '----+----1----+----2----+----3--x');

      await act(async () => {
        userEvent.click(screen.getByText('Update'));
      });
      expect(target.getAllByText('User Name')[0].parentElement!
        .getElementsByClassName('MuiFormHelperText-root')[0]).toHaveTextContent('Max 32 characters');
      expect(userModuleMock.updateName.mock.calls.length).toEqual(0);
      expect(loginMock.refreshUserInfo.mock.calls.length).toEqual(0);
      expect(closeHandlerMock.mock.calls.length).toEqual(0);

      await act(async () => {
        (nameElement as HTMLInputElement).setSelectionRange(0, 99);
        userEvent.type(nameElement, '----+----1----+----2----+----3--');
      });
      expect(target.getAllByText('User Name')[0].parentElement!
        .getElementsByClassName('MuiFormHelperText-root')[0]).toBeUndefined();
    });

    it("cancel", async () => {
      userEvent.click(screen.getByText('Cancel'));
      expect(userModuleMock.updateName.mock.calls.length).toEqual(0);
      expect(loginMock.refreshUserInfo.mock.calls.length).toEqual(0);
      expect(closeHandlerMock.mock.calls.length).toEqual(1);
    });

    it("not close <- click outside", async () => {
      userEvent.click(target.baseElement.getElementsByClassName('MuiDialog-container')[0]);
      expect(userModuleMock.updateName.mock.calls.length).toEqual(0);
      expect(loginMock.refreshUserInfo.mock.calls.length).toEqual(0);
      expect(closeHandlerMock.mock.calls.length).toEqual(0);
    });

  });


  describe("UserNameChangeDialogContext", () => {

    it("open/close", async () => {

      const Stub = () => {
        const dispatch = UserNameChangeDialogContext.useDispatch();
        const user1: User = { id: 'user1', displayName: 'user1name' }
        return (<>
          <Button onClick={() => dispatch(true, { user: user1 })}>open</Button>
          <Button onClick={() => dispatch(false)}>close</Button>
        </>);
      }

      let target: RenderResult;
      await act(async () => {
        target = render(
          <UserNameChangeDialogContext.Provider><Stub /></UserNameChangeDialogContext.Provider>
        );
      });
      await act(async () => { expect(target.baseElement).toMatchSnapshot('1.initial render'); });
      await act(async () => { userEvent.click(target.getByText('open')); });
      await act(async () => { expect(target.baseElement).toMatchSnapshot('2.opend'); });
      await act(async () => { userEvent.click(screen.getByText('close')); });
      await act(async () => { expect(target.baseElement).toMatchSnapshot('3.closed'); });
    });

  });

});