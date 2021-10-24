import {
  InputAdornment, TextField, TextFieldProps
} from "@mui/material";
import React, { ReactNode } from "react";

export type TextFieldLabelProps =
  Omit<TextFieldProps, 'focused|valiant'>
  & { startAdornmentIcon?: ReactNode };

export const TextFieldLabel: React.VFC<TextFieldLabelProps> = ({ InputProps, startAdornmentIcon, ...props }) => {

  return (
    <TextField {...props} focused={false} variant="filled"
      InputProps={{
        ...InputProps,
        readOnly: true,
        inputProps: { tabIndex: -1 },
        startAdornment: startAdornmentIcon && (<InputAdornment position="start">{startAdornmentIcon}</InputAdornment>),
      }}
    />)
}
