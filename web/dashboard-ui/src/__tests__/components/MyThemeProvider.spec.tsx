import { useTheme } from "@emotion/react";
import { useMediaQuery } from "@mui/material";
import { cleanup, renderHook } from "@testing-library/react";
import React from 'react';
import { afterEach, describe, expect, it, MockedFunction, vi } from "vitest";
import { MyThemeProvider } from '../../components/MyThemeProvider';

//--------------------------------------------------
// mock definition
//--------------------------------------------------

vi.mock('@mui/material');

//-----------------------------------------------
// test
//-----------------------------------------------

describe('AuthRoute', () => {

  const useMediaQueryMock = useMediaQuery as MockedFunction<typeof useMediaQuery>;

  afterEach(() => {
    vi.restoreAllMocks();
    cleanup();
  });


  it('normal light', async () => {

    useMediaQueryMock.mockReturnValue(false);
    const { result } = renderHook(() => useTheme(), {
      wrapper: ({ children }) => (<MyThemeProvider>{children}</MyThemeProvider >),
    });
    expect(result.current).toMatchSnapshot();

  });


  it('normal dark', async () => {

    useMediaQueryMock.mockReturnValue(true);
    const { result } = renderHook(() => useTheme(), {
      wrapper: ({ children }) => (<MyThemeProvider>{children}</MyThemeProvider >),
    });
    expect(result.current).toMatchSnapshot();

  });

});
