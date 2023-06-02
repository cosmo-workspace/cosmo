import { AddTwoTone, Badge, Clear, DeleteTwoTone, ExpandLess, ExpandMore, ManageAccountsTwoTone, MoreVert, RefreshTwoTone, SearchTwoTone } from "@mui/icons-material";
import { Box, Card, CardHeader, Chip, Collapse, Divider, Fab, Grid, IconButton, InputAdornment, ListItemIcon, ListItemText, Menu, MenuItem, Paper, Stack, TextField, Tooltip, Typography } from "@mui/material";
import React, { useEffect, useState } from "react";
import { useLogin } from "../../components/LoginProvider";
import { User } from "../../proto/gen/dashboard/v1alpha1/user_pb";
import { NameAvatar } from "../atoms/NameAvatar";
import { SelectableChip } from "../atoms/SelectableChips";
import { PasswordDialogContext } from "../organisms/PasswordDialog";
import { RoleChangeDialogContext } from "../organisms/RoleChangeDialog";
import { UserCreateConfirmDialogContext, UserCreateDialogContext, UserDeleteDialogContext, UserInfoDialogContext } from "../organisms/UserActionDialog";
import { hasAdminForRole, hasPrivilegedRole, isAdminRole, isPrivilegedRole, useUserModule } from "../organisms/UserModule";
import { UserNameChangeDialogContext } from "../organisms/UserNameChangeDialog";
import { PageTemplate } from "../templates/PageTemplate";

/**
 * view
 */
const UserMenu: React.VFC<{ user: User }> = ({ user: us }) => {
  const { loginUser } = useLogin();
  const [anchorEl, setAnchorEl] = React.useState<null | HTMLElement>(null);
  const roleChangeDialogDispatch = RoleChangeDialogContext.useDispatch();
  const userDeleteDialogDispatch = UserDeleteDialogContext.useDispatch();
  const userNameChangeDispatch = UserNameChangeDialogContext.useDispatch();

  return (<>
    <Box>
      <IconButton
        color="inherit"
        disabled={loginUser?.name === us.name}
        onClick={e => setAnchorEl(e.currentTarget)}>
        <MoreVert fontSize="small" />
      </IconButton>
      <Menu
        anchorEl={anchorEl}
        open={Boolean(anchorEl)}
        onClose={() => setAnchorEl(null)}
      >
        <MenuItem onClick={() => {
          userNameChangeDispatch(true, { user: us });
          setAnchorEl(null);
        }}>
          <ListItemIcon><Badge fontSize="small" /></ListItemIcon>
          <ListItemText>Change Name...</ListItemText>
        </MenuItem>
        <MenuItem onClick={() => {
          roleChangeDialogDispatch(true, { user: us });
          setAnchorEl(null);
        }}>
          <ListItemIcon><ManageAccountsTwoTone fontSize="small" /></ListItemIcon>
          <ListItemText>Change Role...</ListItemText>
        </MenuItem>
        <MenuItem onClick={() => {
          userDeleteDialogDispatch(true, { user: us });
          setAnchorEl(null);
        }}>
          <ListItemIcon><DeleteTwoTone fontSize="small" /></ListItemIcon>
          <ListItemText>Remove User...</ListItemText>
        </MenuItem>
      </Menu>
    </Box>
  </>);
};

const UserList: React.VFC = () => {
  const hooks = useUserModule();
  const [searchStr, setSearchStr] = useState('');
  const userCreateDialogDispatch = UserCreateDialogContext.useDispatch();
  const userInfoDialogDispatch = UserInfoDialogContext.useDispatch();

  const [showFilter, setShowFilter] = useState<boolean>(false);

  const { loginUser } = useLogin();
  const [filterRoles, setFilterRoles] = useState<string[]>(
    // without privileged users, default filters are admin roles filters
    !loginUser ? [] : hasPrivilegedRole(loginUser.roles) ? [] : hooks.existingRoles.filter((v) => hasAdminForRole(loginUser.roles, v)));

  const pushFilterRoles = (role: string) => {
    filterRoles && setFilterRoles([...new Set([...filterRoles, role])].sort((a, b) => a < b ? -1 : 1));
  }
  const popFilterRoles = (role: string) => {
    filterRoles && setFilterRoles(filterRoles.filter(v => v !== role));
  }

  const isUserMatchedToFilterRoles = (user: User) => {
    for (const v of user.roles) {
      for (const f of filterRoles) {
        if (v === f) {
          return true
        }
      }
    }
    return false
  }

  useEffect(() => { hooks.getUsers() }, []); // eslint-disable-line

  return (<>
    <Paper sx={{ minWidth: 320, maxWidth: 1200, mb: 1, p: 1 }}>
      <Stack direction='row' alignItems='center' spacing={2}>
        <TextField
          InputProps={searchStr !== "" ? {
            startAdornment: (<InputAdornment position="start"><SearchTwoTone /></InputAdornment>),
            endAdornment: (<InputAdornment position="end">
              <IconButton size="small" tabIndex={-1} onClick={() => { setSearchStr("") }} >
                <Clear />
              </IconButton>
            </InputAdornment>)
          } : {
            startAdornment: (<InputAdornment position="start"><SearchTwoTone /></InputAdornment>),
          }}
          placeholder="Search"
          size='small'
          value={searchStr}
          onChange={e => setSearchStr(e.target.value)}
          sx={{ flexGrow: 0.5 }}
        />
        <Box sx={{ flexGrow: 1 }} />
        <Tooltip title="Refresh" placement="top">
          <IconButton color="inherit" onClick={() => { hooks.getUsers() }}>
            <RefreshTwoTone />
          </IconButton>
        </Tooltip>
        <Tooltip title="Add new user" placement="top">
          <Fab size='small' color='primary' onClick={() => userCreateDialogDispatch(true)} sx={{ flexShrink: 0 }} >
            <AddTwoTone />
          </Fab>
        </Tooltip>
      </Stack >
      <Box component="div" sx={{ justifyContent: "flex-end", textOverflow: 'ellipsis', overflow: 'hidden' }} >
        <IconButton size="small" color="inherit" onClick={() => { setShowFilter(!showFilter) }}>
          {showFilter ? < ExpandLess /> : <ExpandMore />}
        </IconButton>
        <Typography color="text.secondary" variant="caption">Filter by Roles</Typography>
        {filterRoles.length > 0 &&
          <Grid container sx={{ pt: 1 }}>
            {filterRoles.map((v, i) =>
              <SelectableChip key={v} label={v} sx={{ m: 0.1 }}
                color={isPrivilegedRole(v) ? "error" : isAdminRole(v) ? "warning" : "default"}
                defaultChecked={true} onChecked={() => { popFilterRoles(v) }} />
            )}
          </Grid>}
        <Collapse in={showFilter} timeout="auto" unmountOnExit sx={{ pt: 1 }}>
          <Divider />
          <Typography color="text.secondary" variant="caption">Existing Roles</Typography>
          <Grid container sx={{ pt: 1 }}>
            {hooks.existingRoles.map((v, i) =>
              <SelectableChip key={v} label={v} sx={{ m: 0.1 }}
                color={isPrivilegedRole(v) ? "error" : isAdminRole(v) ? "warning" : "default"}
                checked={filterRoles?.includes(v)} onChecked={(checked) => { checked ? pushFilterRoles(v) : popFilterRoles(v) }} />
            )}
          </Grid>
        </Collapse>
      </Box>

    </Paper>
    {
      !hooks.users.filter((us) => searchStr === '' || Boolean(us.name.match(searchStr))).length &&
      <Paper sx={{ minWidth: 320, maxWidth: 1200, mb: 1, p: 4 }}>
        <Typography variant='subtitle1' sx={{ color: 'text.secondary', textAlign: 'center' }}>No Users found.</Typography>
      </Paper>
    }
    <Grid container spacing={0.5}>
      {hooks.users
        .filter((us) => searchStr === '' || Boolean(us.name.match(searchStr)))
        .filter((us) => us.status === 'Active')
        .filter((us) => (filterRoles.length == 0 || isUserMatchedToFilterRoles(us)))
        .map((us) =>
          <Grid item key={us.name} xs={12} sm={6} md={6} lg={4}>
            <Card>
              <CardHeader
                avatar={<NameAvatar name={us.displayName} onClick={() => { userInfoDialogDispatch(true, { user: us }) }} />}
                title={<Stack direction='row' sx={{ mr: 2, maxWidth: 350 }}
                  onClick={() => { userInfoDialogDispatch(true, { user: us }) }}>
                  <Typography variant='subtitle1'>{us.name}</Typography>
                  <Box sx={{ flex: '1 1 auto' }} />
                  <div style={{ maxWidth: 150, whiteSpace: 'nowrap' }}>
                    <Box component="div" sx={{ justifyContent: "flex-end", textOverflow: 'ellipsis', overflow: 'hidden' }}>
                      {us.roles && us.roles.map((v, i) => {
                        return <Chip color={isPrivilegedRole(v) ? "error" : isAdminRole(v) ? "warning" : "default"} size='small' key={i} label={v} />
                      })}
                    </Box>
                  </div>
                </Stack>}
                subheader={us.displayName}
                action={<UserMenu user={us} />}
              />
            </Card>
          </Grid>
        )}
    </Grid>
  </>);
};

export const UserPage: React.VFC = () => {
  console.log('UserPage');

  return (
    <PageTemplate title="Users">
      <PasswordDialogContext.Provider>
        <UserCreateConfirmDialogContext.Provider>
          <UserCreateDialogContext.Provider>
            <RoleChangeDialogContext.Provider>
              <UserDeleteDialogContext.Provider>
                <UserInfoDialogContext.Provider>
                  <UserList />
                </UserInfoDialogContext.Provider>
              </UserDeleteDialogContext.Provider>
            </RoleChangeDialogContext.Provider>
          </UserCreateDialogContext.Provider>
        </UserCreateConfirmDialogContext.Provider>
      </PasswordDialogContext.Provider>
    </PageTemplate>
  );
}
