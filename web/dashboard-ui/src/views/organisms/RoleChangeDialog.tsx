import { Add, PersonOutlineTwoTone, Remove } from "@mui/icons-material";
import {
  Button, Dialog, DialogActions, DialogContent, DialogTitle, IconButton, Stack, TextField
} from "@mui/material";
import React from "react";
import { useFieldArray, useForm, UseFormRegisterReturn } from "react-hook-form";
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
  roles: { name: string }[];
}

export const RoleChangeDialog: React.VFC<{ onClose: () => void, user: User }> = ({ onClose, user }) => {
  console.log('RoleChangeDialog');
  const hooks = useUserModule();

  const currentRoles = user.roles.length > 0 ? user.roles.map((v) => { return { name: v } }) : [];

  const { register, handleSubmit, control, formState: { errors } } = useForm<Inputs>({
    defaultValues: { roles: currentRoles },
  });

  const { fields: rolesFields, append: appendRoles, remove: removeRoles } = useFieldArray({ control, name: "roles" });

  return (
    <Dialog open={true}
      fullWidth maxWidth={'xs'}>
      <DialogTitle>Change Role</DialogTitle>
      <form onSubmit={handleSubmit((inp: Inputs) => {
        const protoRoles = inp.roles.filter((v) => { return v.name !== "" }).map((v) => { return v.name })
        hooks.updateRole(user.name, protoRoles)
          .then(() => onClose());
      })}
        autoComplete="new-password">
        <DialogContent>
          <Stack spacing={3}>
            <TextFieldLabel label="Name" fullWidth value={user.name} startAdornmentIcon={<PersonOutlineTwoTone />} />
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
