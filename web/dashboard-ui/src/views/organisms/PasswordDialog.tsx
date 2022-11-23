import { ContentCopy, PersonOutlineTwoTone, VpnKey } from "@mui/icons-material";
import {
  Alert, Button, Dialog, DialogActions, DialogContent, DialogTitle,
  IconButton, InputAdornment, Stack
} from "@mui/material";
import copy from 'copy-to-clipboard';
import { useSnackbar } from "notistack";
import React from "react";
import { DialogContext } from "../../components/ContextProvider";
import { User } from "../../proto/gen/dashboard/v1alpha1/user_pb";
import { TextFieldLabel } from "../atoms/TextFieldLabel";

export const PasswordDialog: React.VFC<{ onClose: () => void, user: User }> = ({ onClose, user }) => {
  console.log('PasswordDialog');
  const { enqueueSnackbar } = useSnackbar();

  const onCopy = (text: string) => {
    copy(text);
    enqueueSnackbar('Copied!', { variant: 'success' });
  }

  return (
    <Dialog open={true} fullWidth maxWidth={'xs'}>
      <DialogTitle>Here you go 🚀</DialogTitle>
      <DialogContent>
        <Stack spacing={3} sx={{ pt: 1 }}>
          <TextFieldLabel label="User ID" fullWidth value={user.name} startAdornmentIcon={<PersonOutlineTwoTone />} />
          <TextFieldLabel label="Default Password" fullWidth value={user.defaultPassword} startAdornmentIcon={<VpnKey />}
            InputProps={{
              endAdornment: (<InputAdornment position="end">
                <IconButton onClick={() => { onCopy(user.defaultPassword!) }}>
                  <ContentCopy />
                </IconButton>
              </InputAdornment>),
            }}
          />
        </Stack>
      </DialogContent>
      <DialogActions>
        <Alert severity='warning'
          action={<Button color="primary" variant="contained" onClick={() => onClose()} >OK</Button>}
        >Make sure to copy default password now. You won’t be able to see it again!</Alert>
      </DialogActions>
    </Dialog>
  );
};

/**
 * Context
 */
export const PasswordDialogContext = DialogContext<{ user: User }>(
  props => (<PasswordDialog {...props} />));
