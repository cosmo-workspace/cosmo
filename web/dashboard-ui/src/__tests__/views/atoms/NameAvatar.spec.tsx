import { createTheme, ThemeProvider } from "@mui/material/styles";
import { cleanup, render } from '@testing-library/react';
import React from 'react';
import { afterEach, describe, expect, it } from "vitest";
import { NameAvatar } from '../../../views/atoms/NameAvatar';

afterEach(cleanup);

describe('NameAvatar', () => {

  it('NameAvatar', () => {
    const { asFragment } = render(
      <NameAvatar />,
    );
    expect(asFragment()).toMatchSnapshot();
  });

  it('NameAvatar name', () => {
    const { asFragment } = render(
      <NameAvatar name='cosmo' />,
    );
    expect(asFragment()).toMatchSnapshot();
  });

  it('NameAvatar name dark', () => {
    const myTheme = createTheme({
      palette: { mode: 'dark' },
    });

    const { asFragment } = render(
      <ThemeProvider theme={myTheme}>
        <NameAvatar name='cosmo' />,
      </ThemeProvider>
    );
    expect(asFragment()).toMatchSnapshot();
  });
});
