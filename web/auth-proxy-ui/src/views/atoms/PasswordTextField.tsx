import {
  IconButton,
  InputAdornment, TextField, TextFieldProps
} from "@mui/material";
import { Visibility, VisibilityOff, VpnKey } from "@mui/icons-material";
import React, { useState } from "react";


export const PasswordTextField: React.VFC<TextFieldProps> = ({ type, ...props }) => {
  const [isPassShow, setIsPassShow] = useState(false);

  return (
    <TextField {...props} type={isPassShow ? 'text' : 'password'}
      InputProps={{
        startAdornment: (<InputAdornment position="start"><VpnKey /></InputAdornment>),
        endAdornment: (<InputAdornment position="end">
          <IconButton tabIndex={-1}
            onPointerDown={() => { setIsPassShow(true) }}
            onPointerUp={() => { setIsPassShow(false) }}
            onPointerOut={() => { setIsPassShow(false) }}
          >
            {isPassShow ? <Visibility /> : <VisibilityOff />}
          </IconButton>
        </InputAdornment>),
      }}
    />
  )
}
