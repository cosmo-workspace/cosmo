import { Button } from "@mui/material";
import { fireEvent, render, screen } from "@testing-library/react";
import React from 'react';
import { ProgressProvider, useProgress } from "../../components/ProgressProvider";

//-----------------------------------------------
// test
//-----------------------------------------------

describe('ProgressProvider', () => {

  afterEach(() => { jest.restoreAllMocks(); });

  it('normal', async () => {

    const MockView = () => {
      const { setMask, releaseMask } = useProgress();
      return (<>
        <Button data-testid="setMask" onClick={() => { setMask() }} />
        <Button data-testid="releaseMask" onClick={() => { releaseMask() }} />
      </>);
    }

    const { asFragment } = render(
      <ProgressProvider>
        <MockView />
      </ProgressProvider>
    );

    expect(asFragment()).toMatchSnapshot();
    fireEvent.click(screen.getByTestId('setMask'));
    expect(asFragment()).toMatchSnapshot();
    fireEvent.click(screen.getByTestId('releaseMask'));
    expect(asFragment()).toMatchSnapshot();
    fireEvent.click(screen.getByTestId('releaseMask'));
    expect(asFragment()).toMatchSnapshot();
  });

});
