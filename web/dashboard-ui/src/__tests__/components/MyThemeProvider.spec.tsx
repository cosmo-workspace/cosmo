import { useTheme } from "@emotion/react";
import { useMediaQuery } from "@mui/material";
import { renderHook } from "@testing-library/react";
import React from 'react';
import { MyThemeProvider } from '../../components/MyThemeProvider';

//--------------------------------------------------
// mock definition
//--------------------------------------------------

jest.mock('@mui/material/useMediaQuery');

//-----------------------------------------------
// test
//-----------------------------------------------

describe('AuthRoute', () => {

  const useMediaQueryMock = useMediaQuery as jest.MockedFunction<typeof useMediaQuery>;

  afterEach(() => { jest.restoreAllMocks(); });


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
