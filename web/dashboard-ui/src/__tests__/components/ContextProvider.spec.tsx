import { Button, Dialog } from "@mui/material";
import { act, cleanup, render, RenderResult, screen } from "@testing-library/react";
import userEvent from '@testing-library/user-event';
import { useState } from "react";
import { DialogContext, ModuleContext } from "../../components/ContextProvider";

//--------------------------------------------------
// test
//--------------------------------------------------
describe("DialogContext", () => {

  beforeEach(async () => {
  });

  afterEach(() => {
    jest.restoreAllMocks();
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

    let target: RenderResult;
    await act(async () => {
      target = render(
        <testDialogContext.Provider><Stub /></testDialogContext.Provider>
      );
    });
    await act(async () => { expect(target.baseElement).toMatchSnapshot('1.initial render'); });
    await act(async () => { userEvent.click(target.getByText('open')); });
    await act(async () => { expect(target.baseElement).toMatchSnapshot('2.opend'); });
    await act(async () => { userEvent.click(screen.getByText('close')); });
    await act(async () => { expect(target.baseElement).toMatchSnapshot('3.closed'); });
  });

  it("close esc", async () => {

    const Stub = () => {
      const dispatch = testDialogContext.useDispatch();
      return (<>
        <Button onClick={() => dispatch(true)}>open</Button>
        <Button onClick={() => dispatch(false)}>close</Button>
      </>);
    }

    let target: RenderResult;
    await act(async () => {
      target = render(
        <testDialogContext.Provider><Stub /></testDialogContext.Provider>
      );
    });
    await act(async () => { userEvent.click(target.getByText('open')); });
    await act(async () => { userEvent.type(screen.getByText('open'), '{esc}'); });
    await act(async () => { expect(target.baseElement).toMatchSnapshot('closed'); });
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
        <div id="textbox1">{state}</div>
      </div>);
    }
    const Component2: React.FC<React.PropsWithChildren<unknown>> = ({ children }) => {
      const { state, setState } = useTest();
      return (<div>
        <button onClick={() => { setState("22222") }}>button2</button>
        <div id="textbox2">{state}</div>
      </div>);
    }

    const target = render(
      <TestContext.Provider>
        <Component1 />
        <Component2 />
      </TestContext.Provider>
    );

    await act(async () => {
      userEvent.click(target.getByText('button1'));
    });
    expect(target.container.ownerDocument.getElementById('textbox1')).toHaveTextContent('11111');
    expect(target.container.ownerDocument.getElementById('textbox2')).toHaveTextContent('11111');
    await act(async () => {
      userEvent.click(target.getByText('button2'));
    });
    expect(target.container.ownerDocument.getElementById('textbox1')).toHaveTextContent('22222');
    expect(target.container.ownerDocument.getElementById('textbox2')).toHaveTextContent('22222');
  });

});
