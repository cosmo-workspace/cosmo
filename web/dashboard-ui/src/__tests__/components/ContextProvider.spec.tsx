import { Button, Dialog } from "@mui/material";
import { act, cleanup, render, screen } from "@testing-library/react";
import userEvent from '@testing-library/user-event';
import React, { useState } from "react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { DialogContext, ModuleContext } from "../../components/ContextProvider";

//--------------------------------------------------
// test
//--------------------------------------------------
describe("DialogContext", () => {

  beforeEach(async () => {
  });

  afterEach(() => {
    vi.restoreAllMocks();
    cleanup();
  });

  const TestDialog: React.VFC<{ onClose: () => void, hoge: string }> = ({ onClose, hoge }) => {
    return (
      <Dialog open={true} onClose={() => onClose()}>
      </Dialog>
    )
  }

  const testDialogContext = DialogContext<{ hoge: string }>(
    props => (<TestDialog {...props} />));

  it("open/close", async () => {

    const Stub = () => {
      const dispatch = testDialogContext.useDispatch();
      return (<>
        <Button onClick={() => dispatch(true)}>open</Button>
        <Button onClick={() => dispatch(false)}>close</Button>
      </>);
    }
    const user = userEvent.setup();
    render(<testDialogContext.Provider><Stub /></testDialogContext.Provider>);
    await act(async () => { expect(document.body).toMatchSnapshot('1.initial render'); });
    await act(async () => { await user.click(screen.getByText('open')); });
    await act(async () => { expect(document.body).toMatchSnapshot('2.opend'); });
    await act(async () => { await user.click(screen.getByText('close')); });
    await act(async () => { expect(document.body).toMatchSnapshot('3.closed'); });
  });

  it("close esc", async () => {

    const Stub = () => {
      const dispatch = testDialogContext.useDispatch();
      return (<>
        <Button onClick={() => dispatch(true)}>open</Button>
        <Button onClick={() => dispatch(false)}>close</Button>
      </>);
    }
    const user = userEvent.setup();
    render(<testDialogContext.Provider><Stub /></testDialogContext.Provider>);
    await act(async () => { await user.click(screen.getByText('open')); });
    await act(async () => { await user.keyboard('{Esc}'); });
    await act(async () => { expect(document.body).toMatchSnapshot('closed'); });
  });

});


describe('ModuleContext', () => {

  it('normal', async () => {

    const useTestHook = () => {
      const [state, setState] = useState("aaaa");
      return { state, setState }
    }

    const TestContext = ModuleContext(useTestHook);
    const useTest = TestContext.useContext;

    const Component1: React.FC<React.PropsWithChildren<unknown>> = ({ children }) => {
      const { state, setState } = useTest();
      return (<div>
        <button onClick={() => { setState("11111") }}>button1</button>
        <div data-testid="textbox1">{state}</div>
      </div>);
    }
    const Component2: React.FC<React.PropsWithChildren<unknown>> = ({ children }) => {
      const { state, setState } = useTest();
      return (<div>
        <button onClick={() => { setState("22222") }}>button2</button>
        <div data-testid="textbox2">{state}</div>
      </div>);
    }
    const user = userEvent.setup();
    render(
      <TestContext.Provider>
        <Component1 />
        <Component2 />
      </TestContext.Provider>
    );

    await user.click(screen.getByText('button1'));
    expect(screen.getByTestId('textbox1')).toHaveTextContent('11111');
    expect(screen.getByTestId('textbox2')).toHaveTextContent('11111');
    await user.click(screen.getByText('button2'));
    expect(screen.getByTestId('textbox1')).toHaveTextContent('22222');
    expect(screen.getByTestId('textbox2')).toHaveTextContent('22222');
  });

});
