import { Button } from "@mui/material";
import '@testing-library/jest-dom';
import { act, cleanup, render, screen } from "@testing-library/react";
import userEvent from '@testing-library/user-event';
import { useSnackbar } from "notistack";
import React from "react";
import { MockedFunction, afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { useLogin } from "../../../components/LoginProvider";
import { User } from "../../../proto/gen/dashboard/v1alpha1/user_pb";
import { useUserModule } from "../../../views/organisms/UserModule";
import { UserNameChangeDialog, UserNameChangeDialogContext } from "../../../views/organisms/UserNameChangeDialog";

//--------------------------------------------------
// mock definition
//--------------------------------------------------
vi.mock("notistack");
vi.mock("../../../components/LoginProvider");
vi.mock("../../../views/organisms/UserModule");

type MockedMemberFunction<T extends (...args: any) => any> = {
  [P in keyof ReturnType<T>]: MockedFunction<ReturnType<T>[P]>;
};

const useUserModuleMock = useUserModule as MockedFunction<typeof useUserModule>;
const userModuleMock: MockedMemberFunction<typeof useUserModule> = {
  existingRoles: [] as any,
  users: [] as any,
  getUsers: vi.fn(),
  createUser: vi.fn(),
  updateName: vi.fn(),
  updateRole: vi.fn(),
  deleteUser: vi.fn(),
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

const useSnackbarMock = useSnackbar as MockedFunction<typeof useSnackbar>;
const snackbarMock: MockedMemberFunction<typeof useSnackbar> = {
  enqueueSnackbar: vi.fn(),
  closeSnackbar: vi.fn(),
};

const closeHandlerMock = vi.fn();

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
    vi.restoreAllMocks();
    cleanup();
  });


  describe("render", () => {

    it("render", async () => {
      const user1: User = new User({ name: 'user1', roles: ["CosmoAdmin"], displayName: 'user1 name' });
      render(
        <UserNameChangeDialog onClose={() => closeHandlerMock()} user={user1} />
      );
      expect(document.body).toMatchSnapshot();
    });

  });


  describe("behavior", () => {

    const user1: User = new User({ name: 'user1', roles: ["CosmoAdmin"], displayName: 'user1 name' });

    it("ok", async () => {
      const user = userEvent.setup();
      const { baseElement } = render(<UserNameChangeDialog onClose={() => closeHandlerMock()} user={user1} />);

      const nameElement = baseElement.querySelector('[name="name"]')!;
      await user.type(nameElement, 'New Name', { initialSelectionStart: 0, initialSelectionEnd: 99 });

      userModuleMock.updateName.mockResolvedValue({} as any);
      loginMock.refreshUserInfo.mockResolvedValue({} as any);
      await user.click(screen.getByText('Update'));
      expect(userModuleMock.updateName.mock.calls).toMatchObject([["user1", "New Name"]]);
      expect(loginMock.refreshUserInfo.mock.calls).toMatchObject([[]]);
      expect(closeHandlerMock.mock.calls.length).toEqual(1);
    });

    it("ng required", async () => {
      const user = userEvent.setup();
      const { baseElement } = render(<UserNameChangeDialog onClose={() => closeHandlerMock()} user={user1} />);

      expect(screen.getAllByText('User Name')[0].parentElement!
        .getElementsByClassName('MuiFormHelperText-root')[0]).toBeUndefined();

      const nameElement = baseElement.querySelector('[name="name"]')!;
      await user.type(nameElement, '{backspace}', { initialSelectionStart: 0, initialSelectionEnd: 99 });
      await act(async () => {
        await user.click(screen.getByText('Update'));
      });
      expect(screen.getAllByText('User Name')[0].parentElement!
        .getElementsByClassName('MuiFormHelperText-root')[0]).toHaveTextContent('Required');
      expect(userModuleMock.updateName.mock.calls.length).toEqual(0);
      expect(loginMock.refreshUserInfo.mock.calls.length).toEqual(0);
      expect(closeHandlerMock.mock.calls.length).toEqual(0);
    });

    it("ng over 32", async () => {
      const user = userEvent.setup();
      const { baseElement } = render(<UserNameChangeDialog onClose={() => closeHandlerMock()} user={user1} />);

      expect(screen.getAllByText('User Name')[0].parentElement!
        .getElementsByClassName('MuiFormHelperText-root')[0]).toBeUndefined();

      const nameElement = baseElement.querySelector('[name="name"]')!;
      await user.type(nameElement, '----+----1----+----2----+----3--x', { initialSelectionStart: 0, initialSelectionEnd: 99 });
      await act(async () => {
        await user.click(screen.getByText('Update'));
      });

      expect(screen.getAllByText('User Name')[0].parentElement!
        .getElementsByClassName('MuiFormHelperText-root')[0]).toHaveTextContent('Max 32 characters');
      expect(userModuleMock.updateName.mock.calls.length).toEqual(0);
      expect(loginMock.refreshUserInfo.mock.calls.length).toEqual(0);
      expect(closeHandlerMock.mock.calls.length).toEqual(0);

      await act(async () => {
        await user.type(nameElement, '----+----1----+----2----+----3--', { initialSelectionStart: 0, initialSelectionEnd: 99 });
      });
      expect(screen.getAllByText('User Name')[0].parentElement!
        .getElementsByClassName('MuiFormHelperText-root')[0]).toBeUndefined();
    });

    it("cancel", async () => {
      const user = userEvent.setup();
      render(<UserNameChangeDialog onClose={() => closeHandlerMock()} user={user1} />);
      await user.click(screen.getByText('Cancel'));
      expect(userModuleMock.updateName.mock.calls.length).toEqual(0);
      expect(loginMock.refreshUserInfo.mock.calls.length).toEqual(0);
      expect(closeHandlerMock.mock.calls.length).toEqual(1);
    });

    it("not close <- click outside", async () => {
      const user = userEvent.setup();
      const { baseElement } = render(<UserNameChangeDialog onClose={() => closeHandlerMock()} user={user1} />);
      await user.click(baseElement.getElementsByClassName('MuiDialog-container')[0]);
      expect(userModuleMock.updateName.mock.calls.length).toEqual(0);
      expect(loginMock.refreshUserInfo.mock.calls.length).toEqual(0);
      expect(closeHandlerMock.mock.calls.length).toEqual(0);
    });

  });


  describe("UserNameChangeDialogContext", () => {

    it("open/close", async () => {

      const Stub = () => {
        const dispatch = UserNameChangeDialogContext.useDispatch();
        const user1: User = new User({ name: 'user1', displayName: 'user1name' });
        return (<>
          <Button onClick={() => dispatch(true, { user: user1 })}>open</Button>
          <Button onClick={() => dispatch(false)}>close</Button>
        </>);
      }
      const user = userEvent.setup();
      render(<UserNameChangeDialogContext.Provider><Stub /></UserNameChangeDialogContext.Provider>);
      await act(async () => { expect(document.body).toMatchSnapshot('1.initial render'); });
      await act(async () => { await user.click(screen.getByText('open')); });
      await act(async () => { expect(document.body).toMatchSnapshot('2.opend'); });
      await act(async () => { await user.click(screen.getByText('close')); });
      await act(async () => { expect(document.body).toMatchSnapshot('3.closed'); });
    });

  });

});