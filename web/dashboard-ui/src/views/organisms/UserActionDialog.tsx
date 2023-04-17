import { Add, Close, ExpandLess, ExpandMore, PersonOutlineTwoTone, Remove, SecurityOutlined } from "@mui/icons-material";
import {
  Alert, Button, Checkbox, Chip, Collapse, Dialog, DialogActions, DialogContent, DialogTitle,
  Divider,
  FormControlLabel,
  FormHelperText,
  Grid, IconButton, InputAdornment, List, ListItem, ListItemText, Paper, Stack, Table, TableBody, TableCell, TableContainer, TableRow, TextField, Tooltip, Typography
} from "@mui/material";
import React, { useEffect, useState } from "react";
import { useFieldArray, useForm, UseFormRegisterReturn } from "react-hook-form";
import { DialogContext } from "../../components/ContextProvider";
import { Template } from "../../proto/gen/dashboard/v1alpha1/template_pb";
import { User, UserAddons } from "../../proto/gen/dashboard/v1alpha1/user_pb";
import { NameAvatar } from "../atoms/NameAvatar";
// import { SelectableChip } from "../atoms/SelectableChip";
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

const UserActionDialog: React.VFC<UserActionDialogProps> = ({ title, actions, user, onClose, }) => {
  console.log(user)
  const [openUserAddon, setOpenUserAddon] = useState<boolean>(false);

  const handleOpenUserAddonClick = () => {
    setOpenUserAddon(!openUserAddon);
  };

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
        <Stack spacing={2}>
          <Stack alignItems="center" >
            <NameAvatar name={user?.displayName} sx={{ width: 50, height: 50 }} />
          </Stack>
          <TextFieldLabel label="Name" fullWidth value={user.name} startAdornmentIcon={<PersonOutlineTwoTone />} />
          <TextFieldLabel label="Display Name" fullWidth value={user.displayName} startAdornmentIcon={<PersonOutlineTwoTone />} />
          <TextFieldLabel label="AuthType" fullWidth value={user.authType} startAdornmentIcon={<SecurityOutlined />} />
          <Typography color="text.secondary" display="block" variant="caption" >Roles</Typography>
          <Stack alignItems="center" >
            <Grid container justifyContent="center" sx={{ width: 300 }} >
              {user?.roles && user.roles.map((v, i) => {
                return (
                  <Grid item key={i} >
                    <Chip size="small" key={i} label={v} />
                  </Grid>)
              })}
            </Grid>
          </Stack>
          <Divider />
          {Boolean(user.addons.length) && <Stack spacing={1}>
            <Typography
              color="text.secondary"
              display="block"
              variant="caption"
            >
              User Addons
              <IconButton size="small" aria-label="openUserAddon" onClick={handleOpenUserAddonClick}>
                {openUserAddon ? <ExpandLess fontSize="small" /> : <ExpandMore fontSize="small" />}
              </IconButton>
            </Typography>
            <Collapse in={openUserAddon} timeout="auto" unmountOnExit>
              <List component="nav">
                {user.addons.map((v, i) =>
                  <React.Fragment key={i}>
                    <ListItem>
                      <ListItemText
                        primary={
                          <Typography
                            color="text.secondary"
                            display="block"
                            variant="caption"
                          >* {v.template}</Typography>}
                        secondary={
                          <TableContainer component={Paper}>
                            <Table aria-label={v.template}>
                              <TableBody>
                                {Object.keys(v.vars).map((key, j) =>
                                  <TableRow key={j} sx={{ '&:last-child td, &:last-child th': { border: 0 } }} >
                                    <TableCell component="th" scope="row">{key}</TableCell>
                                    <TableCell align="right">{v.vars[key]}</TableCell>
                                  </TableRow>
                                )}
                              </TableBody>
                            </Table>
                          </TableContainer>
                        }
                      />
                    </ListItem>
                  </React.Fragment>
                )}
              </List>
            </Collapse>
          </Stack>}
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
      title='User Info'
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
              hooks.deleteUser(user.name)
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
  definedRoles: { enabled: boolean }[];
  isCosmoAdmin: boolean;
  roles: { name: string }[];
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

  const { register, handleSubmit, watch, control, formState: { errors, dirtyFields } } = useForm<Inputs>({
    defaultValues: {}
  });

  const { fields: addonsFields, replace: replaceAddons } = useFieldArray({ control, name: "addons" });

  const templ = useTemplates();
  useEffect(() => { templ.getUserAddonTemplates(); }, []);  // eslint-disable-line
  useEffect(() => {
    replaceAddons(templ.templates.map(t => ({ template: t, enable: false, vars: [] })));
  }, [templ.templates]);  // eslint-disable-line

  const { fields: rolesFields, append: appendRoles, remove: removeRoles } = useFieldArray({ control, name: "roles" });

  const definedRoles = ['cosmo-admin'];

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

        console.log("inp.id", inp.id, "inp.name", inp.name, "inp.roles", inp.roles, "userAddons", userAddons)
        const protoUserAddons = userAddons.map(ua => new UserAddons(ua));
        const protoRoles = inp.roles.filter((v) => { return v.name !== "" }).map((v) => { return v.name })
        hooks.createUser(inp.id, inp.name, protoRoles, protoUserAddons)
          .then(newUser => {
            onClose();
            passwordDialogDispatch(true, { user: newUser! });
            hooks.getUsers();
          });
      })}
        autoComplete="new-password">
        <DialogContent>
          <Stack spacing={2}>
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
            {/* <Typography color="text.secondary" display="block" variant="caption" >Roles</Typography>
            <Stack alignItems="center" border="solid lightgrey 1px" borderRadius={1} p={1} >
              {definedRoles.map((label, index) =>
                <Grid container justifyContent="center" sx={{ width: 300 }} key={index} >
                  <SelectableChip label={label} color="primary" {...registerMui(register(`isCosmoAdmin` as const))} />
                </Grid>
              )}
            </Stack> */}
            {rolesFields.map((field, index) =>
              <TextField label="Role" key={index} fullWidth
                {...registerMui(register(`roles.${index}.name`))}
                defaultValue={field.name}
                InputProps={{
                  endAdornment: <IconButton onClick={() => { removeRoles(index) }} ><Remove /></IconButton>
                }}
              />
            )}
            <Button variant="outlined" onClick={() => { appendRoles({ name: '' }) }} startIcon={<Add />}>
              Add Role
            </Button>
            <Divider />
            <Stack spacing={1}>
              {Boolean(templ.templates.length) && <Typography
                color="text.secondary"
                display="block"
                variant="caption"
              >
                Enable User Addons
              </Typography>}
              {addonsFields.map((field, index) =>
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
