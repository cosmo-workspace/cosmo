import { Close, PersonOutlineTwoTone, SecurityTwoTone, SupervisorAccountTwoTone } from "@mui/icons-material";
import {
  Alert, Button, Checkbox, Dialog, DialogActions, DialogContent, DialogTitle,
  IconButton, InputAdornment, MenuItem, Stack, TextField
} from "@mui/material";
import { useState } from "react";
import { useForm, UseFormRegisterReturn } from "react-hook-form";
import { User } from "../../api/dashboard/v1alpha1";
import { DialogContext } from "../../components/ContextProvider";
import { TextFieldLabel } from "../atoms/TextFieldLabel";
import { PasswordDialogContext } from "./PasswordDialog";
import { useUserModule } from "./UserModule";

const registerMui = ({ ref, ...rest }: UseFormRegisterReturn) => ({
  inputRef: ref, ...rest
})

/**
 * UserActionDialog
 */
interface UserActionDialogProps {
  title: string
  actions: React.ReactNode
  user: User,
  onClose: () => void,
}

const UserActionDialog: React.VFC<UserActionDialogProps> = ({ title, actions, user, onClose }) => {

  return (
    <Dialog open={true} onClose={() => onClose()} fullWidth maxWidth={'xs'}>
      <DialogTitle>{title}
        <IconButton
          sx={{ position: 'absolute', right: 8, top: 8, color: (theme) => theme.palette.grey[500] }}
          onClick={() => onClose()}>
          <Close />
        </IconButton>
      </DialogTitle>
      <DialogContent>
        <Stack spacing={3} sx={{ mt: 1 }}>
          <TextFieldLabel label="User ID" fullWidth value={user.id} startAdornmentIcon={<PersonOutlineTwoTone />} />
          <TextFieldLabel label="User Name" fullWidth value={user.displayName} startAdornmentIcon={<PersonOutlineTwoTone />} />
          <TextFieldLabel label="Role" fullWidth value={user.role} startAdornmentIcon={<SupervisorAccountTwoTone />} />
          <TextFieldLabel label="AuthType" fullWidth value={user.authType} startAdornmentIcon={<SecurityTwoTone />} />
        </Stack>
      </DialogContent>
      <DialogActions>{actions}</DialogActions>
    </Dialog>
  );
};

/**
 * Info
 */
export const UserInfoDialog: React.VFC<{ onClose: () => void, user: User }> = ({ onClose, user }) => {
  console.log('UserInfoDialog');
  return (
    <UserActionDialog
      title='User'
      onClose={() => onClose()}
      user={user}
      actions={<Button variant="contained" color="primary" onClick={() => { onClose() }}>Close</Button>} />
  );
}

/**
 * Delete
 */
export const UserDeleteDialog: React.VFC<{ onClose: () => void, user: User }> = ({ onClose, user }) => {
  console.log('UserDeleteDialog');
  const hooks = useUserModule();
  const [lock, setLock] = useState(false);

  return (
    <UserActionDialog
      title='Delete User 👋'
      onClose={() => onClose()}
      user={user}
      actions={<Alert severity='warning'
        action={<>
          <Checkbox color="warning" onChange={e => setLock(e.target.checked)} />
          <Button variant="contained" color="secondary" disabled={!lock}
            onClick={() => {
              hooks.deleteUser(user.id)
                .then(() => onClose());
            }}>Delete</Button>
        </>}
      >This action is NOT recoverable. Are you sure to delete it?</Alert>} />
  );
};

/**
 * Create
 */
interface Inputs {
  id: string;
  name: string;
  role?: string;
}
export const UserCreateDialog: React.VFC<{ onClose: () => void }> = ({ onClose }) => {
  console.log('UserCreateDialog');
  const hooks = useUserModule();
  const { register, handleSubmit, formState: { errors } } = useForm<Inputs>();
  const passwordDialogDispatch = PasswordDialogContext.useDispatch();

  return (
    <Dialog open={true}
      fullWidth maxWidth={'xs'}>
      <DialogTitle>Create New User 🎉</DialogTitle>
      <form onSubmit={handleSubmit((inp: Inputs) => {
        hooks.createUser(inp.id, inp.name, inp.role)
          .then(newUser => {
            onClose();
            passwordDialogDispatch(true, { user: newUser! });
          });
      })}
        autoComplete="new-password">
        <DialogContent>
          <Stack spacing={3}>
            <TextField label="User ID" fullWidth autoFocus
              {...registerMui(register('id', {
                required: { value: true, message: "Required" },
                pattern: { value: /^[a-z0-9]*$/, message: "Only lowercase alphanumeric characters are allowed" },
                maxLength: { value: 128, message: "Max 128 characters" },
              }))}
              error={Boolean(errors.id)}
              helperText={(errors.id && errors.id.message) || "Lowercase Alphanumeric"}
              InputProps={{
                autoComplete: "off",
                startAdornment: (<InputAdornment position="start"><PersonOutlineTwoTone /></InputAdornment>),
              }}
            />
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
            <TextField label="Role" select fullWidth defaultValue=''
              {...registerMui(register('role'))}
              error={Boolean(errors.role)} >
              {[
                (<MenuItem key="" value=""><em>none</em></MenuItem>),
                (<MenuItem key="cosmo-admin" value="cosmo-admin"><em>cosmo-admin</em></MenuItem>),
              ]}
            </TextField>
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => onClose()} color="primary">Cancel</Button>
          <Button type="submit" variant="contained" color="primary">Create</Button>
        </DialogActions>
      </form>
    </Dialog>
  );
};

/**
 * Context
 */
export const UserInfoDialogContext = DialogContext<{ user: User }>(
  props => (<UserInfoDialog {...props} />));
export const UserDeleteDialogContext = DialogContext<{ user: User }>(
  props => (<UserDeleteDialog {...props} />));
export const UserCreateDialogContext = DialogContext(
  props => (<UserCreateDialog {...props} />));
