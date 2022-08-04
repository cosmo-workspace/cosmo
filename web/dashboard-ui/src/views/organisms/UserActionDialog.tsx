import { Close, ExtensionRounded, PersonOutlineTwoTone, SecurityOutlined, SupervisorAccountTwoTone } from "@mui/icons-material";
import {
  Alert, Button, Checkbox, Collapse, Dialog, DialogActions, DialogContent, DialogTitle,
  Divider,
  FormControlLabel,
  IconButton, InputAdornment, MenuItem, Stack, TextField, Tooltip, Typography
} from "@mui/material";
import { useEffect, useState } from "react";
import { useForm, UseFormRegisterReturn } from "react-hook-form";
import { User } from "../../api/dashboard/v1alpha1";
import { DialogContext } from "../../components/ContextProvider";
import { TextFieldLabel } from "../atoms/TextFieldLabel";
import { PasswordDialogContext } from "./PasswordDialog";
import { useTemplates, useUserModule } from "./UserModule";

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
  console.log(user)
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
          <TextFieldLabel label="AuthType" fullWidth value={user.authType} startAdornmentIcon={<SecurityOutlined />} />
          {user.addons?.map((v, i) => {
            return <TextFieldLabel label="Addons" key={i} fullWidth value={v.template} startAdornmentIcon={<ExtensionRounded />} />
          })}
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
  enableAddons: boolean[];
  addonVars: string[][];
}
export const UserCreateDialog: React.VFC<{ onClose: () => void }> = ({ onClose }) => {
  console.log('UserCreateDialog');
  const hooks = useUserModule();
  const passwordDialogDispatch = PasswordDialogContext.useDispatch();

  const { register, handleSubmit, watch, formState: { errors } } = useForm<Inputs>();
  const [isRequiredVarErrors, setIsRequiredVarErrors] = useState<Map<string, boolean>>(new Map());

  const templ = useTemplates();
  useEffect(() => { templ.getUserAddonTemplates() }, []);  // eslint-disable-line

  return (
    <Dialog open={true}
      fullWidth maxWidth={'xs'}>
      <DialogTitle>Create New User 🎉</DialogTitle>
      <form onSubmit={handleSubmit((inp: Inputs) => {

        const addons = templ.templates.map((addonTmpl, i) => {
          if (!inp.enableAddons![i]) {
            setIsRequiredVarErrors(isRequiredVarErrors.set(String(i), false));
            return { template: "" }
          }
          if (!addonTmpl.requiredVars) {
            return { template: addonTmpl.name, clusterScoped: addonTmpl.isClusterScope }
          }

          var vars: { [key: string]: string; } = {};
          var isErr = false;
          for (let j = 0; j < addonTmpl.requiredVars!.length; j++) {
            const isEmpty = !Boolean(inp.addonVars[i][j]);

            setIsRequiredVarErrors(isRequiredVarErrors.set(String(i) + String(j), isEmpty));
            if (isEmpty) { isErr = true; continue };

            vars[addonTmpl.requiredVars[j].varName!] = inp.addonVars![i]![j]!
          }
          setIsRequiredVarErrors(isRequiredVarErrors.set(String(i), isErr));
          return { template: addonTmpl.name, vars: vars, clusterScoped: addonTmpl.isClusterScope }
        });
        for (let i = 0; i < inp.enableAddons.length; i++) { if (isRequiredVarErrors.get(String(i))) return }

        const userAddons = addons.filter((v) => { return v.template !== "" });

        console.log("inp.id", inp.id, "inp.name", inp.name, "inp.role", inp.role, "userAddons", userAddons)
        hooks.createUser(inp.id, inp.name, inp.role, userAddons)
          .then(newUser => {
            onClose();
            passwordDialogDispatch(true, { user: newUser! });
            hooks.getUsers();
          });
      })}
        autoComplete="new-password">
        <DialogContent>
          <Stack spacing={3}>
            <TextField label="User ID" fullWidth autoFocus
              {...registerMui(register('id', {
                required: { value: true, message: "Required" },
                pattern: {
                  value: /^[a-z0-9]([-a-z0-9]*[a-z0-9])?$/,
                  message: 'Only lowercase alphanumeric characters or "-" are allowed'
                },
                maxLength: { value: 128, message: "Max 128 characters" },
              }))}
              error={Boolean(errors.id)}
              helperText={(errors.id && errors.id.message) || 'Lowercase Alphanumeric or "-"'}
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
            <Divider />
            <Stack spacing={1}>
              {Boolean(templ.templates.length) && <Typography
                color="text.secondary"
                display="block"
                variant="caption"
              >
                Enable User Addons
              </Typography>}
              {templ.templates.map((tmpl, i) =>
                <Stack key={tmpl.name}>
                  <Tooltip title={tmpl.description || "No description"} placement="bottom" arrow enterDelay={1000}>
                    <FormControlLabel control={
                      <Checkbox defaultChecked={Boolean(tmpl.isDefaultUserAddon)}
                        {...registerMui(register(`enableAddons.${i}`))}
                      />} label={tmpl.name} />
                  </Tooltip>

                  <Collapse in={tmpl.requiredVars && watch('enableAddons')[i]} timeout="auto" unmountOnExit>
                    <Stack spacing={2} sx={{ m: 2 }}>
                      {tmpl.requiredVars?.map((required, j) =>
                        <TextField label={required.varName} fullWidth defaultValue={required.defaultValue} key={String(i) + String(j)}
                          {...registerMui(register(`addonVars.${i}.${j}` as const))}
                          error={Boolean(isRequiredVarErrors.get(String(i) + String(j)))}
                          helperText={isRequiredVarErrors.get(String(i) + String(j)) && "Required"}
                        >
                        </TextField>
                      )}
                    </Stack>
                  </Collapse>
                </Stack>
              )}
            </Stack>
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
