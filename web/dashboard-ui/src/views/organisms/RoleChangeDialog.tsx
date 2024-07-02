import { Add, ExpandLess, PersonOutlineTwoTone } from "@mui/icons-material";
import {
  Box,
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  FormHelperText,
  Grid,
  IconButton,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import React, { useState } from "react";
import { UseFormRegisterReturn, useFieldArray, useForm } from "react-hook-form";
import { DialogContext } from "../../components/ContextProvider";
import { useLogin } from "../../components/LoginProvider";
import { User } from "../../proto/gen/dashboard/v1alpha1/user_pb";
import { FormSelectableChip } from "../atoms/SelectableChips";
import { TextFieldLabel } from "../atoms/TextFieldLabel";
import { useUserModule } from "./UserModule";

const registerMui = ({ ref, ...rest }: UseFormRegisterReturn) => ({
  inputRef: ref,
  ...rest,
});

/**
 * view
 */
interface Inputs {
  roles: { name: string; enabled: boolean }[];
  customRole: string;
}

export const RoleChangeDialog: React.VFC<{
  onClose: () => void;
  user: User;
}> = ({ onClose, user }) => {
  console.log("RoleChangeDialog");
  const hooks = useUserModule();
  const { refreshUserInfo } = useLogin();

  const {
    register,
    handleSubmit,
    control,
    formState: { errors },
    getValues,
    setValue,
    setError,
  } = useForm<Inputs>({
    defaultValues: {
      roles: hooks.existingRoles.map((v) => ({
        name: v,
        enabled: user.roles.includes(v),
      })),
      customRole: "",
    },
  });

  const { fields, append } = useFieldArray({
    control,
    name: "roles",
    rules: {
      validate: (fieldArrayValues) => {
        // check that no duplicates exist
        const values = fieldArrayValues
          .map((item) => item.name)
          .filter((v) => v !== "");
        const uniqueValues = [...new Set(values)];
        return values.length === uniqueValues.length || "No duplicates allowed";
      },
    },
  });

  const [openCustomInput, setOpenCustomInput] = useState<boolean>(false);

  return (
    <Dialog open={true} fullWidth maxWidth={"xs"}>
      <DialogTitle>Change Role</DialogTitle>
      <form
        onSubmit={handleSubmit(async (inp: Inputs) => {
          console.log(inp);
          let protoRoles = inp.roles
            .filter((v) => v.enabled)
            .map((v) => v.name);
          protoRoles = [...new Set(protoRoles)]; // remove duplicates
          console.log("protoRoles", protoRoles);
          await hooks.updateRole(user.name, protoRoles);
          await refreshUserInfo();
          onClose();
        })}
        autoComplete="new-password"
      >
        <DialogContent>
          <Stack spacing={3}>
            <TextFieldLabel
              label="Name"
              fullWidth
              value={user.name}
              startAdornmentIcon={<PersonOutlineTwoTone />}
            />
            <Typography
              color="text.secondary"
              display="block"
              variant="caption"
            >
              Roles
            </Typography>
            <Grid container>
              {fields.map((v, index) => (
                <FormSelectableChip
                  defaultChecked={v.enabled}
                  key={index}
                  control={control}
                  label={v.name}
                  color="primary"
                  sx={{ m: 0.05 }}
                  {...register(`roles.${index}.enabled` as const)}
                />
              ))}
            </Grid>
            {Boolean(errors.roles?.root?.message) && (
              <FormHelperText error={Boolean(errors.roles?.root?.message)}>
                {errors.roles?.root?.message}
              </FormHelperText>
            )}
            <Box display="flex" alignItems="center">
              <Typography color="text.secondary" variant="caption">
                Add Custom Role
              </Typography>
              <IconButton
                size="small"
                onClick={() => setOpenCustomInput(!openCustomInput)}
              >
                {openCustomInput ? <ExpandLess /> : <Add />}
              </IconButton>
            </Box>
            {openCustomInput && (
              <TextField
                label="Custom Role"
                {...register(`customRole`)}
                InputProps={{
                  endAdornment: (
                    <IconButton
                      size="small"
                      onClick={() => {
                        if (!getValues(`customRole`)) return;
                        if (
                          fields
                            .map((v) => v.name)
                            .includes(getValues(`customRole`))
                        ) {
                          setError(`customRole`, {
                            message: "Role already exists",
                          });
                          return;
                        }
                        append({
                          name: getValues(`customRole`),
                          enabled: true,
                        });
                        setValue(`customRole`, "");
                        setError(`customRole`, {});
                      }}
                    >
                      <Add />
                    </IconButton>
                  ),
                }}
              />
            )}
            {Boolean(errors.customRole?.message) && (
              <FormHelperText error={Boolean(errors.customRole?.message)}>
                {errors.customRole?.message}
              </FormHelperText>
            )}
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => onClose()} color="primary">
            Cancel
          </Button>
          <Button type="submit" variant="contained" color="secondary">
            Update
          </Button>
        </DialogActions>
      </form>
    </Dialog>
  );
};

/**
 * Context
 */
export const RoleChangeDialogContext = DialogContext<{ user: User }>(
  (props) => <RoleChangeDialog {...props} />
);
