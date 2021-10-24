import React from 'react';
import { cleanup, fireEvent, render } from '@testing-library/react';
import { TextFieldLabel } from '../../../views/atoms/TextFieldLabel';
import { PersonOutlineTwoTone } from '@mui/icons-material';

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