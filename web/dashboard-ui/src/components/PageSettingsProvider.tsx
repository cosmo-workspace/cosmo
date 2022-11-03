import React, { createContext, useContext, useEffect, useReducer } from 'react';

interface PageSettings {
  isOpen: boolean;
};

interface PageSettingsContext {
  pageSettings: PageSettings;
  setPageSettings: (setting: PageSettings) => void;
  updatePageSettings: (setting: PageSettings) => void;
};

const Context = createContext<PageSettingsContext>({
  pageSettings: {
    isOpen: true,
  },
  setPageSettings: (_setting) => { },
  updatePageSettings: (_setting) => { },
});

interface Action {
  type: 'SET' | 'UPDATE';
  pageSettings: PageSettings;
}
function reducer(state: PageSettings, action: Action) {
  switch (action.type) {
    case 'SET':
      return action.pageSettings;
    case 'UPDATE':
      return { ...state, ...action.pageSettings }
    default:
      throw new Error()
  }
}

export const PageSettingsProvider: React.FC<React.PropsWithChildren<unknown>> = ({ children }) => {
  const persistKey = 'csm_PageSettings';
  const persistPageSettings = JSON.parse(localStorage.getItem('csm_PageSettings') || "{}");
  const [pageSettings, dispatch] = useReducer(reducer, persistPageSettings || {})

  useEffect(() => {
    try {
      localStorage.setItem(persistKey, JSON.stringify(pageSettings))
    } catch (error) {
      console.warn(error)
    }
  }, [pageSettings])

  return (
    <Context.Provider value={{
      pageSettings,
      setPageSettings: (settings) => { dispatch({ type: 'SET', pageSettings: settings }) },
      updatePageSettings: (settings) => { dispatch({ type: 'UPDATE', pageSettings: settings }) }
    }}>
      {children}
    </Context.Provider>
  )
}

export function usePageSettings() {
  return useContext(Context);
}
