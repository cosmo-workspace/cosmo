import { AddTwoTone, Badge, Clear, DeleteTwoTone, ManageAccountsTwoTone, MoreVert, RefreshTwoTone, SearchTwoTone } from "@mui/icons-material";
import { Box, Card, CardHeader, Chip, Fab, Grid, IconButton, InputAdornment, ListItemIcon, ListItemText, Menu, MenuItem, Paper, Stack, TextField, Typography } from "@mui/material";
import React, { useEffect, useState } from "react";
import { useLogin } from "../../components/LoginProvider";
import { User } from "../../proto/gen/dashboard/v1alpha1/user_pb";
import { NameAvatar } from "../atoms/NameAvatar";
import { PasswordDialogContext } from "../organisms/PasswordDialog";
import { RoleChangeDialogContext } from "../organisms/RoleChangeDialog";
import { UserCreateDialogContext, UserDeleteDialogContext, UserInfoDialogContext } from "../organisms/UserActionDialog";
import { useUserModule } from "../organisms/UserModule";
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
        <IconButton color="inherit" onClick={() => { hooks.getUsers() }}>
          <RefreshTwoTone />
        </IconButton>
        <Fab size='small' color='primary' onClick={() => userCreateDialogDispatch(true)} sx={{ flexShrink: 0 }} >
          <AddTwoTone />
        </Fab>
      </Stack >
    </Paper>
    {!hooks.users.filter((us) => searchStr === '' || Boolean(us.name.match(searchStr))).length &&
      <Paper sx={{ minWidth: 320, maxWidth: 1200, mb: 1, p: 4 }}>
        <Typography variant='subtitle1' sx={{ color: 'text.secondary', textAlign: 'center' }}>No Users found.</Typography>
      </Paper>
    }
    <Grid container spacing={0.5}>
      {hooks.users
        .filter((us) => searchStr === '' || Boolean(us.name.match(searchStr)))
        .filter((us) => us.status === 'Active').map((us) =>
          <Grid item key={us.name} xs={12} sm={6} md={4}>
            <Card>
              <CardHeader
                avatar={<NameAvatar name={us.displayName} onClick={() => { userInfoDialogDispatch(true, { user: us }) }} />}
                title={<Stack direction='row' sx={{ mr: 2, maxWidth: 350 }}
                  onClick={() => { userInfoDialogDispatch(true, { user: us }) }}>
                  <Typography variant='subtitle1'>{us.name}</Typography>
                  <Box sx={{ flex: '1 1 auto' }} />
                  <div style={{ maxWidth: 150, whiteSpace: 'nowrap' }}>
                    <Box component="div" sx={{ justifyContent: "flex-end", textOverflow: 'ellipsis', overflow: 'hidden'  }}>
                      {us.roles && us.roles.map((v, i) => {
                        return <Chip size='small' key={i} label={v} />
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
        <UserCreateDialogContext.Provider>
          <RoleChangeDialogContext.Provider>
            <UserDeleteDialogContext.Provider>
              <UserInfoDialogContext.Provider>
                <UserList />
              </UserInfoDialogContext.Provider>
            </UserDeleteDialogContext.Provider>
          </RoleChangeDialogContext.Provider>
        </UserCreateDialogContext.Provider>
      </PasswordDialogContext.Provider>
    </PageTemplate>
  );
}
