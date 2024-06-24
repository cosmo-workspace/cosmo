import { ContentCopy } from "@mui/icons-material";
import { Fab } from "@mui/material";
import { styled } from "@mui/material/styles";
import copy from "copy-to-clipboard";
import hljs from "highlight.js";
import "highlight.js/styles/default.css";
import { useSnackbar } from "notistack";
import React, { useState } from "react";

const StyledPre = styled("pre")({
  fontFamily: "Menlo, Monaco, 'Courier New', monospace",
  fontSize: 12,
  lineHeight: 1.6,
  margin: 0,
  padding: 16,
  whiteSpace: "pre",
  wordWrap: "break-word",
  overflow: "auto",
  border: "1px solid #ccc",
  borderRadius: "4px",
  backgroundColor: "#1E1E1E",
  "& .hljs-attr": {
    color: "#9CDCFE",
  },
  "& .hljs-string": {
    color: "#CE9178",
  },
  "& .hljs-number": {
    color: "#B5CEA8",
  },
  "& .hljs-literal": {
    color: "#569CD6",
  },
});

const YAMLTextArea: React.FC<{
  code: string;
}> = ({ code }) => {
  const [hover, setHover] = useState(false);
  const { enqueueSnackbar } = useSnackbar();

  const onCopy = (text: string) => {
    copy(text);
    enqueueSnackbar("Copied!", { variant: "success" });
  };

  const highlightedCode = hljs.highlight(code, {
    language: "yaml",
  }).value;

  const highlightedCodeWithSpaces = highlightedCode.replace(
    /(^|\n)( +)/g,
    function (_, newline, spaces) {
      return newline + "&nbsp;".repeat(spaces.length);
    }
  );

  return (
    <div
      onMouseOver={() => setHover(true)}
      onMouseLeave={() => setHover(false)}
    >
      <StyledPre
        dangerouslySetInnerHTML={{ __html: highlightedCodeWithSpaces }}
      />
      {hover && (
        <Fab
          color="secondary"
          aria-label="copy"
          onClick={() => {
            onCopy(code);
          }}
          size="medium"
          sx={{
            position: "absolute",
            bottom: 80,
            right: 64,
          }}
        >
          <ContentCopy fontSize="small" />
        </Fab>
      )}
    </div>
  );
};

export default YAMLTextArea;
