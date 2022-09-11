import { colors, useMediaQuery } from "@mui/material";
import { createTheme, ThemeProvider } from "@mui/material/styles";
import React from "react";

const MyTheme = () => {
  const prefersDarkMode = useMediaQuery('(prefers-color-scheme: dark)', { noSsr: true });

  return createTheme({
    components: {
      MuiOutlinedInput: {
        defaultProps: {
          notched: true,
        },
      },
      MuiInputLabel: {
        defaultProps: {
          shrink: true,
        },
      },
    },
    palette: {
      mode: prefersDarkMode ? 'dark' : undefined,
      primary: colors.deepPurple,
      secondary: colors.pink,
    },
  })
}

export const MyThemeProvider: React.FC<React.PropsWithChildren<unknown>> = ({ children }) => {

  const myTheme = MyTheme();

  return (
    <ThemeProvider theme={myTheme}>
      {children}
    </ThemeProvider>
  );
};
