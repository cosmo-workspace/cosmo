import { PersonOutlineTwoTone } from "@mui/icons-material";
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
import { DeletePolicy, User } from "../../proto/gen/dashboard/v1alpha1/user_pb";
import { TextFieldLabel } from "../atoms/TextFieldLabel";
import { useUserModule } from "./UserModule";

/**
 * view
 */
interface Inputs {
  deletePolicy: string;
}

export const UserDeletePolicyChangeDialog: React.VFC<{
  onClose: () => void;
  user: User;
}> = ({ onClose, user }) => {
  console.log("UserDeletePolicyChangeDialog");
  const { updateDeletePolicy } = useUserModule();
  const { control, handleSubmit } = useForm<Inputs>({});

  const onChangeDeletePolicy = async (data: Inputs) => {
    const deletePolicy =
      DeletePolicy[data.deletePolicy as keyof typeof DeletePolicy];
    console.log("deletePolicy", deletePolicy);
    await updateDeletePolicy(user.name, deletePolicy);
    onClose();
  };

  return (
    <Dialog open={true} fullWidth maxWidth={"xs"}>
      <DialogTitle>Change Delete Policy</DialogTitle>
      <form onSubmit={handleSubmit(onChangeDeletePolicy)}>
        <DialogContent>
          <Stack spacing={3}>
            <TextFieldLabel
              label="User ID"
              fullWidth
              value={user.name}
              startAdornmentIcon={<PersonOutlineTwoTone />}
            />
            <Controller
              name="deletePolicy"
              control={control}
              defaultValue={
                DeletePolicy[user.deletePolicy || DeletePolicy.delete]
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
export const UserDeletePolicyChangeDialogContext = DialogContext<{
  user: User;
}>((props) => <UserDeletePolicyChangeDialog {...props} />);
