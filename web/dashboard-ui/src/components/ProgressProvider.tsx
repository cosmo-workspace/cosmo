import { Backdrop, CircularProgress } from '@mui/material';
import React, { createContext, useContext, useState } from 'react';

const DispatchContext = createContext<React.Dispatch<React.SetStateAction<number>>>(undefined as any);

/**
 * provider
 */
export const ProgressProvider: React.FC<React.PropsWithChildren<unknown>> = ({ children }) => {
  const [count, setCount] = useState(0);
  return (<>
    <DispatchContext.Provider value={setCount}>
      {children}
    </DispatchContext.Provider>
    <Backdrop sx={{ zIndex: (theme) => theme.zIndex.drawer + 1000 }} open={count > 0}>
      <CircularProgress />
    </Backdrop>
  </>);
}

export function useProgress() {
  const setCount = useContext(DispatchContext);
  return {
    setMask: () => setCount(count => { console.log('setMask', count + 1); return count + 1 }),
    releaseMask: () => setCount(count => { console.log('releaseMask', count); return count > 0 ? count - 1 : 0 }),
  }
}