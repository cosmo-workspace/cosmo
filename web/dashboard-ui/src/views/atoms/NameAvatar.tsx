import { Avatar, AvatarProps, Typography } from "@mui/material";
import { AccountCircle } from "@mui/icons-material";
import React from "react";

export const NameAvatar: React.VFC<{ name?: string } & AvatarProps> = (props) => {
  return (
    props.name ?
      <Avatar {...props} sx={{ bgcolor: stringToColor(props.name), ...props.sx }}>
        <Typography sx={{ color: (theme) => (theme.palette.mode === 'light' ? 'white' : 'black') }} fontSize='inherit'>
          {props.name.substring(0, 1).toUpperCase()}
        </Typography>
      </Avatar>
      :
      <Avatar {...props} ><AccountCircle /></Avatar>
  )
}

const stringToColor = (str: string) => {
  let hash = 0;
  for (let i = 0; i < str.length; i++) {
    hash = str.charCodeAt(i) + ((hash << 5) - hash);
  }
  let color = '#';
  for (let i = 0; i < 3; i++) {
    let value = (hash >> (i * 8)) & 0xFF;
    color += ('00' + value.toString(16)).substr(-2);
  }
  return color;
}
