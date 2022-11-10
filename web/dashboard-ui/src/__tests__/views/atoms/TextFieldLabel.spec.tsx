import { PersonOutlineTwoTone } from '@mui/icons-material';
import { cleanup, render } from '@testing-library/react';
import React from 'react';
import { afterEach, describe, expect, it } from "vitest";
import { TextFieldLabel } from '../../../views/atoms/TextFieldLabel';

afterEach(cleanup);


describe('TextFieldLabel', () => {

  it('TextFieldLabel', () => {
    const { asFragment } = render(
      <TextFieldLabel label='label1' fullWidth value='value1' />,
    );
    expect(asFragment()).toMatchSnapshot();
  });

  it('TextFieldLabel startAdornmentIcon', () => {
    const { asFragment } = render(
      <TextFieldLabel label='label1' fullWidth value='value1' startAdornmentIcon={<PersonOutlineTwoTone />} />,
    );
    expect(asFragment()).toMatchSnapshot();
  });

});