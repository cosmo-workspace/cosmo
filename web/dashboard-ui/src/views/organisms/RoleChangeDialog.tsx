import { Add, PersonOutlineTwoTone, Remove } from "@mui/icons-material";
import {
  Button, Dialog, DialogActions, DialogContent, DialogTitle, FormHelperText, Grid, IconButton, Stack, TextField, Typography
} from "@mui/material";
import React from "react";
import { useFieldArray, useForm, UseFormRegisterReturn } from "react-hook-form";
import { DialogContext } from "../../components/ContextProvider";
import { User } from "../../proto/gen/dashboard/v1alpha1/user_pb";
import { FormSelectableChip } from "../atoms/SelectableChips";
import { TextFieldLabel } from "../atoms/TextFieldLabel";
import { useUserModule } from "./UserModule";

const registerMui = ({ ref, ...rest }: UseFormRegisterReturn) => ({
  inputRef: ref, ...rest
})

/**
 * view
 */
interface Inputs {
  existingRoles: { name: string, enabled: boolean }[];
  roles: { name: string }[];
}

export const RoleChangeDialog: React.VFC<{ onClose: () => void, user: User }> = ({ onClose, user }) => {
  console.log('RoleChangeDialog');
  const hooks = useUserModule();

  const { register, handleSubmit, control, formState: { errors, defaultValues } } = useForm<Inputs>({
    defaultValues: {
      existingRoles: hooks.existingRoles.map(v => ({ name: v, enabled: user.roles.includes(v) }))
    },
  });

  const { fields: rolesFields, append: appendRoles, remove: removeRoles } = useFieldArray({
    control,
    name: "roles",
    rules: {
      validate: (fieldArrayValues) => {
        // check that no duplicates exist
        let values = fieldArrayValues.map((item) => item.name).filter((v) => v !== "");
        values.push(...hooks.existingRoles);
        const uniqueValues = [...new Set(values)];
        return values.length === uniqueValues.length || "No duplicates allowed";
      }
    }
  });

  return (
    <Dialog open={true}
      fullWidth maxWidth={'xs'}>
      <DialogTitle>Change Role</DialogTitle>
      <form onSubmit={handleSubmit((inp: Inputs) => {
        console.log(inp)
        let protoRoles = inp.roles.filter((v) => { return v.name !== "" }).map((v) => { return v.name })
        console.log(protoRoles)
        inp.existingRoles.forEach((v, i) => {
          if (v.enabled) {
            protoRoles.push(v.name)
          }
        })
        protoRoles = [...new Set(protoRoles)]; // remove duplicates
        console.log("protoRoles", protoRoles)
        hooks.updateRole(user.name, protoRoles)
          .then(() => onClose());
      })}
        autoComplete="new-password">
        <DialogContent>
          <Stack spacing={3}>
            <TextFieldLabel label="Name" fullWidth value={user.name} startAdornmentIcon={<PersonOutlineTwoTone />} />
            <Typography color="text.secondary" display="block" variant="caption" >Roles</Typography>
            <Grid container>
              {hooks.existingRoles.map((v, index) =>
                <FormSelectableChip defaultChecked={defaultValues?.existingRoles && defaultValues.existingRoles[index]?.enabled} key={index} control={control} label={v} color="primary" sx={{ m: 0.05 }}
                  {...register(`existingRoles.${index}.enabled` as const)} />
              )}
            </Grid>
            {rolesFields.map((field, index) =>
              <TextField label="Role" key={index} fullWidth
                {...registerMui(register(`roles.${index}.name`, { required: { value: true, message: "Required" }, }))}
                defaultValue={field.name}
                InputProps={{
                  endAdornment: <IconButton onClick={() => { removeRoles(index) }} ><Remove /></IconButton>
                }}
                error={Boolean(errors.roles?.[index]?.name)}
                helperText={errors.roles?.[index]?.name?.message}
              />
            )}
            {Boolean(errors.roles?.root?.message) && <FormHelperText error={Boolean(errors.roles?.root?.message)}>
              {errors.roles?.root?.message}
            </FormHelperText>}
            <Button variant="outlined" onClick={() => { appendRoles({ name: '' }) }} startIcon={<Add />}>
              Add Custom Role
            </Button>
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
export const RoleChangeDialogContext = DialogContext<{ user: User }>(
  props => (<RoleChangeDialog {...props} />));
