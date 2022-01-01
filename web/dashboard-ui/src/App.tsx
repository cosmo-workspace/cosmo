import { CircularProgress } from '@mui/material';
import { SnackbarProvider } from 'notistack';
import React, { Suspense } from 'react';
import { HashRouter, Navigate, Route, Routes } from 'react-router-dom';
import './App.css';
import { AuthRoute } from './components/AuthRoute';
import { LoginProvider } from './components/LoginProvider';
import { MyThemeProvider } from './components/MyThemeProvider';
import { PageSettingsProvider } from './components/PageSettingsProvider';
import { ProgressProvider } from './components/ProgressProvider';
import { PasswordChangeDialogContext } from './views/organisms/PasswordChangeDialog';
import { UserContext } from './views/organisms/UserModule';
import { UserNameChangeDialogContext } from './views/organisms/UserNameChangeDialog';
import { SignIn } from './views/pages/SignIn';
import { UserPage } from './views/pages/UserPage';
import { WorkspacePage } from './views/pages/WorkspacePage';

// const WorkspacePage = lazy(
//   () => import('./views/pages/WorkspacePage').then(module => ({ default: module.WorkspacePage }))
// );
// const UserPage = lazy(
//   () => import('./views/pages/UserPage').then(module => ({ default: module.UserPage }))
// );

const Loading: React.VFC = () => {
  return (
    <div
      style={{
        position: 'absolute',
        top: 0,
        bottom: 0,
        left: 0,
        right: 0,
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
      }}
    >
      <CircularProgress size={50} />
    </div>
  )
}

function SwitchApp() {
  return (
    <Routes>
      <Route path="/signin" element={<SignIn />} />
      <Route path="/workspace" element={<AuthRoute><WorkspacePage /></AuthRoute>} />
      <Route path="/user" element={<AuthRoute admin><UserPage /></AuthRoute>} />
      <Route path='*' element={<Navigate to="/workspace" />} />
    </Routes>
  );
}


function App() {
  console.log('App');
  return (
    <Suspense fallback={<Loading />}>
      <div>
        <MyThemeProvider>
          <PageSettingsProvider>
            <SnackbarProvider maxSnack={3} anchorOrigin={{ vertical: 'bottom', horizontal: 'center', }}>
              <ProgressProvider>
                <HashRouter >
                  <LoginProvider>
                    <UserContext.Provider>
                      <UserNameChangeDialogContext.Provider>
                        <PasswordChangeDialogContext.Provider>
                          <SwitchApp />
                        </PasswordChangeDialogContext.Provider>
                      </UserNameChangeDialogContext.Provider>
                    </UserContext.Provider>
                  </LoginProvider>
                </HashRouter >
              </ProgressProvider>
            </SnackbarProvider>
          </PageSettingsProvider>
        </MyThemeProvider>
      </div>
    </Suspense>
  );
}

export default App;
