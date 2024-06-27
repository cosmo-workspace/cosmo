import {
  AccountCircle,
  Close,
  ExpandLess,
  ExpandMore,
  OpenInNewTwoTone,
  Tune,
} from "@mui/icons-material";
import {
  Box,
  Collapse,
  Dialog,
  DialogContent,
  DialogTitle,
  IconButton,
  Link,
  List,
  ListItemButton,
  ListItemIcon,
  ListItemText,
  Stack,
  Tab,
  Tabs,
  useMediaQuery,
  useTheme,
} from "@mui/material";
import "highlight.js/styles/default.css";
import React, { useEffect, useRef, useState } from "react";
import { DialogContext } from "../../components/ContextProvider";
import { useLogin } from "../../components/LoginProvider";
import { Event } from "../../proto/gen/dashboard/v1alpha1/event_pb";
import { User, UserAddon } from "../../proto/gen/dashboard/v1alpha1/user_pb";
import { useUserService } from "../../services/DashboardServices";
import { EventsDataGrid } from "../atoms/EventsDataGrid";
import YAMLTextArea from "../atoms/YAMLTextArea";

const UserAddonListItem: React.FC<{
  addon: UserAddon;
  open: boolean;
  onClick: () => void;
}> = ({ addon, open, onClick }) => {
  return (
    <>
      <ListItemButton onClick={onClick}>
        <ListItemIcon>
          <Tune />
        </ListItemIcon>
        <ListItemText primary={addon.template} />
        {open ? <ExpandLess /> : <ExpandMore />}
      </ListItemButton>
      <Collapse in={open} timeout="auto" unmountOnExit>
        <YAMLTextArea code={addon.raw || "No yaml"}></YAMLTextArea>
      </Collapse>
    </>
  );
};

const UserInfoDialog: React.FC<{
  onClose: () => void;
  userName: string;
}> = ({ onClose, userName }) => {
  const userService = useUserService();
  const { clock, updateClock } = useLogin();

  const theme = useTheme();
  const isUpSM = useMediaQuery(theme.breakpoints.up("sm"), { noSsr: true });

  const [user, setUser] = useState<User>();
  const [events, setEvents] = useState<Event[]>([]);
  const [showTab, setShowTab] = useState<"yaml" | "events" | "addons">("yaml");

  const [openAddonItemIndex, setOpenAddonItemIndex] = useState(-1);
  const dialogContentRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    userService
      .getUser({
        userName: userName,
        withRaw: true,
      })
      .then((res) => {
        setUser(res.user);
      });

    userService
      .getEvents({
        userName: userName,
      })
      .then((res) => {
        setEvents(res.items);
      });

    updateClock();
  }, [userName]);

  return (
    <Dialog open={true} scroll="paper" fullWidth maxWidth="md">
      <DialogTitle>
        <Stack direction="row" alignItems="center" spacing={1}>
          <AccountCircle fontSize="large" />
          <span>{userName}</span>
          <Box sx={{ flexGrow: 1 }} />
          <IconButton
            sx={{
              color: (theme) => theme.palette.grey[500],
            }}
            onClick={() => onClose()}
          >
            <Close />
          </IconButton>
        </Stack>
      </DialogTitle>
      <Box sx={{ borderBottom: 1, borderColor: "divider" }}>
        <Tabs
          value={showTab}
          onChange={(_: React.SyntheticEvent, newValue: "yaml") => {
            setShowTab(newValue);
          }}
        >
          <Tab label="Live Manifest" value="yaml" />
          <Tab label="Addons" value="addons" />
          <Tab label="Events" value="events" />
        </Tabs>
      </Box>
      {showTab === "yaml" && (
        <YAMLTextArea code={user?.raw || "No yaml"}></YAMLTextArea>
      )}
      {showTab === "addons" && (
        <DialogContent ref={dialogContentRef}>
          <List>
            {user?.addons?.map((addon, index) => (
              <UserAddonListItem
                addon={addon}
                open={index === openAddonItemIndex}
                key={index}
                onClick={() => {
                  if (dialogContentRef.current) {
                    dialogContentRef.current.scrollTop = 0;
                  }
                  setOpenAddonItemIndex(
                    openAddonItemIndex === index ? -1 : index
                  );
                }}
              />
            ))}
          </List>
        </DialogContent>
      )}
      {showTab === "events" && (
        <DialogContent>
          <EventsDataGrid
            events={events}
            clock={clock}
            dataGridProps={{
              initialState: {
                sorting: {
                  sortModel: [{ field: "eventTime", sort: "desc" }],
                },
                columns: {
                  columnVisibilityModel: {
                    type: false,
                    reportingController: false,
                    series: false,
                    regardingWorkspace: false,
                    note: isUpSM,
                  },
                },
              },
            }}
          />
          <Stack direction="row" alignItems="right">
            <Box sx={{ flexGrow: 1 }} />
            <Link
              variant="body2"
              href={`#/event?user=${userName}`}
              target="_blank"
            >
              View all events...
              {
                <OpenInNewTwoTone
                  fontSize="inherit"
                  sx={{ position: "relative", top: "0.2em" }}
                />
              }
            </Link>
          </Stack>
        </DialogContent>
      )}
      <Box sx={{ borderBottom: 1, borderColor: "divider" }} />
    </Dialog>
  );
};

export const UserInfoDialogContext = DialogContext<{
  userName: string;
}>((props) => <UserInfoDialog {...props} />);
