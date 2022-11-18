import { Close } from "@mui/icons-material";
import {
  Alert, Button, Checkbox, Dialog, DialogActions, DialogContent, DialogTitle, FormControlLabel,
  IconButton, Stack, TextField, Typography
} from "@mui/material";
import { Controller, useForm, UseFormRegisterReturn } from "react-hook-form";
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
export const NetworkRuleUpsertDialog: React.VFC<{ workspace: Workspace, networkRule?: NetworkRule, onClose: () => void }>
  = ({ workspace, networkRule, onClose }) => {
    console.log('NetworkRuleUpsertDialog', networkRule);
    const networkRuleModule = useNetworkRule();
    const { register, handleSubmit, setValue, control, formState: { errors } } = useForm<NetworkRule>({
      defaultValues: networkRule || { portNumber: 0, httpPath: '/' },
    });

    const upsertRule = (newRule: NetworkRule) => {
      if (!(newRule.httpPath || '').startsWith('/')) {
        newRule.httpPath = '/' + newRule.httpPath;
      }
      networkRuleModule.upsertNetwork(workspace, newRule).then(() => onClose());
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
              {networkRule ?
                <TextFieldLabel label="Port Name" fullWidth value={networkRule.name} />
                :
                <TextField label="Port Name" fullWidth autoFocus
                  {...registerMui(register('name', {
                    required: { value: true, message: "Required" },
                    maxLength: { value: 128, message: "Max 128 characters" },
                    validate: {
                      alpha: v => /^[^a-z]+$/.test(v) === false || 'Must contain at least one letter (a-z)',
                      hyphen: v => /^[^-](.*[^-])?$/.test(v) || 'Must start and end with an alphanumeric character',
                      chars: v => /^[a-z0-9]([-a-z0-9]*[a-z0-9])?$/.test(v) || 'Only lowercase alphanumeric charactor and - are allowed',
                    },
                    onChange: e => { setValue('group', e.target.value); }
                  }))}
                  error={Boolean(errors.name)}
                  helperText={(errors.name && errors.name.message) || 'Lowercase Alphanumeric or in ["-"]'}
                />
              }
              <TextField label="Port Number" fullWidth type='number'
                {...registerMui(register('portNumber', {
                  required: { value: true, message: "Required" },
                  valueAsNumber: true,
                  min: { value: 1, message: "Min 1" },
                  max: { value: 65535, message: "Max 65535" },
                }))}
                error={Boolean(errors.portNumber)}
                helperText={(errors.portNumber && errors.portNumber.message) || '1 - 65535.'}
              />
              <TextField label="Group" fullWidth
                {...registerMui(register('group', {
                  required: { value: true, message: "Required" },
                  maxLength: { value: 128, message: "Max 128 characters" },
                  validate: {
                    hyphen: v => /^[^-](.*[^-])?$/.test(v || '') || 'Must start and end with an alphanumeric character',
                    chars: v => /^[a-z0-9]([-a-z0-9]*[a-z0-9])?$/.test(v || '') || 'Only lowercase alphanumeric charactor and - are allowed',
                  },
                }))}
                error={Boolean(errors.group)}
                helperText={(errors.group && errors.group.message)}
              />
              <TextField label="HTTP Path" fullWidth
                {...registerMui(register('httpPath', {
                  required: { value: true, message: "Required" },
                  maxLength: { value: 127, message: "Max 127 characters" },
                }))}
                error={Boolean(errors.httpPath)}
                helperText={(errors.httpPath && errors.httpPath.message)}
              />
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
  workspace: Workspace, networkRule: NetworkRule, onClose: () => void
}> = ({ workspace, networkRule, onClose }) => {
  console.log('NetworkRuleDeleteDialog', networkRule);
  const networkRuleModule = useNetworkRule();
  const deleteRule = () => {
    networkRuleModule.removeNetwork(workspace, networkRule.name).then(() => onClose());
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
          <TextFieldLabel label="Port Name" fullWidth value={networkRule.name} />
          <TextFieldLabel label="Port Number" fullWidth value={networkRule.portNumber} />
          <TextFieldLabel label="HTTP Path" fullWidth value={networkRule.httpPath} />
          <TextFieldLabel label="Group" fullWidth value={networkRule.group} />
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
export const NetworkRuleUpsertDialogContext = DialogContext<{ workspace: Workspace, networkRule?: NetworkRule }>(
  props => (<NetworkRuleUpsertDialog {...props} />));
export const NetworkRuleDeleteDialogContext = DialogContext<{ workspace: Workspace, networkRule: NetworkRule }>(
  props => (<NetworkRuleDeleteDialog {...props} />));
