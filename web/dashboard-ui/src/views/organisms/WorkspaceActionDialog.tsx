import { Close, PersonOutlineTwoTone, WebTwoTone } from "@mui/icons-material";
import {
  Alert,
  Button,
  Checkbox,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  IconButton,
  InputAdornment,
  MenuItem,
  Stack,
  TextField,
  Tooltip,
} from "@mui/material";
import { useEffect, useState } from "react";
import { UseFormRegisterReturn, useForm } from "react-hook-form";
import { DialogContext } from "../../components/ContextProvider";
import { Template } from "../../proto/gen/dashboard/v1alpha1/template_pb";
import { Workspace } from "../../proto/gen/dashboard/v1alpha1/workspace_pb";
import { TextFieldLabel } from "../atoms/TextFieldLabel";
import { useTemplates, useWorkspaceModule } from "./WorkspaceModule";

const registerMui = ({ ref, ...rest }: UseFormRegisterReturn) => ({
  inputRef: ref,
  ...rest,
});

/**
 * WorkspaceActionDialog
 */
const WorkspaceActionDialog: React.VFC<{
  workspace: Workspace;
  title: string;
  actions: React.ReactNode;
  onClose?: () => void;
}> = ({ workspace, onClose, title, actions }) => {
  return (
    <Dialog open={true} onClose={onClose} fullWidth maxWidth={"xs"}>
      <DialogTitle>
        {title}
        <IconButton
          sx={{
            position: "absolute",
            right: 8,
            top: 8,
            color: (theme) => theme.palette.grey[500],
          }}
          onClick={onClose}
        >
          <Close />
        </IconButton>
      </DialogTitle>
      <DialogContent>
        <Stack spacing={3} sx={{ mt: 1 }}>
          <TextFieldLabel
            label="Owner ID"
            fullWidth
            value={workspace.ownerName}
            startAdornmentIcon={<PersonOutlineTwoTone />}
          />
          <TextFieldLabel
            label="Workspace Name"
            fullWidth
            value={workspace.name}
            startAdornmentIcon={<WebTwoTone />}
          />
        </Stack>
      </DialogContent>
      <DialogActions>{actions}</DialogActions>
    </Dialog>
  );
};

/**
 * Start
 */
export const WorkspaceStartDialog: React.VFC<{
  workspace: Workspace;
  onClose: () => void;
}> = (props) => {
  console.log("WorkspaceStartDialog");
  const { workspace, onClose } = props;
  const hooks = useWorkspaceModule();
  return (
    <WorkspaceActionDialog
      {...props}
      title="Run Workspace 💡"
      actions={
        <Button
          variant="contained"
          color="secondary"
          onClick={() => {
            hooks.runWorkspace(workspace).then(() => onClose());
          }}
        >
          Start
        </Button>
      }
    />
  );
};

/**
 * Stop
 */
export const WorkspaceStopDialog: React.VFC<{
  workspace: Workspace;
  onClose: () => void;
}> = (props) => {
  console.log("WorkspaceStopDialog");
  const { workspace, onClose } = props;
  const hooks = useWorkspaceModule();
  return (
    <WorkspaceActionDialog
      {...props}
      title="Stop Workspace 💤"
      actions={
        <Button
          variant="contained"
          color="secondary"
          onClick={() => {
            hooks.stopWorkspace(workspace).then(() => onClose());
          }}
        >
          Stop
        </Button>
      }
    />
  );
};

/**
 * WorkspaceChangeDeletePolicy
 */
export const WorkspaceChangeDeletePolicyDialog: React.VFC<{
  workspace: Workspace;
  onClose: () => void;
}> = (props) => {
  console.log("WorkspaceChangeDeletePolicyDialog");
  const { workspace, onClose } = props;
  const hooks = useWorkspaceModule();
  return (
    <WorkspaceActionDialog
      {...props}
      title="Protect workspace"
      actions={
        <Button
          variant="contained"
          color="secondary"
          onClick={() => {
            hooks.stopWorkspace(workspace).then(() => onClose());
          }}
        >
          Stop
        </Button>
      }
    />
  );
};

/**
 * Delete
 */
export const WorkspaceDeleteDialog: React.VFC<{
  workspace: Workspace;
  onClose: () => void;
}> = (props) => {
  console.log("WorkspaceDeleteDialog");
  const [lock, setLock] = useState(false);
  const { workspace, onClose } = props;
  const hooks = useWorkspaceModule();
  return (
    <WorkspaceActionDialog
      {...props}
      title="Delete Workspace 💣"
      actions={
        <Alert
          severity="warning"
          action={
            <>
              <Checkbox
                color="warning"
                onChange={(e) => setLock(e.target.checked)}
              />
              <Button
                variant="contained"
                color="secondary"
                disabled={!lock}
                onClick={() => {
                  hooks.deleteWorkspace(workspace).then(() => onClose());
                }}
              >
                Delete
              </Button>
            </>
          }
        >
          This action is NOT recoverable. Are you sure to delete it?
        </Alert>
      }
    />
  );
};

/**
 * Create
 */
interface Inputs {
  wsName: string;
  templateName: string;
  vars: string[];
}
export const WorkspaceCreateDialog: React.VFC<{ onClose: () => void }> = ({
  onClose,
}) => {
  console.log("WorkspaceCreateDialog");
  const hooks = useWorkspaceModule();
  const { user } = useWorkspaceModule();
  const [template, setTemplate] = useState<Template>(new Template());
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<Inputs>();
  const { templates, getTemplates, hasRequiredAddons } = useTemplates();
  useEffect(() => {
    getTemplates({ useRoleFilter: true });
  }, []); // eslint-disable-line

  const registerTemplate = registerMui(
    register("templateName", {
      required: { value: true, message: "Required" },
    })
  );

  const isNoTemplates =
    templates.filter((v) => hasRequiredAddons(v, user)).length === 0;

  return (
    <Dialog open={true} fullWidth maxWidth={"xs"}>
      <DialogTitle>Create New Workspace 🚀</DialogTitle>
      <form
        onSubmit={handleSubmit(async (inp: Inputs) => {
          const vars: { [key: string]: string } = {};
          console.log("inp", inp);
          template.requiredVars?.forEach((rqvar, i) => {
            vars[rqvar.varName] = inp.vars[i];
          });
          hooks
            .createWorkspace(user.name, inp.wsName, inp.templateName, vars)
            .then(() => onClose());
        })}
      >
        <DialogContent>
          <Stack spacing={3}>
            <TextFieldLabel
              label="User ID"
              fullWidth
              value={user.name}
              startAdornmentIcon={<PersonOutlineTwoTone />}
            />
            <TextField
              label="Workspace Name"
              fullWidth
              autoFocus
              defaultValue=""
              {...registerMui(
                register("wsName", {
                  required: { value: true, message: "Required" },
                  pattern: {
                    value: /^[a-z0-9-]*$/,
                    message:
                      "Only lowercase alphanumeric characters and - are allowed",
                  },
                  maxLength: { value: 128, message: "Max 128 characters" },
                })
              )}
              error={Boolean(errors.wsName)}
              helperText={
                (errors.wsName && errors.wsName.message) ||
                'Lowercase Alphanumeric or in ["-", "_"]'
              }
              InputProps={{
                startAdornment: (
                  <InputAdornment position="start">
                    <WebTwoTone />
                  </InputAdornment>
                ),
              }}
            />
            <TextField
              label="Template"
              select
              fullWidth
              defaultValue=""
              {...registerTemplate}
              onChange={(e) => {
                registerTemplate.onChange(e);
                const tmpl = templates.find((v) => v.name === e.target.value);
                setTemplate(tmpl!);
              }}
              error={Boolean(errors.templateName?.message || isNoTemplates)}
              helperText={
                isNoTemplates
                  ? "No available Templates. Please contact administrators."
                  : errors.templateName?.message
              }
            >
              {templates
                .filter((v) => hasRequiredAddons(v, user))
                .map((t) => (
                  <MenuItem key={t.name} value={t.name}>
                    <Tooltip
                      title={t.description || "No description"}
                      placement="bottom"
                      arrow
                      enterDelay={1000}
                    >
                      <div>{t.name}</div>
                    </Tooltip>
                  </MenuItem>
                ))}
            </TextField>
            {template.requiredVars?.map((rqvar, index) => (
              <TextField
                label={rqvar.varName}
                fullWidth
                defaultValue={rqvar.defaultValue}
                key={index}
                {...registerMui(
                  register(`vars.${index}` as const, {
                    required: { value: true, message: "Required" },
                  })
                )}
                error={Boolean(errors.vars && errors.vars[index])}
                helperText={errors.vars && errors.vars[index]?.message}
              ></TextField>
            ))}
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => onClose()} color="primary">
            Cancel
          </Button>
          <Button type="submit" variant="contained" color="primary">
            Create
          </Button>
        </DialogActions>
      </form>
    </Dialog>
  );
};

/**
 * Context
 */
export const WorkspaceCreateDialogContext = DialogContext((props) => (
  <WorkspaceCreateDialog {...props} />
));
export const WorkspaceDeleteDialogContext = DialogContext<{
  workspace: Workspace;
}>((props) => <WorkspaceDeleteDialog {...props} />);
export const WorkspaceStartDialogContext = DialogContext<{
  workspace: Workspace;
}>((props) => <WorkspaceStartDialog {...props} />);
export const WorkspaceStopDialogContext = DialogContext<{
  workspace: Workspace;
}>((props) => <WorkspaceStopDialog {...props} />);
