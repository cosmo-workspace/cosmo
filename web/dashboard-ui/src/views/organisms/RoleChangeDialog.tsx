import { PersonOutlineTwoTone } from "@mui/icons-material";
import {
  Button, Dialog, DialogActions, DialogContent, DialogTitle, MenuItem, Stack, TextField
} from "@mui/material";
import React from "react";
import { useForm, UseFormRegisterReturn } from "react-hook-form";
import { DialogContext } from "../../components/ContextProvider";
import { User } from "../../proto/gen/dashboard/v1alpha1/user_pb";
import { TextFieldLabel } from "../atoms/TextFieldLabel";
import { useUserModule } from "./UserModule";

const registerMui = ({ ref, ...rest }: UseFormRegisterReturn) => ({
  inputRef: ref, ...rest
})

/**
 * view
 */
interface Inputs {
  role: string;
}

export const RoleChangeDialog: React.VFC<{ onClose: () => void, user: User }> = ({ onClose, user }) => {
  console.log('RoleChangeDialog');
  const hooks = useUserModule();

  const { register, handleSubmit, formState: { errors } } = useForm<Inputs>({
    defaultValues: { role: user.role || "" },
  });

  return (
    <Dialog open={true}
      fullWidth maxWidth={'xs'}>
      <DialogTitle>Change Role</DialogTitle>
      <form onSubmit={handleSubmit((inp: Inputs) => {
        hooks.updateRole(user.name, inp.role)
          .then(() => onClose());
      })}
        autoComplete="new-password">
        <DialogContent>
          <Stack spacing={3}>
            <TextFieldLabel label="User ID" fullWidth value={user.name} startAdornmentIcon={<PersonOutlineTwoTone />} />
            <TextField label="Role" select fullWidth defaultValue={user.role || ''}
              {...registerMui(register('role'))}
              error={Boolean(errors.role)}
              helperText={errors.role && errors.role.message}
            >
              {[
                (<MenuItem key="" value=""><em>none</em></MenuItem>),
                (<MenuItem key="cosmo-admin" value="cosmo-admin"><em>cosmo-admin</em></MenuItem>),
              ]}
            </TextField>
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => onClose()} color="primary">Cancel</Button>
          <Button type="submit" variant="contained" color="primary">Update</Button>
        </DialogActions>
      </form>
    </Dialog>
  );
};

/**
 * Context
 */
export const RoleChangeDialogContext = DialogContext<{ user: User }>(
  props => (<RoleChangeDialog {...props} />));
