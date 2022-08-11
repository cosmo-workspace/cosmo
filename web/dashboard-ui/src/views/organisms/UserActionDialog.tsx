import { Close, ExtensionRounded, PersonOutlineTwoTone, SecurityOutlined, SupervisorAccountTwoTone } from "@mui/icons-material";
import {
  Alert, Button, Checkbox, Collapse, Dialog, DialogActions, DialogContent, DialogTitle,
  Divider,
  FormControlLabel,
  FormHelperText,
  IconButton, InputAdornment, MenuItem, Stack, TextField, Tooltip, Typography
} from "@mui/material";
import React, { useEffect, useState } from "react";
import { useFieldArray, useForm, UseFormRegisterReturn } from "react-hook-form";
import { Template, User } from "../../api/dashboard/v1alpha1";
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
type Inputs = {
  id: string;
  name: string;
  role?: string;
  addons: {
    template: Template;
    enable: boolean;
    vars: string[];
  }[];
}
export const UserCreateDialog: React.VFC<{ onClose: () => void }> = ({ onClose }) => {
  console.log('UserCreateDialog');
  const hooks = useUserModule();
  const passwordDialogDispatch = PasswordDialogContext.useDispatch();

  const { register, handleSubmit, watch, control, formState: { errors } } = useForm<Inputs>();
  const { fields, replace } = useFieldArray({ control, name: "addons" });

  const templ = useTemplates();
  useEffect(() => { templ.getUserAddonTemplates(); }, []);  // eslint-disable-line
  useEffect(() => {
    replace(templ.templates.map(t => ({ template: t })));
  }, [templ.templates]);  // eslint-disable-line


  return (
    <Dialog open={true}
      fullWidth maxWidth={'xs'}>
      <DialogTitle>Create New User 🎉</DialogTitle>
      <form onSubmit={handleSubmit((inp: Inputs) => {
        console.log(inp)
        const userAddons = inp.addons.filter(v => v.enable)
          .map((inpAddon) => {
            const vars: { [key: string]: string; } = {};
            inpAddon.vars.forEach((v, i) => {
              vars[inpAddon.template.requiredVars?.[i].varName!] = v;
            });
            return { template: inpAddon.template.name, vars: vars, clusterScoped: inpAddon.template.isClusterScope }
          });

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
              {fields.map((field, index) =>
                <React.Fragment key={field.id}>
                  <Tooltip title={field.template.description || "No description"} placement="bottom" arrow enterDelay={1000}>
                    <>
                      <FormControlLabel label={field.template.name} control={
                        <Checkbox defaultChecked={field.template.isDefaultUserAddon || false}
                          {...registerMui(register(`addons.${index}.enable` as const, {
                            required: { value: field.template.isDefaultUserAddon || false, message: "Required" },
                          }))}
                        />}
                      />
                      <FormHelperText error={Boolean(errors.addons?.[index]?.enable)}>
                        {errors.addons?.[index]?.enable?.message}
                      </FormHelperText>
                    </>
                  </Tooltip>
                  <Collapse in={(watch('addons')[index].enable)} timeout="auto" unmountOnExit>
                    <Stack spacing={2} sx={{ m: 2 }}>
                      {field.template.requiredVars?.map((required, j) =>
                        <TextField key={field.id + j}
                          size="small" fullWidth
                          label={required.varName}
                          defaultValue={required.defaultValue}
                          {...registerMui(register(`addons.${index}.vars.${j}` as const, {
                            required: watch('addons')[index].enable,
                          }))}
                          error={Boolean(errors.addons?.[index]?.vars?.[j])}
                          helperText={errors.addons?.[index]?.vars?.[j] && "Required"}
                        />
                      )}
                    </Stack>
                  </Collapse>
                </React.Fragment>
              )}
            </Stack>
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => onClose()} color="primary">Cancel</Button>
          <Button type="submit" variant="contained" color="primary">Create</Button>
        </DialogActions>
      </form>
    </Dialog >
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
