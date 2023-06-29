import { Close, ExpandLess, ExpandMore } from "@mui/icons-material";
import {
  Alert,
  Box,
  Button, Checkbox, Collapse, Dialog, DialogActions, DialogContent, DialogTitle, Divider, FormControlLabel,
  IconButton, Stack, TextField, Typography
} from "@mui/material";
import { useState } from "react";
import { Controller, UseFormRegisterReturn, useForm } from "react-hook-form";
import { DialogContext } from "../../components/ContextProvider";
import { NetworkRule, Workspace } from "../../proto/gen/dashboard/v1alpha1/workspace_pb";
import { TextFieldLabel } from "../atoms/TextFieldLabel";
import { useNetworkRule } from "./WorkspaceModule";

const registerMui = ({ ref, ...rest }: UseFormRegisterReturn) => ({
  inputRef: ref, ...rest
})

/**
 * view
 */
export const NetworkRuleUpsertDialog: React.VFC<{ workspace: Workspace, networkRule?: NetworkRule, index: number, onClose: () => void, defaultOpenHttpOptions?: boolean, isMain?: boolean }>
  = ({ workspace, networkRule, onClose, index, defaultOpenHttpOptions, isMain }) => {
    console.log('NetworkRuleUpsertDialog', networkRule);
    const networkRuleModule = useNetworkRule();
    const { register, handleSubmit, setValue, control, formState: { errors } } = useForm<NetworkRule>({
      defaultValues: networkRule || { portNumber: 8080, httpPath: '/' },
    });

    const [openHttpOptions, setOpenHttpOptions] = useState<boolean>(defaultOpenHttpOptions || false);

    const handleOpenHttpOptionsClick = () => {
      setOpenHttpOptions(!openHttpOptions);
    };

    const upsertRule = (newRule: NetworkRule) => {
      if (!(newRule.httpPath || '').startsWith('/')) {
        newRule.httpPath = '/' + newRule.httpPath;
      }
      networkRuleModule.upsertNetwork(workspace, newRule, index).then(() => onClose());
    }

    return (
      <Dialog open={true} fullWidth maxWidth={'xs'} >
        <DialogTitle >
          {networkRule ? "Edit NetworkRule" : "Add New NetworkRule"}
          <IconButton
            sx={{ position: 'absolute', right: 8, top: 8, color: (theme) => theme.palette.grey[500] }}
            onClick={() => onClose()}>
            <Close />
          </IconButton>
        </DialogTitle>
        <DialogContent>
          <form onSubmit={handleSubmit((data) => { upsertRule(data); })}>
            <Stack sx={{ mt: 1 }} spacing={2}>
              <TextField label="Port Number" fullWidth type='number' disabled={isMain}
                {...registerMui(register('portNumber', {
                  required: { value: true, message: "Required" },
                  valueAsNumber: true,
                  min: { value: 1, message: "Min 1" },
                  max: { value: 65535, message: "Max 65535" },
                }))}
                error={Boolean(errors.portNumber)}
                helperText={(errors.portNumber && errors.portNumber.message) || '1 - 65535.'}
              />
              <Typography
                color="text.secondary"
                display="block"
                variant="caption"
              >
                HTTP Options
                <IconButton size="small" aria-label="openHttpOptions" onClick={handleOpenHttpOptionsClick}>
                  {openHttpOptions ? <ExpandLess fontSize="small" /> : <ExpandMore fontSize="small" />}
                </IconButton>
              </Typography>
              <Collapse in={openHttpOptions} timeout="auto" unmountOnExit>
                <TextField label="Custom Host Prefix" fullWidth disabled={isMain}
                  {...registerMui(register('customHostPrefix', {
                    value: networkRule?.customHostPrefix,
                    maxLength: { value: 128, message: "Max 128 characters" },
                    validate: {
                      hyphen: v => /^[^-](.*[^-])?$|^$/.test(v || '') || 'Must start and end with an alphanumeric character',
                      chars: v => /^[a-z0-9]([-a-z0-9]*[a-z0-9])?$|^$/.test(v || '') || 'Only lowercase alphanumeric charactor and - are allowed',
                    },
                  }))}
                  error={Boolean(errors.customHostPrefix)}
                  helperText={(errors.customHostPrefix && errors.customHostPrefix.message)}
                />
                <Box sx={{ margin: 2 }} />
                <TextField label="HTTP Path" fullWidth disabled={isMain}
                  {...registerMui(register('httpPath', {
                    required: { value: true, message: "Required" },
                    maxLength: { value: 127, message: "Max 127 characters" },
                  }))}
                  error={Boolean(errors.httpPath)}
                  helperText={(errors.httpPath && errors.httpPath.message)}
                />
              </Collapse>
              {isMain && <Alert severity="info" >Main Network Rule values cannot be changed</Alert>}
              <Divider />
              <FormControlLabel
                sx={{ my: 2 }}
                control={<Controller
                  name="public"
                  control={control}
                  defaultValue={false}
                  render={({ field }) => <Checkbox checked={field.value} {...field} />}
                />}
                label={<>
                  <Stack spacing={2}>
                    public
                    <Typography color="text.secondary" variant="caption" >
                      No authentication is required for this URL.
                    </Typography>
                  </Stack>
                </>}
              />
            </Stack>
            <DialogActions>
              <Button type='submit' variant="contained" color="primary">
                {!networkRule ? 'ADD' : 'UPDATE'}
              </Button>
            </DialogActions>
          </form>
        </DialogContent>
      </Dialog>
    );
  };

export const NetworkRuleDeleteDialog: React.VFC<{
  workspace: Workspace, networkRule: NetworkRule, index: number, onClose: () => void
}> = ({ workspace, networkRule, index, onClose }) => {
  console.log('NetworkRuleDeleteDialog', networkRule);
  const networkRuleModule = useNetworkRule();
  const deleteRule = () => {
    networkRuleModule.removeNetwork(workspace, index).then(() => onClose());
  }

  return (
    <Dialog open={true} fullWidth maxWidth={'xs'} >
      <DialogTitle >
        Network Rule
        <IconButton
          sx={{ position: 'absolute', right: 8, top: 8, color: (theme) => theme.palette.grey[500] }}
          onClick={() => onClose()}>
          <Close />
        </IconButton>
      </DialogTitle>
      <DialogContent>
        <Stack sx={{ mt: 1 }} spacing={2}>
          <TextFieldLabel label="Port" fullWidth value={networkRule.portNumber} />
          <Typography variant="body2">HTTP Options</Typography>
          <TextFieldLabel label="Custom Host Prefix" fullWidth value={networkRule.customHostPrefix} />
          <TextFieldLabel label="Path" fullWidth value={networkRule.httpPath} />
        </Stack>
        <Alert severity='warning' sx={{ my: 2 }}
          action={<Button variant="contained" color="secondary" onClick={deleteRule}>DELETE</Button>}
        >Are you sure to delete it?</Alert>
      </DialogContent>
    </Dialog>
  );
};

/**
 * Context
 */
export const NetworkRuleUpsertDialogContext = DialogContext<{ workspace: Workspace, networkRule?: NetworkRule, index: number, defaultOpenHttpOptions?: boolean, isMain?: boolean }>(
  props => (<NetworkRuleUpsertDialog {...props} />));
export const NetworkRuleDeleteDialogContext = DialogContext<{ workspace: Workspace, networkRule: NetworkRule, index: number }>(
  props => (<NetworkRuleDeleteDialog {...props} />));
