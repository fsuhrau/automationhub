import React, {useCallback, useEffect, useState} from 'react';
import {useNavigate} from 'react-router-dom';
import {
    Alert,
    Box,
    Button,
    Card,
    CardContent,
    Dialog,
    DialogActions,
    DialogContent,
    DialogTitle,
    FormControl,
    IconButton,
    InputLabel,
    MenuItem,
    Select,
    Stack,
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableRow,
    TextField,
    Tooltip,
    Typography,
} from '@mui/material';
import CssBaseline from '@mui/material/CssBaseline';
import {ContentCopy} from '@mui/icons-material';
import DeleteForeverIcon from '@mui/icons-material/DeleteForever';
import {LocalizationProvider, MobileDatePicker} from '@mui/x-date-pickers';
import {AdapterDayjs} from '@mui/x-date-pickers/AdapterDayjs';
import dayjs, {Dayjs} from 'dayjs';
import TopBar from '../../components/TopBar';
import DateTime from '../../components/DateTime';
import {useAuth} from '../../hooks/AuthProvider';
import {useError} from '../../ErrorProvider';
import IProject from '../../project/project';
import {getProjects} from '../../project/project.service';
import {
    createPersonalToken,
    deletePersonalToken,
    getPersonalTokens,
    getProfile,
    ICreatedPersonalToken,
    IPersonalToken,
    IProfile,
} from '../../services/profile.service';

const errorMessage = (error: unknown): string => {
    const e = error as { response?: { data?: { message?: string } }, message?: string };
    return e.response?.data?.message ?? e.message ?? String(error);
};

const copy = (text: string) => navigator.clipboard.writeText(text).catch(() => undefined);

const CopyButton: React.FC<{ text: string, label?: string }> = ({text, label = 'Copy'}) =>
    <Tooltip title={label}>
        <IconButton size="small" color="primary" onClick={() => copy(text)}><ContentCopy fontSize="small"/></IconButton>
    </Tooltip>;

const Code: React.FC<{ children: string }> = ({children}) =>
    <Box sx={{display: 'flex', alignItems: 'flex-start', gap: 1}}>
        <Box component="pre" sx={{
            m: 0, p: 1.5, flex: 1, overflowX: 'auto', fontSize: 13, whiteSpace: 'pre',
            bgcolor: 'background.default', border: 1, borderColor: 'divider', borderRadius: 1,
        }}>{children}</Box>
        <CopyButton text={children}/>
    </Box>;

// The MCP endpoint of a project, as the master serves it.
const mcpUrl = (project: string) => `${window.location.origin}/api/${encodeURIComponent(project)}/mcp`;

const mcpCommand = (project: string, token: string) =>
    `claude mcp add --transport http hub \\\n  ${mcpUrl(project)} \\\n  --header "X-Auth-Token: ${token}"`;

// Dev containers must not hold the token: it goes into the host's Keychain
// and "automationhub mcp-proxy" adds it on the host (account = master host).
const keychainCommand = () =>
    `security add-generic-password -U -s automationhub -a ${window.location.host} -w`;

const proxyCommand = (project: string) =>
    `automationhub mcp-proxy --upstream ${window.location.origin} --project ${project}`;

const ProfilePage: React.FC = () => {
    const navigate = useNavigate();
    const auth = useAuth();
    const {setError} = useError();

    const [profile, setProfile] = useState<IProfile | null>(null);
    const [profileError, setProfileError] = useState<string | null>(null);
    const [tokens, setTokens] = useState<IPersonalToken[]>([]);
    const [projects, setProjects] = useState<IProject[]>([]);
    const [project, setProject] = useState('');

    const [name, setName] = useState('');
    const [expiresAt, setExpiresAt] = useState<Dayjs | null>(null);
    const [created, setCreated] = useState<ICreatedPersonalToken | null>(null);
    const [revoke, setRevoke] = useState<IPersonalToken | null>(null);
    const [busy, setBusy] = useState(false);

    const loadTokens = useCallback(() => {
        getPersonalTokens().then(setTokens).catch(e => setError(errorMessage(e)));
    }, [setError]);

    useEffect(() => {
        if (auth.authState !== 'authenticated') return;
        getProfile().then(p => {
            setProfile(p);
            loadTokens();
        }).catch(e => setProfileError(errorMessage(e)));
        getProjects().then(list => {
            setProjects(list);
            if (list.length > 0) setProject(p => p || list[0].identifier);
        }).catch(() => undefined);
    }, [auth.authState, loadTokens]);

    const onCreate = async () => {
        setBusy(true);
        try {
            const token = await createPersonalToken(name.trim(), expiresAt);
            setCreated(token);
            setName('');
            setExpiresAt(null);
            loadTokens();
        } catch (e) {
            setError(errorMessage(e));
        } finally {
            setBusy(false);
        }
    };

    const onRevoke = async () => {
        if (!revoke) return;
        try {
            await deletePersonalToken(revoke.id);
            if (created?.id === revoke.id) setCreated(null);
            loadTokens();
        } catch (e) {
            setError(errorMessage(e));
        } finally {
            setRevoke(null);
        }
    };

    return (
        <>
            <CssBaseline enableColorScheme/>
            <TopBar/>
            <Box sx={{maxWidth: 1000, mx: 'auto', px: 2, py: 4}}>
                <Typography variant="h5" component="h1" sx={{mb: 3}}>Profile</Typography>

                {auth.authState === 'disabled' && <Alert severity="info">
                    This master runs without authentication, so there are no user accounts or personal tokens.
                    Every API call and the MCP endpoint are open; use a project access token
                    (project settings) to name the caller.
                </Alert>}
                {profileError && <Alert severity="warning">{profileError}</Alert>}

                {profile && <Stack spacing={3}>
                    <Card variant="outlined">
                        <CardContent>
                            <Typography variant="h6">{profile.name}</Typography>
                            {profile.email && <Typography color="text.secondary">{profile.email}</Typography>}
                        </CardContent>
                    </Card>

                    <Card variant="outlined">
                        <CardContent>
                            <Typography variant="h6" gutterBottom>Personal access tokens</Typography>
                            <Typography variant="body2" color="text.secondary" sx={{mb: 2}}>
                                A personal token acts as you, in every project: use it for the MCP server
                                (Claude Code) and scripts. Device locks show your name. The token is only shown
                                once, when you create it; tokens can only be created and revoked here.
                            </Typography>

                            {created && <Alert severity="success" onClose={() => setCreated(null)} sx={{mb: 2}}>
                                <Typography variant="body2" sx={{mb: 1}}>
                                    Token <b>{created.name}</b> created. Copy it now, it won't be shown again.
                                </Typography>
                                <Code>{created.token}</Code>
                                {projects.length > 0 && <Box sx={{mt: 2}}>
                                    <Stack direction="row" spacing={2} alignItems="center" sx={{mb: 1}}>
                                        <Typography variant="body2">Add it to Claude Code for project</Typography>
                                        <FormControl size="small">
                                            <InputLabel id="mcp-project">Project</InputLabel>
                                            <Select labelId="mcp-project" label="Project" value={project}
                                                    onChange={e => setProject(e.target.value)}>
                                                {projects.map(p => <MenuItem key={p.identifier} value={p.identifier}>{p.name}</MenuItem>)}
                                            </Select>
                                        </FormControl>
                                    </Stack>
                                    {project && <Stack spacing={1}>
                                        <Typography variant="body2">
                                            <b>Dev containers</b> must not hold the token. Keep it on the host: store it in the
                                            Keychain (the command asks for the value), then run the proxy on the host. It adds
                                            the token to MCP requests, and the container connects to the proxy instead.
                                        </Typography>
                                        <Code>{keychainCommand()}</Code>
                                        <Code>{proxyCommand(project)}</Code>
                                        <Typography variant="body2" sx={{pt: 1}}>
                                            Claude Code running directly on your machine (no container):
                                        </Typography>
                                        <Code>{mcpCommand(project, created.token)}</Code>
                                    </Stack>}
                                </Box>}
                            </Alert>}

                            <Stack direction={{xs: 'column', sm: 'row'}} spacing={2} alignItems={{sm: 'center'}} sx={{mb: 2}}>
                                <TextField size="small" label="Name" placeholder="e.g. claude-code" value={name}
                                           onChange={e => setName(e.target.value)}
                                           onKeyDown={e => e.key === 'Enter' && name.trim() && !busy && onCreate()}/>
                                <LocalizationProvider dateAdapter={AdapterDayjs}>
                                    <MobileDatePicker label="Expires (optional)" value={expiresAt} minDate={dayjs().add(1, 'day')}
                                                      onChange={setExpiresAt}
                                                      slotProps={{textField: {size: 'small'}, actionBar: {actions: ['clear', 'cancel', 'accept']}}}/>
                                </LocalizationProvider>
                                <Button variant="contained" disabled={!name.trim() || busy} onClick={onCreate}>Create token</Button>
                            </Stack>

                            {tokens.length === 0
                                ? <Typography variant="body2" color="text.secondary">No personal tokens yet.</Typography>
                                : <Box sx={{overflowX: 'auto'}}>
                                    <Table size="small">
                                        <TableHead>
                                            <TableRow>
                                                <TableCell>Name</TableCell>
                                                <TableCell>Token</TableCell>
                                                <TableCell>Created</TableCell>
                                                <TableCell>Last used</TableCell>
                                                <TableCell>Expires</TableCell>
                                                <TableCell/>
                                            </TableRow>
                                        </TableHead>
                                        <TableBody>
                                            {tokens.map(t => <TableRow key={t.id}>
                                                <TableCell>{t.name}</TableCell>
                                                <TableCell sx={{fontFamily: 'monospace'}}>ahp_…{t.hint}</TableCell>
                                                <TableCell><DateTime format="YYYY/MM/DD HH:mm" value={t.createdAt}/></TableCell>
                                                <TableCell>{t.lastUsedAt ? <DateTime format="YYYY/MM/DD HH:mm" value={t.lastUsedAt}/> : 'Never'}</TableCell>
                                                <TableCell>{t.expiresAt ? <DateTime format="YYYY/MM/DD" value={t.expiresAt}/> : 'Never'}</TableCell>
                                                <TableCell align="right">
                                                    <Tooltip title="Revoke">
                                                        <Tooltip title="Revoke"><IconButton size="small" color="secondary" aria-label={`Revoke ${t.name}`} onClick={() => setRevoke(t)}><DeleteForeverIcon/></IconButton></Tooltip>
                                                    </Tooltip>
                                                </TableCell>
                                            </TableRow>)}
                                        </TableBody>
                                    </Table>
                                </Box>}
                        </CardContent>
                    </Card>
                </Stack>}
            </Box>

            <Dialog open={revoke !== null} onClose={() => setRevoke(null)}>
                <DialogTitle>Revoke token "{revoke?.name}"?</DialogTitle>
                <DialogContent>
                    <Typography>Clients using this token stop working right away.</Typography>
                </DialogContent>
                <DialogActions>
                    <Button onClick={() => setRevoke(null)}>Cancel</Button>
                    <Button color="error" variant="contained" onClick={onRevoke}>Revoke</Button>
                </DialogActions>
            </Dialog>
        </>
    );
};

export default ProfilePage;
