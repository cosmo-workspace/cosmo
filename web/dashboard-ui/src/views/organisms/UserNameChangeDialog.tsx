import { PersonOutlineTwoTone } from "@mui/icons-material";
import {
  Button, Dialog, DialogActions, DialogContent, DialogTitle, InputAdornment, Stack, TextField
} from "@mui/material";
import React from "react";
import { useForm, UseFormRegisterReturn } from "react-hook-form";
import { DialogContext } from "../../components/ContextProvider";
import { useLogin } from "../../components/LoginProvider";
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
  name: string;
}

export const UserNameChangeDialog: React.VFC<{ onClose: () => void, user: User }> = ({ onClose, user }) => {
  console.log('UserNameChangeDialog');
  const hooks = useUserModule();
  const login = useLogin();

  const { register, handleSubmit, formState: { errors } } = useForm<Inputs>({
    defaultValues: { name: user.displayName },
  });

  const onChangeName = async (data: Inputs) => {
    console.log(hooks);
    await hooks.updateName(user.name, data.name);
    await login.refreshUserInfo();
    onClose();
  }

  return (
    <Dialog open={true}
      fullWidth maxWidth={'xs'}>
      <DialogTitle>Change Name</DialogTitle>
      <form onSubmit={handleSubmit(onChangeName)}>
        <DialogContent>
          <Stack spacing={3}>
            <TextFieldLabel label="User ID" fullWidth value={user.name} startAdornmentIcon={<PersonOutlineTwoTone />} />
            <TextField label="User Name" fullWidth
              {...registerMui(register('name', {
                required: { value: true, message: "Required" },
                maxLength: { value: 32, message: "Max 32 characters" }
              }))}
              error={Boolean(errors.name)}
              helperText={errors.name && errors.name.message}
              InputProps={{
                autoComplete: "off",
                startAdornment: (<InputAdornment position="start"><PersonOutlineTwoTone /></InputAdornment>),
              }}
            />
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => onClose()} color="primary">Cancel</Button>
          <Button type="submit" variant="contained" color="secondary">Update</Button>
        </DialogActions>
      </form>
    </Dialog>
  );
};

/**
 * Context
 */
export const UserNameChangeDialogContext = DialogContext<{ user: User }>(
  props => (<UserNameChangeDialog {...props} />));
