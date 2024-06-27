import { Close } from "@mui/icons-material";
import {
  Box,
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  IconButton,
  Stack,
  Tab,
  Tabs,
} from "@mui/material";
import { styled } from "@mui/material/styles";
import "highlight.js/styles/default.css";
import React, { useEffect, useState } from "react";
import { DialogContext } from "../../components/ContextProvider";
import { Workspace } from "../../proto/gen/dashboard/v1alpha1/workspace_pb";
import { useWorkspaceService } from "../../services/DashboardServices";
import YAMLTextArea from "../atoms/YAMLTextArea";

const StyledDialogContent = styled(DialogContent)({
  overflow: "auto",
});

const WorkspaceInfoDialog: React.FC<{
  onClose: () => void;
  ws: Workspace;
}> = ({ onClose, ws }) => {
  const wsService = useWorkspaceService();

  const [code, setCode] = useState("");
  const [showTab, setShowTab] = useState<"objects">("objects");

  useEffect(() => {
    wsService
      .getWorkspace({
        wsName: ws.name,
        userName: ws.ownerName,
        withRaw: true,
      })
      .then((res) => {
        setCode(res.workspace?.raw || "no yaml");
      });
  }, [ws]);

  return (
    <Dialog open={true} scroll="paper" fullWidth maxWidth="md">
      <DialogTitle>
        <Stack direction="row" alignItems="center" spacing={1}>
          Workspace Object
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

      <StyledDialogContent>
        <Box sx={{ borderBottom: 1, borderColor: "divider" }}>
          <Tabs
            value={showTab}
            onChange={(_: React.SyntheticEvent, newValue: "objects") => {
              setShowTab(newValue);
            }}
            aria-label="basic tabs example"
          >
            <Tab label="Live Manifest" value="objects" />
          </Tabs>
        </Box>
        {showTab === "objects" && <YAMLTextArea code={code}></YAMLTextArea>}
      </StyledDialogContent>
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
