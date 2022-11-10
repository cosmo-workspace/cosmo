import { cleanup, fireEvent, render } from '@testing-library/react';
import React from 'react';
import { afterEach, describe, expect, it } from "vitest";
import { PasswordTextField } from '../../../views/atoms/PasswordTextField';

afterEach(cleanup);


describe('PasswordTextField', () => {

  it('PasswordTextField', () => {
    const target = render(
      <PasswordTextField label="Password" margin="normal" fullWidth />,
    );
    const { asFragment, container } = target;

    expect(asFragment()).toMatchSnapshot();
    fireEvent.pointerDown(container.getElementsByClassName('MuiIconButton-root')[0]);
    expect(asFragment()).toMatchSnapshot();
    fireEvent.pointerUp(container.getElementsByClassName('MuiIconButton-root')[0]);
    expect(asFragment()).toMatchSnapshot();

    fireEvent.pointerDown(container.getElementsByClassName('MuiIconButton-root')[0]);
    expect(asFragment()).toMatchSnapshot();
    fireEvent.pointerOut(container.getElementsByClassName('MuiIconButton-root')[0]);
    expect(asFragment()).toMatchSnapshot();
  });
});