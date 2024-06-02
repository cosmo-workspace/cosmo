import { ContentCopy, Info, Warning } from "@mui/icons-material";
import {
  Box,
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  Divider,
  IconButton,
  InputAdornment,
  Stack,
  TextField,
  TextFieldProps,
} from "@mui/material";
import copy from "copy-to-clipboard";
import { useSnackbar } from "notistack";
import React, { useState } from "react";
import { UseFormRegisterReturn } from "react-hook-form";
import { DialogContext } from "../../components/ContextProvider";
import { Event } from "../../proto/gen/dashboard/v1alpha1/event_pb";

const registerMui = ({ ref, ...rest }: UseFormRegisterReturn) => ({
  inputRef: ref,
  ...rest,
});

const ClipboardTextField = (props: TextFieldProps) => {
  const [focused, setFocused] = useState(false);
  const { enqueueSnackbar } = useSnackbar();

  const onCopy = (text) => { // eslint-disable-line
    copy(text);
    enqueueSnackbar("Copied!", { variant: "success" });
  };

  return (
    <TextField
      {...props}
      onMouseOver={() => setFocused(true)}
      onMouseLeave={() => setFocused(false)}
      InputProps={focused
        ? {
          endAdornment: (
            <InputAdornment position="end">
              <IconButton
                onClick={() => {
                  onCopy(props.value);
                }}
              >
                <ContentCopy fontSize="small" />
              </IconButton>
            </InputAdornment>
          ),
        }
        : undefined}
    />
  );
};

export const EventDetailDialog: React.FC<
  { onClose: () => void; event: Event }
> = ({ onClose, event }) => {
  console.log("EventDetailDialog");

  return (
    <Dialog open={true} onClose={() => onClose()} fullWidth>
      <DialogTitle>
        <Box display="flex" alignItems="center">
          {event.type == "Normal"
            ? <Info sx={{ marginRight: 1 }} color="success" />
            : <Warning sx={{ marginRight: 1 }} color="warning" />}
          {event.reason}
        </Box>
      </DialogTitle>
      <DialogContent>
        <Stack spacing={2} sx={{ marginTop: 2 }}>
          <ClipboardTextField
            variant="standard"
            label="Event ID"
            value={event.id}
          />
          <Stack direction="row" spacing={2} justifyContent="space-between">
            <ClipboardTextField
              variant="standard"
              fullWidth
              label="FirstTimestamp"
              value={event.eventTime?.toDate().toLocaleString()}
            />
            <ClipboardTextField
              variant="standard"
              fullWidth
              label="Type"
              value={event.type}
            />
          </Stack>
          <Stack direction="row" spacing={2} justifyContent="space-between">
            <ClipboardTextField
              variant="standard"
              fullWidth
              label="LastTimestamp"
              value={event.series?.lastObservedTime?.toDate().toLocaleString()}
            />
            <ClipboardTextField
              variant="standard"
              fullWidth
              label="Count"
              value={event.series?.count || 1}
            />
          </Stack>
          <Stack direction="row" spacing={2}>
            <ClipboardTextField
              variant="standard"
              fullWidth
              label="Object.Kind"
              value={event.regarding?.kind}
            />
            <ClipboardTextField
              variant="standard"
              fullWidth
              label="Reporter"
              value={event.reportingController}
            />
          </Stack>
          <Stack direction="row" spacing={2}>
            <ClipboardTextField
              variant="standard"
              fullWidth
              label="Object.Name"
              value={event.regarding?.name}
            />
            <ClipboardTextField
              variant="standard"
              fullWidth
              label="Object.Namespace"
              value={event.regarding?.namespace}
            />
          </Stack>
          <ClipboardTextField
            variant="standard"
            label="Reason"
            value={event.reason}
          />
          <ClipboardTextField
            color={event.type == "Normal" ? "success" : "warning"}
            focused
            variant="standard"
            multiline
            label="Message"
            value={event.note}
          />
          <Divider />
        </Stack>
      </DialogContent>
      <DialogActions>
        <Button
          variant="contained"
          color="primary"
          onClick={() => {
            onClose();
          }}
        >
          Close
        </Button>
      </DialogActions>
    </Dialog>
  );
};

/**
 * Context
 */
export const EventDetailDialogContext = DialogContext<{ event: Event }>(
  (props) => <EventDetailDialog {...props} />,
);
