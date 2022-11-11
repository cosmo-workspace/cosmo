import { Box, createTheme, ThemeProvider } from '@mui/material';
import { cleanup, render } from '@testing-library/react';
import React from 'react';
import { afterEach, describe, expect, it } from "vitest";
import { AlertTooltip } from '../../../views/atoms/AlertTooltip';

afterEach(cleanup);


describe('AlertTooltip', () => {

  it('AlertTooltip', () => {
    const target = render(
      <AlertTooltip arrow placement="top"
        title='title1' >
        <Box />
      </AlertTooltip>
    );
    const { asFragment } = target;

    expect(asFragment()).toMatchSnapshot();
  });


  it('AlertTooltip open', () => {
    const target = render(
      <AlertTooltip arrow placement="top" open={true}
        title='title2' >
        <Box />
      </AlertTooltip>
    );
    const { asFragment } = target;

    expect(asFragment()).toMatchSnapshot();
  });


  it('AlertTooltip open dark', () => {
    const myTheme = createTheme({
      palette: { mode: 'dark' },
    });

    const target = render(
      <ThemeProvider theme={myTheme}>
        <AlertTooltip arrow placement="top" open={true} title='title3' >
          <Box />
        </AlertTooltip>
      </ThemeProvider>
    );
    const { asFragment } = target;
    expect(asFragment()).toMatchSnapshot();
  });
});
