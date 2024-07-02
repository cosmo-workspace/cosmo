import { AccountCircle } from "@mui/icons-material";
import {
  Avatar,
  AvatarProps,
  Typography,
  TypographyProps,
  styled,
} from "@mui/material";
import React from "react";

export const NameAvatar: React.FC<
  {
    name?: string;
    typographyProps?: TypographyProps;
  } & AvatarProps
> = ({ name, typographyProps, ...props }) => {
  const bgColor = stringToColor(name || "");
  const rgb = hexToRgb(bgColor) || { r: 0, g: 0, b: 0 };
  const lum = luminance(rgb.r, rgb.g, rgb.b);

  const StyledAvator = props?.onClick
    ? styled(Avatar)({
        backgroundColor: bgColor,
        transition: "background-color 0.1s ease-in-out",
        ":hover": {
          filter: "brightness(80%)",
        },
      })
    : styled(Avatar)({
        backgroundColor: bgColor,
      });

  return name ? (
    <StyledAvator {...props}>
      <Typography
        sx={{
          color: (theme) => (lum > 0.5 ? "#000" : "#fff"),
        }}
        variant="body1"
        {...typographyProps}
      >
        {name.substring(0, 1).toUpperCase()}
      </Typography>
    </StyledAvator>
  ) : (
    <Avatar {...props}>
      <AccountCircle />
    </Avatar>
  );
};

const stringToColor = (str: string) => {
  let hash = 0;
  for (let i = 0; i < str.length; i++) {
    hash = str.charCodeAt(i) + ((hash << 5) - hash);
  }
  let color = "#";
  for (let i = 0; i < 3; i++) {
    let value = (hash >> (i * 8)) & 0xff;
    color += ("00" + value.toString(16)).substr(-2);
  }
  return color;
};

function hexToRgb(hex: string): { r: number; g: number; b: number } | null {
  const result = /^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i.exec(hex);
  return result
    ? {
        r: parseInt(result[1], 16),
        g: parseInt(result[2], 16),
        b: parseInt(result[3], 16),
      }
    : null;
}

function luminance(r: number, g: number, b: number): number {
  return (0.299 * r + 0.587 * g + 0.114 * b) / 255;
}
