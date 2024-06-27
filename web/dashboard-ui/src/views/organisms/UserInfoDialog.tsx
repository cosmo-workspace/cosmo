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
import { User } from "../../proto/gen/dashboard/v1alpha1/user_pb";
import { useUserService } from "../../services/DashboardServices";
import YAMLTextArea from "../atoms/YAMLTextArea";

const StyledDialogContent = styled(DialogContent)({
  overflow: "auto",
});

const UserInfoDialog: React.FC<{
  onClose: () => void;
  user: User;
}> = ({ onClose, user }) => {
  const userService = useUserService();

  const [code, setCode] = useState("");
  const [showTab, setShowTab] = useState<"objects">("objects");

  useEffect(() => {
    userService
      .getUser({
        userName: user.name,
        withRaw: true,
      })
      .then((res) => {
        setCode(res.user?.raw || "no yaml");
      });
  }, [user]);

  return (
    <Dialog open={true} scroll="paper" fullWidth maxWidth="md">
      <DialogTitle>
        <Stack direction="row" alignItems="center" spacing={1}>
          User Object
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

export const UserInfoDialogContext = DialogContext<{
  user: User;
}>((props) => <UserInfoDialog {...props} />);
