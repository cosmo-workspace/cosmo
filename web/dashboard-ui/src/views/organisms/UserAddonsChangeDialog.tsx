import { PersonOutlineTwoTone } from "@mui/icons-material";
import {
    Button, Checkbox,
    Collapse, Dialog, DialogActions, DialogContent, DialogTitle,
    FormControlLabel,
    FormHelperText,
    Stack,
    TextField, Tooltip, Typography
} from "@mui/material";
import React, { useEffect } from "react";
import { UseFormRegisterReturn, useFieldArray, useForm } from "react-hook-form";
import { DialogContext } from "../../components/ContextProvider";
import { Template } from "../../proto/gen/dashboard/v1alpha1/template_pb";
import { User, UserAddon } from "../../proto/gen/dashboard/v1alpha1/user_pb";
import { TextFieldLabel } from "../atoms/TextFieldLabel";
import { useTemplates, useUserModule } from "./UserModule";

const registerMui = ({ ref, ...rest }: UseFormRegisterReturn) => ({
    inputRef: ref, ...rest
})

/**
 * view
 */
interface Inputs {
    addons: {
        template: Template;
        enable: boolean;
        vars: string[];
    }[];
}
export const UserAddonChangeDialog: React.FC<{ onClose: () => void, user: User }> = ({ onClose, user }) => {
    console.log('UserAddonChangeDialog');
    const hooks = useUserModule();

    const { register, handleSubmit, watch, control, formState: { errors } } = useForm<Inputs>({
        defaultValues: {}
    });

    const { fields: addonsFields, replace: replaceAddons } = useFieldArray({ control, name: "addons" });

    const currentAddons = new Map<string, UserAddon>();
    user.addons.forEach((v) => {
        currentAddons.set(v.template, v)
    })

    const templ = useTemplates();
    useEffect(() => { templ.getUserAddonTemplates(); }, []);  // eslint-disable-line
    useEffect(() => {
        const tt = templ.templates.map(t => ({ template: t, enable: false, vars: [] }));
        replaceAddons(tt);
    }, [templ.templates]);  // eslint-disable-line

    return (
        <Dialog open={true}
            fullWidth maxWidth={'xs'}>
            <DialogTitle>Change UserAddons</DialogTitle>
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
                const protoUserAddons = userAddons.map(ua => new UserAddon(ua));
                console.log("protoUserAddons", protoUserAddons)

                // call API
                hooks.updateAddons(user.name, protoUserAddons)
                    .then(() => onClose());
            })}
                autoComplete="new-password">
                <DialogContent>
                    <Stack spacing={3}>
                        <TextFieldLabel label="Name" fullWidth value={user.name} startAdornmentIcon={<PersonOutlineTwoTone />} />
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
                                    <FormControlLabel label={field.template.name} control={
                                        <Tooltip title={field.template.description || "No description"} placement="bottom" arrow enterDelay={1000}>
                                            <Checkbox defaultChecked={Boolean(currentAddons.get(field.template.name)) || field.template.isDefaultUserAddon || false}
                                                {...registerMui(register(`addons.${index}.enable` as const, {
                                                    required: { value: field.template.isDefaultUserAddon || false, message: "Required" },
                                                }))}
                                            />
                                        </Tooltip>}
                                    />
                                    <FormHelperText error={Boolean(errors.addons?.[index]?.enable)}>
                                        {errors.addons?.[index]?.enable?.message}
                                    </FormHelperText>
                                    <Collapse in={(watch('addons')[index].enable)}>
                                        <Stack spacing={2}>
                                            {field.template.requiredVars?.map((required, j) =>
                                                <TextField key={field.id + j}
                                                    size="small" fullWidth
                                                    label={required.varName}
                                                    defaultValue={currentAddons.get(field.template.name)
                                                        && currentAddons.get(field.template.name)!.vars[required.varName]
                                                        || required.defaultValue}
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
                    <Button type="submit" variant="contained" color="secondary">Update</Button>
                </DialogActions>
            </form>
        </Dialog>
    );
};

/**
 * Context
 */
export const UserAddonChangeDialogContext = DialogContext<{ user: User }>(
    props => (<UserAddonChangeDialog {...props} />));