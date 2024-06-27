import { Close, DomainVerification } from "@mui/icons-material";
import {
  Box,
  Button,
  Dialog,
  DialogActions,
  DialogTitle,
  IconButton,
  Stack,
  Tab,
  Tabs,
} from "@mui/material";
import "highlight.js/styles/default.css";
import React, { useEffect, useState } from "react";
import { DialogContext } from "../../components/ContextProvider";
import { Workspace } from "../../proto/gen/dashboard/v1alpha1/workspace_pb";
import { useWorkspaceService } from "../../services/DashboardServices";
import YAMLTextArea from "../atoms/YAMLTextArea";

const WorkspaceInfoDialog: React.FC<{
  onClose: () => void;
  ws: Workspace;
}> = ({ onClose, ws }) => {
  const wsService = useWorkspaceService();

  const [yaml, setYAML] = useState("");
  const [instance, setInstance] = useState("");
  const [showTab, setShowTab] = useState<"yaml" | "instance">("yaml");

  useEffect(() => {
    wsService
      .getWorkspace({
        wsName: ws.name,
        userName: ws.ownerName,
        withRaw: true,
      })
      .then((res) => {
        setYAML(res.workspace?.raw || "No yaml");
        setInstance(res.workspace?.rawInstance || "no instance yaml");
      })
      .catch((err) => {
        console.error(err);
      });
  }, [ws]);

  return (
    <Dialog open={true} scroll="paper" fullWidth maxWidth="md">
      <DialogTitle>
        <Stack direction="row" alignItems="center" spacing={1}>
          <DomainVerification fontSize="large" />
          <span>{ws.name}</span>
          <Box sx={{ flexGrow: 1 }} />
          <IconButton
            sx={{
              color: (theme) => theme.palette.grey[500],
            }}
            onClick={() => onClose()}
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
          <Tab label="Live Manifest" value="yaml" />
          <Tab label="Instance" value="instance" />
        </Tabs>
      </Box>
      {showTab === "yaml" && <YAMLTextArea code={yaml}></YAMLTextArea>}
      {showTab === "instance" && <YAMLTextArea code={instance}></YAMLTextArea>}
      <DialogActions>
        <Button onClick={() => onClose()} variant="contained" color="primary">
          Close
        </Button>
      </DialogActions>
    </Dialog>
  );
};

export const WorkspaceInfoDialogContext = DialogContext<{
  ws: Workspace;
}>((props) => <WorkspaceInfoDialog {...props} />);
