import { AddTwoTone, CheckCircleOutlined, Clear, ContentCopy, DeleteTwoTone, DoubleArrow, EditTwoTone, ErrorOutline, ExpandLessTwoTone, ExpandMoreTwoTone, LockOutlined, MoreVertTwoTone, OpenInNewTwoTone, PlayCircleFilledWhiteTwoTone, PublicOutlined, RefreshTwoTone, SearchTwoTone, StopCircleOutlined, StopCircleTwoTone, WebTwoTone } from '@mui/icons-material';
import { Alert, Avatar, Box, Card, CardContent, CardHeader, Chip, CircularProgress, Collapse, Divider, Fab, Grid, IconButton, InputAdornment, Link, ListItemIcon, ListItemText, Menu, MenuItem, Paper, Stack, TextField, Tooltip, Typography, useMediaQuery, useTheme } from '@mui/material';
import copy from 'copy-to-clipboard';
import { useSnackbar } from "notistack";
import React, { useEffect, useRef, useState } from "react";
import { useLogin } from '../../components/LoginProvider';
import { NetworkRule, Workspace } from '../../proto/gen/dashboard/v1alpha1/workspace_pb';
import { AlertTooltip } from '../atoms/AlertTooltip';
import { NameAvatar } from '../atoms/NameAvatar';
import { NetworkRuleDeleteDialogContext, NetworkRuleUpsertDialogContext } from '../organisms/NetworkRuleActionDialog';
import { hasPrivilegedRole } from '../organisms/UserModule';
import { WorkspaceCreateDialogContext, WorkspaceDeleteDialogContext, WorkspaceStartDialogContext, WorkspaceStopDialogContext } from '../organisms/WorkspaceActionDialog';
import { WorkspaceContext, WorkspaceUsersContext, computeStatus, useWorkspaceModule, useWorkspaceUsersModule } from '../organisms/WorkspaceModule';
import { PageTemplate } from '../templates/PageTemplate';

/**
 * view
 */

const StatusChip: React.VFC<{ statusLabel: string }> = ({ statusLabel }) => {
  switch (statusLabel) {
    case 'Running':
      return (<Chip variant="outlined" size='small' icon={<CheckCircleOutlined />} color='success' label={statusLabel} />);
    case 'Stopped':
      return (<Chip variant="outlined" size='small' icon={<StopCircleOutlined />} color='error' label={statusLabel} />);
    case 'Error':
    case 'CrashLoopBackOff':
      return (<Chip variant="outlined" size='small' icon={<ErrorOutline />} color='error' label={statusLabel} />);
    default:
      return (<Chip variant="outlined" size='small' icon={<CircularProgress color="info" size={13} />} color='info' label={statusLabel} />);
  }
}

const WorkspaceMenu: React.VFC<{ workspace: Workspace }> = ({ workspace }) => {
  const [anchorEl, setAnchorEl] = React.useState<null | HTMLElement>(null);
  const startDialogDispatch = WorkspaceStartDialogContext.useDispatch();
  const stopDialogDispatch = WorkspaceStopDialogContext.useDispatch();
  const deleteDialogDispatch = WorkspaceDeleteDialogContext.useDispatch();

  return (<>
    <IconButton
      color="inherit"
      onClick={(e) => { setAnchorEl(e.currentTarget); e.stopPropagation(); }}>
      <MoreVertTwoTone />
    </IconButton>
    <Menu anchorEl={anchorEl} open={Boolean(anchorEl)} onClose={() => setAnchorEl(null)}>
      <MenuItem onClick={() => { setAnchorEl(null); startDialogDispatch(true, { workspace: workspace }); }}
        disabled={!Boolean(workspace.name)}>
        <ListItemIcon><PlayCircleFilledWhiteTwoTone fontSize="small" /></ListItemIcon>
        <ListItemText>Start workspace...</ListItemText>
      </MenuItem>
      <MenuItem onClick={() => { setAnchorEl(null); stopDialogDispatch(true, { workspace: workspace }); }}
        disabled={!Boolean(workspace.name)}>
        <ListItemIcon><StopCircleTwoTone fontSize="small" /></ListItemIcon>
        <ListItemText>Stop workspace...</ListItemText>
      </MenuItem>
      <MenuItem onClick={() => { setAnchorEl(null); deleteDialogDispatch(true, { workspace: workspace }); }}
        disabled={!Boolean(workspace.name)}>
        <ListItemIcon><DeleteTwoTone fontSize="small" /></ListItemIcon>
        <ListItemText>Remove workspace...</ListItemText>
      </MenuItem>
    </Menu>
  </>);
}

const UserSelect: React.VFC = () => {
  const { user, setUser, users, getUsers } = useWorkspaceUsersModule();
  const [anchorEl, setAnchorEl] = React.useState<null | HTMLElement>(null);
  const chipReff = useRef(null);
  return (
    <>
      <Tooltip title="Change User" placement="top">
        <Chip
          ref={chipReff}
          label={user.name}
          avatar={<NameAvatar name={user.displayName} />}
          onClick={(e) => { e.stopPropagation(); getUsers().then(() => setAnchorEl(chipReff.current)); }}
          onDelete={(e) => { e.stopPropagation(); getUsers().then(() => setAnchorEl(chipReff.current)); }}
          deleteIcon={anchorEl ? <ExpandLessTwoTone /> : <ExpandMoreTwoTone />}
        />
      </Tooltip>
      <Menu anchorEl={anchorEl} open={Boolean(anchorEl)} onClose={() => setAnchorEl(null)}>
        {users.map((user, ind) =>
          <MenuItem key={ind} value={user.name} onClick={() => { setAnchorEl(null); setUser(user) }}>
            <Stack>
              <Typography>{user.name}</Typography>
              <Typography color="gray" fontSize="small"> {user.displayName}</Typography>
            </Stack>
          </MenuItem>
        )}
      </Menu>
    </>
  );
}

const NetworkRuleHeader: React.VFC<{ workspace: Workspace }> = ({ workspace }) => {
  console.log('NetworkRuleHeader');
  const Caption = (text: string) => (<Typography variant='caption' sx={{ color: 'text.secondary' }}>{text}</Typography>);
  const SubCaption = (text: string) => (<Typography variant='caption' sx={{ fontSize: 10, color: 'text.disabled' }}>{text}</Typography>);
  const upsertDialogDispatch = NetworkRuleUpsertDialogContext.useDispatch();

  const theme = useTheme();
  const isUpSM = useMediaQuery(theme.breakpoints.up('sm'), { noSsr: true });

  return (<>
    <Grid item xs={12}><Divider /></Grid>
    <Grid item xs={2} sm={1.5} md={1} sx={{ m: 'auto', textAlign: 'center' }}>{Caption('Mode')}</Grid>
    {isUpSM && <Grid item xs={4} sm={7} md={8.5} zeroMinWidth sx={{ m: 'auto' }}>{Caption('URL')}</Grid>}
    <Grid item xs={2} sm={1.5} md={1.5} sx={{ m: 'auto', textAlign: 'center' }}>{Caption('Port #')}</Grid>
    <Grid item xs={3} sm={2} md={1} sx={{ m: 'auto', textAlign: 'center' }}>
      <IconButton onClick={() => { upsertDialogDispatch(true, { workspace: workspace, index: -1 }); }}><AddTwoTone /></IconButton>
    </Grid>
    <Grid item xs={12}><Divider /></Grid>
  </>);
}

const NetworkRuleItem: React.VFC<{ workspace: Workspace, networkRule: NetworkRule, index: number, isMain?: boolean }> = ({ workspace, networkRule, index, isMain }) => {
  console.log('NetworkRuleItem');
  const { enqueueSnackbar } = useSnackbar();
  const onCopy = (text: string) => {
    copy(text);
    enqueueSnackbar('Copied!', { variant: 'success' });
  }

  const upsertDialogDispatch = NetworkRuleUpsertDialogContext.useDispatch();
  const deleteDialogDispatch = NetworkRuleDeleteDialogContext.useDispatch();

  const theme = useTheme();
  const isUpSM = useMediaQuery(theme.breakpoints.up('sm'), { noSsr: true });
  const Body2 = (text?: string | number) => (<Typography variant='body2'>{text}</Typography>);

  return (<>
    {!isUpSM &&
      <Grid item xs={12} sm={0} md={0} zeroMinWidth sx={{ m: 'auto' }}>
        <Link href={networkRule.url || ''} target='_blank'>
          <Typography variant='body2' component='span' noWrap={false} >
            {networkRule.url}
            <OpenInNewTwoTone fontSize='inherit' sx={{ position: 'relative', top: '0.2em' }} />
          </Typography>
        </Link>
      </Grid>
    }
    <Grid item xs={1} sm={1} md={1} sx={{ m: 'auto', textAlign: 'center' }}>
      {networkRule.public
        && <Tooltip title="No authentication is required for this URL"><PublicOutlined /></Tooltip>
        || <Tooltip title="Private URL"><LockOutlined /></Tooltip>}
    </Grid>

    {isUpSM &&
      <Grid item xs={2} sm={7} md={7} zeroMinWidth sx={{ m: 'auto' }}>
        <Link href={networkRule.url || ''} target='_blank'>
          <Typography variant='body2' component='span' noWrap={false} >
            {networkRule.url}
            <OpenInNewTwoTone fontSize='inherit' sx={{ position: 'relative', top: '0.2em' }} />
          </Typography>
        </Link>
        <IconButton size='small' sx={{ position: 'relative', top: '0.2em' }} onClick={() => { onCopy(networkRule.url) }}>
          <ContentCopy fontSize='inherit' />
        </IconButton>
      </Grid>
    }
    <Grid item xs={1} sm={0.5} md={1.5} sx={{ m: 'auto', textAlign: 'center' }}><DoubleArrow /></Grid>
    <Grid item xs={2} sm={1.5} md={1.5} sx={{ m: 'auto', textAlign: 'center' }}>{Body2(networkRule.portNumber)}</Grid>
    <Grid item xs={3} sm={2} md={1} sx={{ m: 'auto', textAlign: 'center' }}>
      <Stack direction='row' alignItems='center' justifyContent='center' spacing={{ xs: 0, sm: 1 }} >
        <IconButton onClick={() => { upsertDialogDispatch(true, { workspace: workspace, networkRule: networkRule, defaultOpenHttpOptions: true, index: index, isMain: isMain }); }}>
          <EditTwoTone />
        </IconButton>
        <IconButton disabled={isMain} onClick={() => { deleteDialogDispatch(true, { workspace: workspace, networkRule: networkRule, index: index }); }}>
          <DeleteTwoTone />
        </IconButton>
      </Stack>
    </Grid>
  </>);
}

const NetworkRuleList: React.VFC<{ workspace: Workspace }> = ({ workspace }) => {
  console.log('NetworkRuleList');

  return (<>
    <Grid container rowSpacing={1} columnSpacing={{ xs: 1, sm: 2, md: 2 }}>
      <Grid item xs={12}><Divider /></Grid>
      {/* network rule header */}
      <NetworkRuleHeader workspace={workspace} />
      {/* network rule detail */}
      {(workspace.spec?.network || []).map((nwRule, index) => {
        return (<NetworkRuleItem workspace={workspace} networkRule={nwRule} index={index} key={index} isMain={nwRule.url == workspace.status?.mainUrl} />)
      })}
      {(workspace.spec?.network || []).length === 0 &&
        <Grid item xs={12} sx={{ p: 2, textAlign: 'center' }}>
          <Typography variant='body2' sx={{ color: 'text.secondary' }}>No NetworkRules found.</Typography>
        </Grid>
      }
      <Grid item xs={12}><Divider /></Grid>
    </Grid>
  </>);
}


const WorkspaceItem: React.VFC<{ workspace: Workspace }> = ({ workspace: ws }) => {
  console.log("WorkspaceItem", ws.status?.phase, ws.spec?.replicas);
  const statusLabel = computeStatus(ws);

  const [expanded, setExpanded] = useState(false);

  return (
    <Grid item key={ws.name} xs={12}>
      <Card >
        <CardHeader
          onClick={() => { console.log(ws); setExpanded(!expanded) }}
          avatar={<Avatar><WebTwoTone /></Avatar>}
          title={
            ws.status && ws.status.mainUrl ? (
              <Link variant='h6' target='_blank' href={ws.status.mainUrl} onClick={(e: any) => e.stopPropagation()}>
                {ws.name} <OpenInNewTwoTone fontSize='inherit' sx={{ position: 'relative', top: '0.2em' }} />
              </Link>
            ) :
              (<Typography variant='h6'>{ws.name}</Typography>)
          }
          subheader={ws.spec && ws.spec.template}
          action={
            <Stack direction='row' spacing={2} alignItems='center'>
              <StatusChip statusLabel={statusLabel} />
              <Box onClick={(e) => e.stopPropagation()}>
                <WorkspaceMenu workspace={ws} />
              </Box>
            </Stack>
          }
        />
        <Collapse in={expanded} timeout="auto" unmountOnExit>
          <CardContent>
            <NetworkRuleList workspace={ws} />
          </CardContent>
        </Collapse>
      </Card>
    </Grid>);
}


const WorkspaceList: React.VFC = () => {
  console.log('WorkspaceList');
  const hooks = useWorkspaceModule();
  const { user } = useWorkspaceUsersModule();
  const { loginUser } = useLogin();
  const isPriv = hasPrivilegedRole(loginUser?.roles || []);
  const [searchStr, setSearchStr] = useState('');
  const [isSearchFocused, setIsSearchFocused] = useState(false);
  const [openTutorialTooltip, setOpenTutorialTooltip] = useState<boolean | undefined>(undefined);
  const createDialogDisptch = WorkspaceCreateDialogContext.useDispatch();

  useEffect(() => { hooks.getWorkspaces(user.name) }, [user]);  // eslint-disable-line

  useEffect(() => {
    if (hooks.workspaces.length === 0 && loginUser!.name === user.name) {
      // When it has never been opened
      if (openTutorialTooltip === undefined) {
        const t = setTimeout(() => setOpenTutorialTooltip(prev => prev === undefined), 5000);
        //Clean up when the watched value changes or is unmounted
        return () => clearTimeout(t);
      }
    } else if (openTutorialTooltip === true) {
      setOpenTutorialTooltip(false);
    }
  }, [hooks.workspaces.length, user.name]);// eslint-disable-line 

  const theme = useTheme();
  const isUpSM = useMediaQuery(theme.breakpoints.up('sm'), { noSsr: true });

  return (<>
    <Paper sx={{ minWidth: 320, maxWidth: 1200, mb: 1, px: 2, py: 1 }}>
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
          onFocus={() => setIsSearchFocused(true)}
          onBlur={() => setIsSearchFocused(false)}
          sx={{ flexGrow: 0.5 }}
        />
        <Box sx={{ flexGrow: 1 }} />
        {isPriv && (isUpSM || (!isSearchFocused && searchStr === "")) && <UserSelect />}
        <Tooltip title="Refresh" placement="top">
          <IconButton
            color="inherit"
            onClick={() => { hooks.refreshWorkspaces(user.name) }}>
            <RefreshTwoTone />
          </IconButton>
        </Tooltip>
        <AlertTooltip arrow placement="top"
          open={openTutorialTooltip || false}
          title={<Alert severity="info" onClick={() => { setOpenTutorialTooltip(false) }}>Create your first workspace!</Alert>} >
          <Fab size='small' color='primary' onClick={() => { setOpenTutorialTooltip(false); createDialogDisptch(true); }} sx={{ flexShrink: 0 }}>
            <AddTwoTone />
          </Fab>
        </AlertTooltip>
      </Stack>
    </Paper>
    {!hooks.workspaces.filter((ws) => searchStr === '' || Boolean(ws.name.match(searchStr))).length &&
      <Paper sx={{ minWidth: 320, maxWidth: 1200, mb: 1, p: 4 }}>
        <Typography variant='subtitle1' sx={{ color: 'text.secondary', textAlign: 'center' }}>No Workspaces found.</Typography>
      </Paper>
    }
    <Grid container spacing={1}>
      {hooks.workspaces.filter((ws) => searchStr === '' || Boolean(ws.name.match(searchStr))).map(ws =>
        <WorkspaceItem workspace={ws} key={ws.name} />
      )}
    </Grid>
  </>);
};

export const WorkspacePage: React.VFC = () => {
  console.log('WorkspacePage');

  return (
    <PageTemplate title="Workspaces">
      <div>
        <WorkspaceContext.Provider>
          <WorkspaceUsersContext.Provider>
            <WorkspaceCreateDialogContext.Provider>
              <WorkspaceStartDialogContext.Provider>
                <WorkspaceStopDialogContext.Provider>
                  <WorkspaceDeleteDialogContext.Provider>
                    <NetworkRuleUpsertDialogContext.Provider>
                      <NetworkRuleDeleteDialogContext.Provider>
                        <WorkspaceList />
                      </NetworkRuleDeleteDialogContext.Provider>
                    </NetworkRuleUpsertDialogContext.Provider>
                  </WorkspaceDeleteDialogContext.Provider>
                </WorkspaceStopDialogContext.Provider>
              </WorkspaceStartDialogContext.Provider>
            </WorkspaceCreateDialogContext.Provider>
          </WorkspaceUsersContext.Provider>
        </WorkspaceContext.Provider>
      </div>
    </PageTemplate>
  );
};
