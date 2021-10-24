import {
  darken, lighten, styled, Tooltip, tooltipClasses, TooltipProps
} from "@mui/material";
import React from "react";

export const AlertTooltip = styled(({ className, ...props }: TooltipProps) => (
  <Tooltip {...props} arrow classes={{ popper: className }} />
))(({ theme }) => ({
  [`& .${tooltipClasses.arrow}`]: {
    fontSize: '1rem',
    color: (theme.palette.mode === 'light' ? lighten : darken)(theme.palette['info'].light, 0.9),
  },
  [`& .${tooltipClasses.tooltip}`]: {
    borderRadius: "4px",
    boxShadow: theme.shadows[5],
    backgroundColor: (theme.palette.mode === 'light' ? lighten : darken)(theme.palette['info'].light, 0.9),
  },
}));
