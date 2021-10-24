import { CircularProgress } from '@mui/material';
import { SnackbarProvider } from 'notistack';
import React, { Suspense } from 'react';
import { HashRouter, Redirect, Route, Switch } from 'react-router-dom';
import './App.css';
import { AuthRoute } from './components/AuthRoute';
import { LoginProvider } from './components/LoginProvider';
import { MyThemeProvider } from './components/MyThemeProvider';
import { PageSettingsProvider } from './components/PageSettingsProvider';
import { ProgressProvider } from './components/ProgressProvider';
import { PasswordChangeDialogContext } from './views/organisms/PasswordChangeDialog';
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
    <Switch>
      <Route path="/signin" component={SignIn} exact />
      <AuthRoute path="/workspace" component={WorkspacePage} exact />
      <AuthRoute path="/user" component={UserPage} admin exact />
      <Route><Redirect to="/workspace" /></Route>
    </Switch>
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
                    <PasswordChangeDialogContext.Provider>
                      <SwitchApp />
                    </PasswordChangeDialogContext.Provider>
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
