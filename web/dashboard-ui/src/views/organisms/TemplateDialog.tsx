import {
  Close,
  DescriptionOutlined,
  ExpandLess,
  ExpandMore,
  TextSnippetOutlined,
} from "@mui/icons-material";
import {
  Box,
  Collapse,
  Dialog,
  DialogContent,
  DialogTitle,
  IconButton,
  List,
  ListItemButton,
  ListItemIcon,
  ListItemText,
  Stack,
  Tab,
  Tabs,
  useMediaQuery,
  useTheme,
} from "@mui/material";
import "highlight.js/styles/default.css";
import React, { useRef, useState } from "react";
import { parseAllDocuments, parseDocument, stringify } from "yaml";
import { Template } from "../../proto/gen/dashboard/v1alpha1/template_pb";
import YAMLTextArea from "../atoms/YAMLTextArea";

const ObjectListItem: React.FC<{
  title: string;
  yaml: string;
  open: boolean;
  onClick: () => void;
}> = ({ title, yaml, open, onClick }) => {
  return (
    <>
      <ListItemButton onClick={onClick}>
        <ListItemIcon>
          <TextSnippetOutlined />
        </ListItemIcon>
        <ListItemText primary={title} />
        {open ? <ExpandLess /> : <ExpandMore />}
      </ListItemButton>
      <Collapse in={open} timeout="auto" unmountOnExit>
        <YAMLTextArea code={yaml || "No yaml"}></YAMLTextArea>
      </Collapse>
    </>
  );
};

export const TemplateDialog: React.FC<{
  template: Template;
  open: boolean;
  onClose: () => void;
}> = ({ template, open, onClose }) => {
  const theme = useTheme();
  const isUpSM = useMediaQuery(theme.breakpoints.up("sm"), { noSsr: true });

  const [showTab, setShowTab] = useState<"yaml" | "objects">("yaml");

  const [openAddonItemIndex, setOpenAddonItemIndex] = useState(-1);
  const dialogContentRef = useRef<HTMLDivElement>(null);

  const templateDoc = parseDocument(template?.raw || "").toJSON();

  return (
    <Dialog
      open={open}
      onClose={onClose}
      scroll="paper"
      fullWidth
      maxWidth="md"
    >
      <DialogTitle>
        <Stack direction="row" alignItems="center" spacing={1}>
          <DescriptionOutlined fontSize="large" />
          <span>{template.name}</span>
          <Box sx={{ flexGrow: 1 }} />
          <IconButton
            sx={{
              color: (theme) => theme.palette.grey[500],
            }}
            onClick={onClose}
          >
            <Close />
          </IconButton>
        </Stack>
      </DialogTitle>
      <Box sx={{ borderBottom: 1, borderColor: "divider" }}>
        <Tabs
          value={showTab}
          onChange={(_: React.SyntheticEvent, newValue: "yaml") => {
            setShowTab(newValue);
          }}
        >
          <Tab label="Live Template" value="yaml" />
          <Tab label="Manifests" value="objects" />
        </Tabs>
      </Box>
      {showTab === "yaml" && (
        <YAMLTextArea code={template?.raw || "No yaml"}></YAMLTextArea>
      )}
      {showTab === "objects" && (
        <DialogContent ref={dialogContentRef}>
          <List>
            {parseAllDocuments(
              (templateDoc["spec"]["rawYaml"] as string) || ""
            ).map((doc, index) => (
              <ObjectListItem
                title={`${doc.get("kind")}`}
                yaml={stringify(doc)}
                open={index === openAddonItemIndex}
                key={index}
                onClick={() => {
                  if (dialogContentRef.current) {
                    dialogContentRef.current.scrollTop = 0;
                  }
                  setOpenAddonItemIndex(
                    openAddonItemIndex === index ? -1 : index
                  );
                }}
              />
            ))}
          </List>
        </DialogContent>
      )}
      <Box sx={{ borderBottom: 1, borderColor: "divider" }} />
    </Dialog>
  );
};
