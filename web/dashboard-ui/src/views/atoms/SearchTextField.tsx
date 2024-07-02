import { Clear, SearchTwoTone } from "@mui/icons-material";
import {
  IconButton,
  InputAdornment,
  TextField,
  TextFieldProps,
} from "@mui/material";
import React from "react";

export const SearchTextField: React.FC<
  {
    search: string;
    setSearch: (search: string) => void;
  } & TextFieldProps
> = ({ search, setSearch, ...props }) => {
  return (
    <TextField
      InputProps={
        search !== ""
          ? {
              startAdornment: (
                <InputAdornment position="start">
                  <SearchTwoTone />
                </InputAdornment>
              ),
              endAdornment: (
                <InputAdornment position="end">
                  <IconButton
                    size="small"
                    tabIndex={-1}
                    onClick={() => {
                      setSearch("");
                    }}
                  >
                    <Clear />
                  </IconButton>
                </InputAdornment>
              ),
            }
          : {
              startAdornment: (
                <InputAdornment position="start">
                  <SearchTwoTone />
                </InputAdornment>
              ),
            }
      }
      placeholder="Search"
      size="small"
      value={search}
      onChange={(e) => setSearch(e.target.value)}
      sx={{ flexGrow: 0.5 }}
      {...props}
    />
  );
};
