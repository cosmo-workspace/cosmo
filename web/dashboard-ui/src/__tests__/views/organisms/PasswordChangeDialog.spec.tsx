import { Button } from "@mui/material";
import '@testing-library/jest-dom';
import { act, cleanup, render, screen } from "@testing-library/react";
import userEvent from '@testing-library/user-event';
import { useSnackbar } from "notistack";
import React from "react";
import { afterEach, beforeEach, describe, expect, it, MockedFunction, vi } from "vitest";
import { useLogin } from "../../../components/LoginProvider";
import { PasswordChangeDialog, PasswordChangeDialogContext } from "../../../views/organisms/PasswordChangeDialog";

//--------------------------------------------------
// mock definition
//--------------------------------------------------
vi.mock("notistack");
vi.mock("../../../components/LoginProvider");

type MockedMemberFunction<T extends (...args: any) => any> = {
  [P in keyof ReturnType<T>]: MockedFunction<ReturnType<T>[P]>;
};

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
describe("PasswordChangeDialog", () => {

  beforeEach(async () => {
    useSnackbarMock.mockReturnValue(snackbarMock);
    useLoginMock.mockReturnValue(loginMock);
  });

  afterEach(() => {
    vi.restoreAllMocks();
    cleanup();
  });


  describe("render", () => {

    it("render", async () => {
      render(
        <PasswordChangeDialog onClose={() => closeHandlerMock()} />
      );
      expect(document.body).toMatchSnapshot();
    });

  });


  describe("behavior", () => {

    it("ok", async () => {
      const user = userEvent.setup();
      const { baseElement } = render(<PasswordChangeDialog onClose={() => closeHandlerMock()} />);
      await user.type(baseElement.querySelector('[name="currentPassword"]')!, 'oldPassword');
      await user.type(baseElement.querySelector('[name="newPassword1"]')!, 'newPassword');
      await user.type(baseElement.querySelector('[name="newPassword2"]')!, 'newPassword');
      loginMock.updataPassword.mockResolvedValue({} as any);
      await user.click(screen.getByText('Change Password'));
      expect(loginMock.updataPassword.mock.calls).toMatchObject([["oldPassword", "newPassword"]]);
      expect(closeHandlerMock.mock.calls.length).toEqual(1);
    });

    it("ng required", async () => {
      const user = userEvent.setup();
      render(<PasswordChangeDialog onClose={() => closeHandlerMock()} />);
      expect(screen.getAllByText('Current password')[0].parentElement!
        .getElementsByClassName('MuiFormHelperText-root')[0]).toBeUndefined();
      expect(screen.getAllByText('New password')[0].parentElement!
        .getElementsByClassName('MuiFormHelperText-root')[0]).toBeUndefined();
      expect(screen.getAllByText('Confirm password')[0].parentElement!
        .getElementsByClassName('MuiFormHelperText-root')[0]).toBeUndefined();
      await act(async () => {
        await user.click(screen.getByText('Change Password'));
      });
      expect(screen.getAllByText('Current password')[0].parentElement!
        .getElementsByClassName('MuiFormHelperText-root')[0]).toHaveTextContent('Required');
      expect(screen.getAllByText('New password')[0].parentElement!
        .getElementsByClassName('MuiFormHelperText-root')[0]).toHaveTextContent('Required');
      expect(screen.getAllByText('Confirm password')[0].parentElement!
        .getElementsByClassName('MuiFormHelperText-root')[0]).toHaveTextContent('Required');
      expect(loginMock.updataPassword.mock.calls.length).toEqual(0);
      expect(closeHandlerMock.mock.calls.length).toEqual(0);
    });

    it("ng Contains spaces", async () => {
      const user = userEvent.setup();
      const { baseElement } = render(<PasswordChangeDialog onClose={() => closeHandlerMock()} />);
      await user.type(baseElement.querySelector('[name="currentPassword"]')!, ' 12345 ');
      await user.type(baseElement.querySelector('[name="newPassword1"]')!, '54 3 21');
      await user.type(baseElement.querySelector('[name="newPassword2"]')!, '54 3 21');

      expect(screen.getAllByText('Current password')[0].parentElement!
        .getElementsByClassName('MuiFormHelperText-root')[0]).toBeUndefined();
      expect(screen.getAllByText('New password')[0].parentElement!
        .getElementsByClassName('MuiFormHelperText-root')[0]).toBeUndefined();
      expect(screen.getAllByText('Confirm password')[0].parentElement!
        .getElementsByClassName('MuiFormHelperText-root')[0]).toBeUndefined();
      await act(async () => {
        await user.click(screen.getByText('Change Password'));
      });
      expect(screen.getAllByText('Current password')[0].parentElement!
        .getElementsByClassName('MuiFormHelperText-root')[0]).toHaveTextContent('Contains spaces');
      expect(screen.getAllByText('New password')[0].parentElement!
        .getElementsByClassName('MuiFormHelperText-root')[0]).toHaveTextContent('Contains spaces');
      expect(screen.getAllByText('Confirm password')[0].parentElement!
        .getElementsByClassName('MuiFormHelperText-root')[0]).toHaveTextContent('Contains spaces');
      expect(loginMock.updataPassword.mock.calls.length).toEqual(0);
      expect(closeHandlerMock.mock.calls.length).toEqual(0);
    });

    it("ng Contains spaces", async () => {
      const user = userEvent.setup();
      const { baseElement } = render(<PasswordChangeDialog onClose={() => closeHandlerMock()} />);
      await user.type(baseElement.querySelector('[name="currentPassword"]')!, 'oldPassword');
      await user.type(baseElement.querySelector('[name="newPassword1"]')!, 'newPassword');
      await user.type(baseElement.querySelector('[name="newPassword2"]')!, 'XXXXXXXX');

      expect(screen.getAllByText('Current password')[0].parentElement!
        .getElementsByClassName('MuiFormHelperText-root')[0]).toBeUndefined();
      expect(screen.getAllByText('New password')[0].parentElement!
        .getElementsByClassName('MuiFormHelperText-root')[0]).toBeUndefined();
      expect(screen.getAllByText('Confirm password')[0].parentElement!
        .getElementsByClassName('MuiFormHelperText-root')[0]).toHaveTextContent('Passwords do not match');
      await act(async () => {
        await user.click(screen.getByText('Change Password'));
      });
      expect(screen.getAllByText('Current password')[0].parentElement!
        .getElementsByClassName('MuiFormHelperText-root')[0]).toBeUndefined();
      expect(screen.getAllByText('New password')[0].parentElement!
        .getElementsByClassName('MuiFormHelperText-root')[0]).toBeUndefined();
      expect(screen.getAllByText('Confirm password')[0].parentElement!
        .getElementsByClassName('MuiFormHelperText-root')[0]).toHaveTextContent('Passwords do not match');
      expect(loginMock.updataPassword.mock.calls.length).toEqual(0);
      expect(closeHandlerMock.mock.calls.length).toEqual(0);
    });

    it("cancel", async () => {
      const user = userEvent.setup();
      render(<PasswordChangeDialog onClose={() => closeHandlerMock()} />);
      await user.click(screen.getByText('Cancel'));
      expect(loginMock.updataPassword.mock.calls.length).toEqual(0);
      expect(closeHandlerMock.mock.calls.length).toEqual(1);
    });

    it("close <- click outside", async () => {
      const user = userEvent.setup();
      const { baseElement } = render(<PasswordChangeDialog onClose={() => closeHandlerMock()} />);
      await user.click(baseElement.getElementsByClassName('MuiDialog-container')[0]);
      expect(loginMock.updataPassword.mock.calls.length).toEqual(0);
      expect(closeHandlerMock.mock.calls.length).toEqual(1);
    });

  });


  describe("PasswordChangeDialogContext", () => {

    it("open/close", async () => {

      const Stub = () => {
        const dispatch = PasswordChangeDialogContext.useDispatch();
        return (<>
          <Button onClick={() => dispatch(true)}>open</Button>
          <Button onClick={() => dispatch(false)}>close</Button>
        </>);
      }
      const user = userEvent.setup();
      render(<PasswordChangeDialogContext.Provider><Stub /></PasswordChangeDialogContext.Provider>);
      await act(async () => { expect(document.body).toMatchSnapshot('1.initial render'); });
      await act(async () => { await user.click(screen.getByText('open')); });
      await act(async () => { expect(document.body).toMatchSnapshot('2.opend'); });
      await act(async () => { await user.click(screen.getByText('close')); });
      await act(async () => { expect(document.body).toMatchSnapshot('3.closed'); });
    });
  });

});