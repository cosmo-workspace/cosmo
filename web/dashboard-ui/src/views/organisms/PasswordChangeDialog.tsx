import {
  Button, Dialog, DialogActions, DialogContent, DialogTitle, Stack
} from "@mui/material";
import { PersonOutlineTwoTone } from "@mui/icons-material";
import React, { useState } from "react";
import { useForm, UseFormRegisterReturn } from "react-hook-form";
import { DialogContext } from "../../components/ContextProvider";
import { useLogin } from "../../components/LoginProvider";
import { PasswordTextField } from "../atoms/PasswordTextField";
import { TextFieldLabel } from "../atoms/TextFieldLabel";

const registerMui = ({ ref, ...rest }: UseFormRegisterReturn) => ({
  inputRef: ref, ...rest
});

/**
 * view
 */
interface Inputs {
  currentPassword: string,
  newPassword1: string,
  newPassword2: string,
};

export const PasswordChangeDialog: React.VFC<{ onClose: () => void }> = ({ onClose }) => {
  console.log('PasswordChangeDialog');
  const { register, watch, handleSubmit, formState: { errors } } = useForm<Inputs>();
  const login = useLogin();
  const [isNewPasswordError, setIsNewPasswordError] = useState(false);

  watch(data => {
    setIsNewPasswordError(
      data.newPassword1 !== '' && data.newPassword2 !== '' &&
      data.newPassword1 !== data.newPassword2);
  });

  const onChangePass = async (data: Inputs) => {
    if (isNewPasswordError) return;
    await login.updataPassword(data.currentPassword, data.newPassword1);
    onClose();
  }

  return (
    <Dialog open={true} fullWidth maxWidth={'xs'} onClose={() => onClose()}>
      <DialogTitle>Change Password 🔒</DialogTitle>
      <form noValidate>
        <DialogContent>
          <Stack>
            <TextFieldLabel label="User ID" fullWidth value={login.loginUser?.name || ''} startAdornmentIcon={< PersonOutlineTwoTone />} />

            <PasswordTextField label="Current password" margin="normal" fullWidth autoComplete="current-password" autoFocus
              {...registerMui(register("currentPassword", {
                required: { value: true, message: "Required" },
                pattern: { value: /^[^ ]+$/, message: "Contains spaces" },
              }))}
              error={Boolean(errors.currentPassword)}
              helperText={errors.currentPassword && errors.currentPassword.message}
            />

            <PasswordTextField label="New password" margin="normal" fullWidth autoComplete="new-password"
              {...registerMui(register("newPassword1", {
                required: { value: true, message: "Required" },
                pattern: { value: /^[^ ]+$/, message: "Contains spaces" },
                minLength: { value: 6, message: "Min 6 characters" },
                maxLength: { value: 128, message: "Max 128 characters" },
              }))}
              error={Boolean(errors.newPassword1)}
              helperText={errors.newPassword1 && errors.newPassword1.message}
            />

            <PasswordTextField label="Confirm password" margin="normal" fullWidth autoComplete="new-password"
              {...registerMui(register("newPassword2", {
                required: { value: true, message: "Required" },
                pattern: { value: /^[^ ]+$/, message: "Contains spaces" },
                minLength: { value: 6, message: "Min 6 characters" },
                maxLength: { value: 128, message: "Max 128 characters" },
              }))}
              error={Boolean(errors.newPassword2) || isNewPasswordError}
              helperText={(errors.newPassword2 && errors.newPassword2.message) || (isNewPasswordError && 'Passwords do not match')}
            />
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => onClose()} sx={{ mt: 3, mb: 2 }} color="primary">Cancel</Button>
          <Button onClick={handleSubmit(onChangePass)} variant="contained" sx={{ mt: 3, mb: 2 }}>Change Password</Button>
        </DialogActions>
      </form>
    </Dialog>
  );
}

/**
 * Context
 */
export const PasswordChangeDialogContext = DialogContext(
  props => (<PasswordChangeDialog {...props} />));
