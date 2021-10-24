import React from 'react';
import { MyThemeProvider } from './components/MyThemeProvider';
import { SignIn } from './views/pages/SignIn';

function App() {
  return (
    <div>
      <MyThemeProvider>
        <SignIn />
      </MyThemeProvider>
    </div>
  );
}

export default App;
