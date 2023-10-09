import { Delete, VerifiedOutlined } from "@mui/icons-material";
import {
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  Divider,
  Grid,
  IconButton,
  Tooltip,
  Typography,
  useMediaQuery, useTheme
} from "@mui/material";
import Box from '@mui/material/Box';
import { useSnackbar } from 'notistack';
import { useEffect, useState } from "react";
import { base64url } from '../../components/Base64';
import { DialogContext } from "../../components/ContextProvider";
import { User } from "../../proto/gen/dashboard/v1alpha1/user_pb";
import { Credential } from "../../proto/gen/dashboard/v1alpha1/webauthn_pb";
import { useWebAuthnService } from '../../services/DashboardServices';
import { EditableTypography } from "../atoms/EditableTypography";
import { EllipsisTypography } from "../atoms/EllipsisTypography";
/**
 * view
 */

export const AuthenticatorManageDialog: React.VFC<{ onClose: () => void, user: User }> = ({ onClose, user }) => {
  console.log('AuthenticatorManageDialog');
  const webauthnService = useWebAuthnService();
  const { enqueueSnackbar } = useSnackbar();

  const [credentials, setCredentials] = useState<Credential[]>([]);

  const registerdCredId = localStorage.getItem(`credId`)
  const isRegistered = Boolean(registerdCredId && credentials.map(c => c.id).includes(registerdCredId!));

  const [isWebAuthnAvailable, setIsWebAuthnAvailable] = useState(false);

  const checkWebAuthnAvailable = () => {
    if (window.PublicKeyCredential) {
      PublicKeyCredential.isUserVerifyingPlatformAuthenticatorAvailable()
        .then(uvpaa => { setIsWebAuthnAvailable(uvpaa) });
    }
  }
  useEffect(() => { checkWebAuthnAvailable() }, []);

  console.log("credId", registerdCredId, "isRegistered", isRegistered, "isWebAuthnAvailable", isWebAuthnAvailable);

  /**
   * listCredentials
  */
  const listCredentials = async () => {
    console.log("listCredentials");
    try {
      const resp = await webauthnService.listCredentials({ userName: user.name });
      setCredentials(resp.credentials);
    }
    catch (error) {
      handleError(error);
    }
  }
  useEffect(() => { listCredentials() }, []);

  /**
   * registerNewAuthenticator
   */
  const registerNewAuthenticator = async () => {
    try {
      const resp = await webauthnService.beginRegistration({ userName: user.name });
      const options = JSON.parse(resp.credentialCreationOptions);

      const opt: CredentialCreationOptions = JSON.parse(JSON.stringify(options));
      if (options.publicKey?.user.id) {
        opt.publicKey!.user.id = base64url.decode(options.publicKey?.user.id);
      }
      if (options.publicKey?.challenge) {
        opt.publicKey!.challenge = base64url.decode(options.publicKey?.challenge);
      }

      // Credential is allowed to access only id and type so use any.
      const cred: any = await navigator.credentials.create(opt);
      if (cred === null) {
        console.log("cred is null");
        throw Error('credential is null');
      }

      const credential = {
        id: cred.id,
        rawId: base64url.encode(cred.rawId),
        type: cred.type,
        response: {
          clientDataJSON: base64url.encode(cred.response.clientDataJSON),
          attestationObject: base64url.encode(cred.response.attestationObject)
        }
      };
      localStorage.setItem(`credId`, credential.rawId);

      const finResp = await webauthnService.finishRegistration({ userName: user.name, credentialCreationResponse: JSON.stringify(credential) });
      enqueueSnackbar(finResp.message, { variant: 'success' });
      listCredentials();
    }
    catch (error) {
      handleError(error);
    }
  }

  /**
   * removeCredentials
  */
  const removeCredentials = async (id: string) => {
    console.log("removeCredentials");
    if (!confirm("Are you sure to REMOVE?\nID: " + id)) { return }
    try {
      const resp = await webauthnService.deleteCredential({ userName: user.name, credId: id });
      enqueueSnackbar(resp.message, { variant: 'success' });
      listCredentials();
      if (id === registerdCredId) {
        localStorage.removeItem(`credId`);
      }
    }
    catch (error) {
      handleError(error);
    }
  }

  /**
   * updateCredentialName
  */
  const updateCredentialName = async (id: string, name: string) => {
    try {
      const resp = await webauthnService.updateCredential({ userName: user.name, credId: id, credDisplayName: name });
      enqueueSnackbar(resp.message, { variant: 'success' });
      listCredentials();
    }
    catch (error) {
      handleError(error);
    }
  }

  /**
   * error handler
   */
  const handleError = (error: any) => {
    console.log(error);
    const msg = error?.message;
    error instanceof DOMException || msg && enqueueSnackbar(msg, { variant: 'error' });
  }

  const theme = useTheme();
  const sm = useMediaQuery(theme.breakpoints.up('sm'), { noSsr: true });

  return (
    <Dialog open={true}
      fullWidth maxWidth={'sm'}>
      <DialogTitle>WebAuthn Credentials</DialogTitle>
      <DialogContent>
        <Box alignItems="center">
          {credentials.length === 0
            ? <Typography>No credentials</Typography>
            : <Grid container sx={{ p: 1 }}>
              <Grid item xs={0} sm={0.5} ></Grid>
              <Grid item xs={3} sm={1.5} sx={{ textAlign: 'end' }}> <Typography variant="caption" display="block">Created</Typography></Grid>
              <Grid item xs={8} sm={9} ><Typography variant="caption" display="block" sx={{ pl: 2 }}>Credential ID & Name</Typography></Grid>
              <Grid item xs={1} ></Grid>
              <Grid item xs={12} ><Divider /></Grid>
              {credentials.map((field, index) => {
                return (
                  <>
                    {sm && <Grid item xs={0} sm={0.5} zeroMinWidth sx={{ m: 'auto', textAlign: 'center' }}>
                      {registerdCredId === field.id &&
                        <Tooltip title="This credential is created in your device" placement="top-end">
                          <VerifiedOutlined color="success" fontSize="small" />
                        </Tooltip>
                        || undefined}
                    </Grid>}
                    <Grid item xs={3} sm={1.5} sx={{ m: 'auto', textAlign: 'end' }}>
                      <Typography variant="caption" display="block">{field.timestamp?.toDate().toLocaleDateString()}</Typography>
                      <Typography variant="caption" display="block">{field.timestamp?.toDate().toLocaleTimeString()}</Typography>
                    </Grid>
                    <Grid item xs={8} sm={9} sx={{ m: 'auto', p: 2 }}>
                      <EllipsisTypography placement='top'>{field.id}</EllipsisTypography>
                      <EditableTypography onSave={(input) => { updateCredentialName(field.id, input) }}>{field.displayName}</EditableTypography>
                    </Grid >
                    < Grid item xs={1} sx={{ m: 'auto', textAlign: 'center' }}>
                      <IconButton edge="end" aria-label="delete" onClick={() => { removeCredentials(field.id) }}><Delete /></IconButton>
                    </Grid>
                  </>
                )
              })}
            </Grid>}
        </Box>
      </DialogContent >
      <DialogActions>
        <Button onClick={() => onClose()} color="primary">Close</Button>
        {!isRegistered && isWebAuthnAvailable
          ? <Button onClick={() => registerNewAuthenticator()} variant="contained" color="secondary">Register</Button>
          : undefined}
      </DialogActions>
    </Dialog >
  );
};

/**
 * Context
 */
export const AuthenticatorManageDialogContext = DialogContext<{ user: User }>(
  props => (<AuthenticatorManageDialog {...props} />));
