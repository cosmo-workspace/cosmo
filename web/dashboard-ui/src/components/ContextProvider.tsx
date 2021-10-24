import { createContext, ReactNode, useContext, useState } from "react";

/**
 * DialogContext
 */
type DialogState<T> = { open: boolean, dialogProps?: T }
type DialogProps<T> = T & { onClose: () => void }
type Dispatch<T> = (open: boolean, dialogProp?: T) => void;

export function DialogContext<T>(dialog: (props: DialogProps<T>) => any) {

  const Context = createContext<{ state: DialogState<T>, dispatch: Dispatch<T> }>(undefined as any);

  return {
    Provider: ({ children }: { children: ReactNode }) => {
      const [state, setState] = useState<DialogState<T>>({ open: false });

      const dispatch: Dispatch<T> = (open, dialogProps?) => {
        setState({ open, dialogProps });
      }
      const closeHandler = () => dispatch(false, state.dialogProps);
      return (
        <Context.Provider value={{ state, dispatch }}><>
          {children}
          {state.open && dialog({ ...state.dialogProps, onClose: () => closeHandler() } as DialogProps<T>)}
        </></Context.Provider>
      );
    },
    useDispatch: () => useContext(Context).dispatch,
  }
}

/**
 * ModuleContext
 */
export function ModuleContext<T extends () => any>(useModule: T) {
  const Context = createContext<ReturnType<T>>(undefined as any);
  return {
    Provider: ({ children }: { children: ReactNode }) => {
      const module = useModule();
      return (<Context.Provider value={{ ...module }}>{children}</Context.Provider>);
    },
    useContext: () => useContext(Context),
  }
}
