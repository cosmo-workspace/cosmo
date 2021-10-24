import { Button } from "@mui/material";
import { act, cleanup, render, RenderResult, screen } from "@testing-library/react";
import userEvent from '@testing-library/user-event';
import { useSnackbar } from "notistack";
import { useLogin } from "../../../components/LoginProvider";
import { PasswordChangeDialog, PasswordChangeDialogContext } from "../../../views/organisms/PasswordChangeDialog";

//--------------------------------------------------
// mock definition
//--------------------------------------------------
jest.mock("notistack");
jest.mock("../../../components/LoginProvider");

type MockedMemberFunction<T extends (...args: any) => any> = {
  [P in keyof ReturnType<T>]: jest.MockedFunction<ReturnType<T>[P]>;
};

const useLoginMock = useLogin as jest.MockedFunction<typeof useLogin>;
const loginMock: MockedMemberFunction<typeof useLogin> = {
  loginUser: {} as any,
  verifyLogin: jest.fn(),
  login: jest.fn(),
  logout: jest.fn(),
  updataPassword: jest.fn(),
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
describe("PasswordChangeDialog", () => {

  beforeEach(async () => {
    useSnackbarMock.mockReturnValue(snackbarMock);
    useLoginMock.mockReturnValue(loginMock);
  });

  afterEach(() => {
    jest.restoreAllMocks();
    cleanup();
  });


  describe("render", () => {

    it("render", async () => {
      const target = render(
        <PasswordChangeDialog onClose={() => closeHandlerMock()} />
      );
      const { baseElement } = target;
      expect(baseElement).toMatchSnapshot();
    });

  });


  describe("behavior", () => {

    let target: RenderResult;

    beforeEach(async () => {
      target = render(
        <PasswordChangeDialog onClose={() => closeHandlerMock()} />
      );
    });

    it("ok", async () => {
      const baseElement = target.baseElement;
      userEvent.type(baseElement.querySelector('[name="currentPassword"]')!, 'oldPassword');
      userEvent.type(baseElement.querySelector('[name="newPassword1"]')!, 'newPassword');
      userEvent.type(baseElement.querySelector('[name="newPassword2"]')!, 'newPassword');
      //expect(newPassword2).toHaveValue('newPassword');

      await act(async () => {
        loginMock.updataPassword.mockResolvedValue({} as any);
        userEvent.click(screen.getByText('Change Password'));
      });
      expect(loginMock.updataPassword.mock.calls).toMatchObject([["oldPassword", "newPassword"]]);
      expect(closeHandlerMock.mock.calls.length).toEqual(1);
    });

    it("ng required", async () => {
      expect(target.getAllByText('Current password')[0].parentElement!
        .getElementsByClassName('MuiFormHelperText-root')[0]).toBeUndefined();
      expect(target.getAllByText('New password')[0].parentElement!
        .getElementsByClassName('MuiFormHelperText-root')[0]).toBeUndefined();
      expect(target.getAllByText('Confirm password')[0].parentElement!
        .getElementsByClassName('MuiFormHelperText-root')[0]).toBeUndefined();

      await act(async () => {
        userEvent.click(screen.getByText('Change Password'));
      });
      expect(target.getAllByText('Current password')[0].parentElement!
        .getElementsByClassName('MuiFormHelperText-root')[0]).toHaveTextContent('Required');
      expect(target.getAllByText('New password')[0].parentElement!
        .getElementsByClassName('MuiFormHelperText-root')[0]).toHaveTextContent('Required');
      expect(target.getAllByText('Confirm password')[0].parentElement!
        .getElementsByClassName('MuiFormHelperText-root')[0]).toHaveTextContent('Required');
      expect(loginMock.updataPassword.mock.calls.length).toEqual(0);
      expect(closeHandlerMock.mock.calls.length).toEqual(0);
    });

    it("ng Contains spaces", async () => {
      const baseElement = target.baseElement;
      userEvent.type(baseElement.querySelector('[name="currentPassword"]')!, ' 12345 ');
      userEvent.type(baseElement.querySelector('[name="newPassword1"]')!, '54 3 21');
      userEvent.type(baseElement.querySelector('[name="newPassword2"]')!, '54 3 21');

      expect(target.getAllByText('Current password')[0].parentElement!
        .getElementsByClassName('MuiFormHelperText-root')[0]).toBeUndefined();
      expect(target.getAllByText('New password')[0].parentElement!
        .getElementsByClassName('MuiFormHelperText-root')[0]).toBeUndefined();
      expect(target.getAllByText('Confirm password')[0].parentElement!
        .getElementsByClassName('MuiFormHelperText-root')[0]).toBeUndefined();

      await act(async () => {
        userEvent.click(screen.getByText('Change Password'));
      });
      expect(target.getAllByText('Current password')[0].parentElement!
        .getElementsByClassName('MuiFormHelperText-root')[0]).toHaveTextContent('Contains spaces');
      expect(target.getAllByText('New password')[0].parentElement!
        .getElementsByClassName('MuiFormHelperText-root')[0]).toHaveTextContent('Contains spaces');
      expect(target.getAllByText('Confirm password')[0].parentElement!
        .getElementsByClassName('MuiFormHelperText-root')[0]).toHaveTextContent('Contains spaces');
      expect(loginMock.updataPassword.mock.calls.length).toEqual(0);
      expect(closeHandlerMock.mock.calls.length).toEqual(0);
    });

    it("ng Contains spaces", async () => {
      const baseElement = target.baseElement;
      userEvent.type(baseElement.querySelector('[name="currentPassword"]')!, 'oldPassword');
      userEvent.type(baseElement.querySelector('[name="newPassword1"]')!, 'newPassword');
      userEvent.type(baseElement.querySelector('[name="newPassword2"]')!, 'XXXXXXXX');

      expect(target.getAllByText('Current password')[0].parentElement!
        .getElementsByClassName('MuiFormHelperText-root')[0]).toBeUndefined();
      expect(target.getAllByText('New password')[0].parentElement!
        .getElementsByClassName('MuiFormHelperText-root')[0]).toBeUndefined();
      expect(target.getAllByText('Confirm password')[0].parentElement!
        .getElementsByClassName('MuiFormHelperText-root')[0]).toHaveTextContent('Passwords do not match');

      await act(async () => {
        userEvent.click(screen.getByText('Change Password'));
      });
      expect(target.getAllByText('Current password')[0].parentElement!
        .getElementsByClassName('MuiFormHelperText-root')[0]).toBeUndefined();
      expect(target.getAllByText('New password')[0].parentElement!
        .getElementsByClassName('MuiFormHelperText-root')[0]).toBeUndefined();
      expect(target.getAllByText('Confirm password')[0].parentElement!
        .getElementsByClassName('MuiFormHelperText-root')[0]).toHaveTextContent('Passwords do not match');
      expect(loginMock.updataPassword.mock.calls.length).toEqual(0);
      expect(closeHandlerMock.mock.calls.length).toEqual(0);
    });

    it("cancel", async () => {
      userEvent.click(screen.getByText('Cancel'));
      expect(loginMock.updataPassword.mock.calls.length).toEqual(0);
      expect(closeHandlerMock.mock.calls.length).toEqual(1);
    });

    it("close <- click outside", async () => {
      userEvent.click(target.baseElement.getElementsByClassName('MuiDialog-container')[0]);
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

      let target: RenderResult;
      await act(async () => {
        target = render(
          <PasswordChangeDialogContext.Provider><Stub /></PasswordChangeDialogContext.Provider>
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