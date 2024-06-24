import {
  Tooltip,
  TooltipProps,
  Typography,
  TypographyProps,
} from "@mui/material";
import React from "react";

export type EllipsisTypographyProps = Omit<TypographyProps, "children"> & {
  children?: string;
  placement?: TooltipProps["placement"];
};

export const EllipsisTypography: React.FC<EllipsisTypographyProps> = (
  props
) => {
  const [isOverflow, setIsOverflow] = React.useState(false);
  const paragraph = React.useRef<HTMLSpanElement>(null);

  React.useEffect(() => {
    const pElement = paragraph.current;
    if (pElement) {
      setIsOverflow(Boolean(pElement.offsetWidth < pElement.scrollWidth));
    }
  }, [paragraph]);

  const title = props.children || "";

  // return isOverflow ? (<Tooltip title={title} placement={placement} > {typoglaphy} </Tooltip >) : typoglaphy
  return isOverflow ? (
    <Tooltip title={title} placement={props.placement}>
      <Typography
        ref={paragraph}
        variant="caption"
        display="block"
        sx={{
          textOverflow: "ellipsis",
          whiteSpace: "nowrap",
          overflow: "hidden",
        }}
        {...props}
      >
        {props.children}
      </Typography>
    </Tooltip>
  ) : (
    <Typography
      ref={paragraph}
      variant="caption"
      display="block"
      sx={{
        textOverflow: "ellipsis",
        whiteSpace: "nowrap",
        overflow: "hidden",
      }}
      {...props}
    >
      {props.children}
    </Typography>
  );
};
