import {
  CheckCircleOutlined,
  ErrorOutline,
  InfoOutlined,
  StopCircleOutlined,
} from "@mui/icons-material";
import { Chip } from "@mui/material";
import React from "react";

export const StatusChip: React.FC<{ label: string }> = ({ label }) => {
  switch (label) {
    case "Running":
      return (
        <Chip
          variant="outlined"
          size="small"
          icon={<CheckCircleOutlined />}
          color="success"
          label={label}
        />
      );
    case "Stopped":
      return (
        <Chip
          variant="outlined"
          size="small"
          icon={<StopCircleOutlined />}
          color="error"
          label={label}
        />
      );
    case "Error":
    case "CrashLoopBackOff":
      return (
        <Chip
          variant="outlined"
          size="small"
          icon={<ErrorOutline />}
          color="error"
          label={label}
        />
      );
    default:
      return (
        <Chip
          variant="outlined"
          size="small"
          icon={<InfoOutlined />}
          color="info"
          label={label}
        />
      );
  }
};
