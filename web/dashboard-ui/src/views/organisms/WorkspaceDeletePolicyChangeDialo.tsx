import { PersonOutlineTwoTone, WebTwoTone } from "@mui/icons-material";
import {
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  FormControlLabel,
  Radio,
  RadioGroup,
  Stack,
  Tooltip,
} from "@mui/material";
import React from "react";
import { Controller, useForm } from "react-hook-form";
import { DialogContext } from "../../components/ContextProvider";
import { DeletePolicy } from "../../proto/gen/dashboard/v1alpha1/user_pb";
import { Workspace } from "../../proto/gen/dashboard/v1alpha1/workspace_pb";
import { TextFieldLabel } from "../atoms/TextFieldLabel";
import { useWorkspaceModule } from "./WorkspaceModule";

/**
 * view
 */
interface Inputs {
  deletePolicy: string;
}

export const WorkspaceDeletePolicyChangeDialog: React.VFC<{
  onClose: () => void;
  workspace: Workspace;
}> = ({ onClose, workspace }) => {
  console.log("WorkspaceDeletePolicyChangeDialog");
  const { updateDeletePolicy } = useWorkspaceModule();
  const { control, handleSubmit } = useForm<Inputs>({});

  const onChangeDeletePolicy = async (data: Inputs) => {
    const deletePolicy =
      DeletePolicy[data.deletePolicy as keyof typeof DeletePolicy];
    console.log("deletePolicy", deletePolicy);
    await updateDeletePolicy(workspace, deletePolicy);
    onClose();
  };

  return (
    <Dialog open={true} fullWidth maxWidth={"xs"}>
      <DialogTitle>Change Delete Policy</DialogTitle>
      <form onSubmit={handleSubmit(onChangeDeletePolicy)}>
        <DialogContent>
          <Stack spacing={3}>
            <TextFieldLabel
              label="Workspace Name"
              fullWidth
              value={workspace.name}
              startAdornmentIcon={<WebTwoTone />}
            />
            <TextFieldLabel
              label="User Name"
              fullWidth
              value={workspace.ownerName}
              startAdornmentIcon={<PersonOutlineTwoTone />}
            />
            <Controller
              name="deletePolicy"
              control={control}
              defaultValue={
                DeletePolicy[workspace.deletePolicy || DeletePolicy.delete]
              }
              rules={{ required: true }}
              render={({ field }) => (
                <RadioGroup
                  {...field}
                  aria-labelledby="delete-policy-select-group"
                  name="delete-policy-select"
                  row
                  sx={{ pl: 2 }}
                >
                  <FormControlLabel
                    value="delete"
                    control={<Radio />}
                    label={
                      <Tooltip title="Enable cascading delete and allow user to delete via UI">
                        <span>Delete</span>
                      </Tooltip>
                    }
                  />
                  <FormControlLabel
                    value="keep"
                    control={<Radio />}
                    label={
                      <Tooltip title="Disable cascading delete and protect deletion via UI">
                        <span>Keep</span>
                      </Tooltip>
                    }
                  />
                </RadioGroup>
              )}
            />
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
export const WorkspaceDeletePolicyChangeDialogContext = DialogContext<{
  workspace: Workspace;
}>((props) => <WorkspaceDeletePolicyChangeDialog {...props} />);
