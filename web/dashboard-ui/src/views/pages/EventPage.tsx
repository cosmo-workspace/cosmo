import styled from "@emotion/styled";
import {
  ExpandLessTwoTone,
  ExpandMoreTwoTone,
  RefreshTwoTone,
} from "@mui/icons-material";
import {
  Box,
  Chip,
  IconButton,
  Menu,
  MenuItem,
  Paper,
  Stack,
  Tooltip,
  Typography,
  useMediaQuery,
  useTheme,
} from "@mui/material";
import React, { useRef } from "react";
import { useLogin } from "../../components/LoginProvider";
import { EventsDataGrid } from "../atoms/EventsDataGrid";
import { NameAvatar } from "../atoms/NameAvatar";
import { SearchTextField } from "../atoms/SearchTextField";
import { EventDetailDialogContext } from "../organisms/EventDetailDialog";
import { EventContext, useEventModule } from "../organisms/EventModule";
import { hasPrivilegedRole } from "../organisms/UserModule";
import { PageTemplate } from "../templates/PageTemplate";

const RotatingRefreshTwoTone = styled(RefreshTwoTone)({
  animation: "rotatingRefresh 1s linear infinite",
  "@keyframes rotatingRefresh": {
    to: {
      transform: "rotate(2turn)",
    },
  },
});

const UserSelect: React.VFC = () => {
  const { user, setUser, users, getUsers } = useEventModule();
  const [anchorEl, setAnchorEl] = React.useState<null | HTMLElement>(null);
  const chipReff = useRef(null);
  return (
    <>
      <Tooltip title="Change User" placement="top">
        <Chip
          ref={chipReff}
          label={user.name}
          avatar={
            <NameAvatar
              name={user.displayName}
              typographyProps={{ variant: "body2" }}
            />
          }
          onClick={(e) => {
            e.stopPropagation();
            getUsers().then(() => setAnchorEl(chipReff.current));
          }}
          onDelete={(e) => {
            e.stopPropagation();
            getUsers().then(() => setAnchorEl(chipReff.current));
          }}
          deleteIcon={anchorEl ? <ExpandLessTwoTone /> : <ExpandMoreTwoTone />}
        />
      </Tooltip>
      <Menu
        anchorEl={anchorEl}
        open={Boolean(anchorEl)}
        onClose={() => setAnchorEl(null)}
      >
        {users.map((user, ind) => (
          <MenuItem
            key={ind}
            value={user.name}
            onClick={() => {
              setAnchorEl(null);
              setUser(user.name);
            }}
          >
            <Stack>
              <Typography>{user.name}</Typography>
              <Typography color="gray" fontSize="small">
                {" "}
                {user.displayName}
              </Typography>
            </Stack>
          </MenuItem>
        ))}
      </Menu>
    </>
  );
};

const EventList: React.VFC = () => {
  console.log("EventList");
  const { search, setSearch, events, getEvents } = useEventModule();
  const { loginUser, clock } = useLogin();
  const isPriv = hasPrivilegedRole(loginUser?.roles || []);
  const theme = useTheme();
  const isUpSM = useMediaQuery(theme.breakpoints.up("sm"), { noSsr: true });
  const [isLoading, setIsLoading] = React.useState(false);

  const searchRegExp = new RegExp(search, "i");

  return (
    <>
      <Paper sx={{ minWidth: 320, px: 2, py: 1 }}>
        <Stack direction="row" alignItems="center" spacing={2}>
          <SearchTextField search={search} setSearch={setSearch} />
          <Box sx={{ flexGrow: 1 }} />
          {isPriv && <UserSelect />}
          <Tooltip title="Refresh" placement="top">
            <IconButton
              color="inherit"
              onClick={() => {
                setIsLoading(true);
                setTimeout(() => {
                  setIsLoading(false);
                }, 1000);
                if (!isLoading) getEvents();
              }}
            >
              {isLoading ? <RotatingRefreshTwoTone /> : <RefreshTwoTone />}
            </IconButton>
          </Tooltip>
        </Stack>
      </Paper>
      <EventsDataGrid
        events={events.filter(
          (event) =>
            search === "" ||
            Boolean(event.id.match(searchRegExp)) ||
            Boolean(event.note.match(searchRegExp)) ||
            Boolean(event.reason.match(searchRegExp)) ||
            Boolean(event.reportingController.match(searchRegExp)) ||
            (event.regardingWorkspace &&
              Boolean(event.regardingWorkspace.match(searchRegExp))) ||
            Boolean(event.regarding?.kind.match(searchRegExp)) ||
            Boolean(event.regarding?.name.match(searchRegExp))
        )}
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
                regardingWorkspace: isUpSM,
                note: isUpSM,
              },
            },
          },
        }}
      />
    </>
  );
};

export const EventPage: React.VFC = () => {
  console.log("EventPage");

  return (
    <PageTemplate title="Events">
      <div>
        <EventContext.Provider>
          <EventDetailDialogContext.Provider>
            <EventList />
          </EventDetailDialogContext.Provider>
        </EventContext.Provider>
      </div>
    </PageTemplate>
  );
};
