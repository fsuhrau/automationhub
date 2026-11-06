import React, {useCallback, useEffect, useState} from 'react';
import {
    Alert,
    Avatar,
    Box,
    Button,
    Chip,
    FormControl,
    IconButton,
    InputLabel,
    Link,
    MenuItem,
    Select,
    Stack,
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableRow,
    TextField,
    ToggleButton,
    ToggleButtonGroup,
    Tooltip,
    Typography,
} from '@mui/material';
import CssBaseline from '@mui/material/CssBaseline';
import {useColorScheme} from '@mui/material/styles';
import {DeleteOutlineRounded, LogoutRounded} from '@mui/icons-material';
import {LocalizationProvider, DatePicker} from '@mui/x-date-pickers';
import {AdapterDayjs} from '@mui/x-date-pickers/AdapterDayjs';
import dayjs, {Dayjs} from 'dayjs';
import {Link as RouterLink} from 'react-router-dom';
import TopBar from '../../components/TopBar';
import {formatDateTime} from '../../components/DateTime';
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
import {Code, ConfirmDialog, messageOf, Section} from '../settings/parts';
import {ago} from '../apps/bundlesModel';
import {isExpired} from '../settings/settingsModel';

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

const TOKEN_PLACEHOLDER = '<your personal token>';

type Setup = 'container' | 'direct';

// GuideStep is one numbered step of the guide, with a command to copy
const GuideStep: React.FC<{ title: string, code?: string, children: React.ReactNode }> = ({title, code, children}) =>
    <Box component="li" sx={{
        counterIncrement: 'step', position: 'relative', pl: 4.5, pb: 2,
        '&::before': {
            content: 'counter(step)', position: 'absolute', left: 0, top: 0, width: 24, height: 24, borderRadius: '50%',
            display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 13, fontWeight: 600,
            bgcolor: 'primary.main', color: 'primary.contrastText',
        },
    }}>
        <Typography variant="body2" sx={{fontWeight: 600, lineHeight: '24px'}}>{title}</Typography>
        <Typography variant="body2" color="text.secondary" sx={{mb: code ? 1 : 0}}>{children}</Typography>
        {code && <Code label="Copy the command">{code}</Code>}
    </Box>;

// ConnectGuide walks through connecting Claude Code to the hub with a
// personal token; with a token just created, its value is filled in
const ConnectGuide: React.FC<{ projects: IProject[], token: ICreatedPersonalToken | null }> = ({projects, token}) => {
    const [setup, setSetup] = useState<Setup>('container');
    const [project, setProject] = useState('');
    useEffect(() => {
        if (!project && projects.length > 0) setProject(projects[0].identifier);
    }, [projects, project]);
    const value = token?.token ?? TOKEN_PLACEHOLDER;

    return <Section title="Connect Claude Code" description="Let Claude Code run tests and control devices through the hub's MCP server, as you.">
        <Stack direction={{xs: 'column', sm: 'row'}} spacing={2} alignItems={{sm: 'center'}} sx={{mb: 2}}>
            <ToggleButtonGroup size="small" exclusive value={setup} onChange={(_, v) => v && setSetup(v)} aria-label="Where Claude Code runs">
                <ToggleButton value="container">In a dev container</ToggleButton>
                <ToggleButton value="direct">Directly on this computer</ToggleButton>
            </ToggleButtonGroup>
            {projects.length > 0 && <FormControl size="small" sx={{minWidth: 200}}>
                <InputLabel id="guide-project">Project</InputLabel>
                <Select labelId="guide-project" label="Project" value={project} onChange={e => setProject(e.target.value)}>
                    {projects.map(p => <MenuItem key={p.identifier} value={p.identifier}>{p.name}</MenuItem>)}
                </Select>
            </FormControl>}
        </Stack>
        <Box component="ol" sx={{m: 0, p: 0, listStyle: 'none', counterReset: 'step'}}>
            <GuideStep title={`Create a personal token above${token ? ` (done: ${token.name})` : ''}`}>
                It's shown once; copy it when it appears.
            </GuideStep>
            {setup === 'container' ? <>
                <GuideStep title="Keep the token on the host, in the Keychain" code={keychainCommand()}>
                    Containers must not hold the token. The command asks for it.
                </GuideStep>
                <GuideStep title="Run the MCP proxy on the host" code={proxyCommand(project || '<project>')}>
                    It adds the token to Claude Code's requests; the container connects to the proxy instead of the hub.
                </GuideStep>
            </> : <GuideStep title="Add the hub to Claude Code" code={mcpCommand(project || '<project>', value)}>
                {token ? 'Your new token is filled in.' : `Put your token in place of ${TOKEN_PLACEHOLDER}.`}
            </GuideStep>}
        </Box>
    </Section>;
};

// ProfilePage is about the signed-in person: who they are, how the hub looks
// for them, and the personal tokens that act as them
const ProfilePage: React.FC = () => {
    const auth = useAuth();
    const {setError} = useError();
    const {mode, setMode} = useColorScheme();

    const [profile, setProfile] = useState<IProfile | null>(null);
    const [profileError, setProfileError] = useState<string | null>(null);
    const [tokens, setTokens] = useState<IPersonalToken[] | null>(null);
    const [projects, setProjects] = useState<IProject[]>([]);

    const [name, setName] = useState('');
    const [expiresAt, setExpiresAt] = useState<Dayjs | null>(null);
    const [tried, setTried] = useState(false);
    const [created, setCreated] = useState<ICreatedPersonalToken | null>(null);
    const [revoking, setRevoking] = useState<IPersonalToken | null>(null);
    const [busy, setBusy] = useState(false);
    const now = Date.now();

    const loadTokens = useCallback(() => {
        getPersonalTokens().then(setTokens).catch(e => setError(messageOf(e)));
    }, [setError]);

    useEffect(() => {
        if (auth.authState !== 'authenticated') return;
        getProfile().then(p => {
            setProfile(p);
            loadTokens();
        }).catch(e => setProfileError(messageOf(e)));
        getProjects().then(setProjects).catch(() => undefined);
    }, [auth.authState, loadTokens]);

    const create = () => {
        setTried(true);
        if (!name.trim()) return;
        setBusy(true);
        createPersonalToken(name.trim(), expiresAt).then(token => {
            setCreated(token);
            setName('');
            setExpiresAt(null);
            setTried(false);
            loadTokens();
        }).catch(e => setError(messageOf(e))).finally(() => setBusy(false));
    };

    const revoke = () => {
        if (!revoking) return;
        setBusy(true);
        deletePersonalToken(revoking.id).then(() => {
            if (created?.id === revoking.id) setCreated(null);
            setRevoking(null);
            loadTokens();
        }).catch(e => setError(messageOf(e))).finally(() => setBusy(false));
    };

    const user = auth.user;
    const shownName = profile?.name ?? user?.name ?? '';

    return (
        <>
            <CssBaseline enableColorScheme/>
            <TopBar/>
            <Box sx={{maxWidth: 1000, mx: 'auto', px: 2, py: 4}}>
                <Typography variant="h5" component="h1">Profile</Typography>
                <Typography variant="body2" color="text.secondary" sx={{mb: 3}}>Your account, how the hub looks for you, and the tokens that act as you.</Typography>

                {auth.authState === 'disabled' && <Alert severity="info" sx={{mb: 3}}>
                    This hub runs without sign-in, so there are no accounts or personal tokens: everyone who reaches it can use it.
                    For scripts, CI and Claude Code, create a project token in a project's settings under Access tokens.
                    {' '}<Link component={RouterLink} to="/">Pick a project</Link>.
                </Alert>}
                {profileError && <Alert severity="warning" sx={{mb: 3}}>{profileError}</Alert>}

                <Stack spacing={3}>
                    {user && <Section title="Account" action={
                        <Button variant="outlined" size="small" startIcon={<LogoutRounded/>} onClick={() => auth.logOut()}>Sign out</Button>}>
                        <Stack direction="row" spacing={2} alignItems="center">
                            <Avatar src={user.avatar || undefined} alt={shownName} sx={{width: 56, height: 56}}>{shownName[0]}</Avatar>
                            <Box sx={{minWidth: 0}}>
                                <Typography variant="h6">{shownName}</Typography>
                                {(profile?.email || user.email) && <Typography variant="body2" color="text.secondary">{profile?.email || user.email}</Typography>}
                                {/* the login only says something when it isn't the name */}
                                {(profile?.login || user.login) && (profile?.login || user.login) !== shownName &&
                                    <Typography variant="caption" color="text.secondary">signed in as {profile?.login || user.login}</Typography>}
                            </Box>
                        </Stack>
                    </Section>}

                    <Section title="Appearance" description="Light, dark, or as your system is set. Stored in this browser.">
                        {mode && <ToggleButtonGroup size="small" exclusive value={mode} onChange={(_, v) => v && setMode(v)} aria-label="Color mode">
                            <ToggleButton value="system">System</ToggleButton>
                            <ToggleButton value="light">Light</ToggleButton>
                            <ToggleButton value="dark">Dark</ToggleButton>
                        </ToggleButtonGroup>}
                    </Section>

                    {profile && <>
                        <Section title="Personal access tokens" description={<>
                            A personal token acts as you in every project: for Claude Code (MCP) and scripts, sent in the <code>X-Auth-Token</code> header.
                            Device locks show your name. The value is shown once, when you create it.
                        </>}>
                            <Box component="form" noValidate onSubmit={e => {
                                e.preventDefault();
                                create();
                            }} sx={{mb: 2}}>
                                <Stack direction={{xs: 'column', sm: 'row'}} spacing={1} alignItems={{sm: 'flex-start'}}>
                                    <TextField size="small" label="Name" placeholder="e.g. claude-code" value={name} onChange={e => setName(e.target.value)}
                                               error={tried && !name.trim()} helperText={tried && !name.trim() ? 'A name is needed.' : 'Says where the token is used.'}
                                               sx={{minWidth: 220}}/>
                                    <LocalizationProvider dateAdapter={AdapterDayjs}>
                                        <DatePicker label="Expires (optional)" value={expiresAt} onChange={setExpiresAt} minDate={dayjs().add(1, 'day')}
                                                    slotProps={{textField: {size: 'small', helperText: 'Empty: until revoked'}, field: {clearable: true}}}/>
                                    </LocalizationProvider>
                                    <Button type="submit" variant="contained" disabled={busy}>Create token</Button>
                                </Stack>
                            </Box>

                            {created && <Alert severity="success" onClose={() => setCreated(null)} sx={{mb: 2, '& .MuiAlert-message': {flex: 1, minWidth: 0}}}>
                                <Typography variant="body2" sx={{mb: 1}}>
                                    <b>{created.name}</b> is ready. Copy it now; it won't be shown again. The guide below has it filled in.
                                </Typography>
                                <Code label="Copy the token">{created.token}</Code>
                            </Alert>}

                            {tokens === null
                                ? null
                                : tokens.length === 0
                                    ? <Typography variant="body2" color="text.secondary">No personal tokens yet.</Typography>
                                    : <Box sx={{overflowX: 'auto'}}><Table size="small" aria-label="Personal tokens">
                                        <TableHead><TableRow>
                                            <TableCell>Name</TableCell><TableCell>Token</TableCell><TableCell>Last used</TableCell><TableCell>Created</TableCell><TableCell>Expires</TableCell><TableCell/>
                                        </TableRow></TableHead>
                                        <TableBody>
                                            {tokens.map(t => <TableRow key={t.id}>
                                                <TableCell>{t.name}</TableCell>
                                                <TableCell sx={{fontFamily: 'monospace', whiteSpace: 'nowrap'}}>ahp_…{t.hint}</TableCell>
                                                <TableCell sx={{whiteSpace: 'nowrap'}}>
                                                    {t.lastUsedAt
                                                        ? <Tooltip title={formatDateTime(t.lastUsedAt)}><span>{ago(t.lastUsedAt, now)}</span></Tooltip>
                                                        : <Typography variant="body2" color="text.secondary" component="span">never</Typography>}
                                                </TableCell>
                                                <TableCell sx={{whiteSpace: 'nowrap'}}>{formatDateTime(t.createdAt, 'YYYY/MM/DD')}</TableCell>
                                                <TableCell sx={{whiteSpace: 'nowrap'}}>
                                                    {t.expiresAt
                                                        ? <>{formatDateTime(t.expiresAt, 'YYYY/MM/DD')}{isExpired(t.expiresAt, now) && <Chip size="small" color="warning" label="expired" sx={{ml: 1}}/>}</>
                                                        : 'never'}
                                                </TableCell>
                                                <TableCell align="right">
                                                    <Tooltip title="Revoke">
                                                        <IconButton size="small" aria-label={`Revoke ${t.name}`} onClick={() => setRevoking(t)}><DeleteOutlineRounded fontSize="small"/></IconButton>
                                                    </Tooltip>
                                                </TableCell>
                                            </TableRow>)}
                                        </TableBody>
                                    </Table></Box>}
                        </Section>

                        <ConnectGuide projects={projects} token={created}/>
                    </>}
                </Stack>
            </Box>

            <ConfirmDialog open={!!revoking} title={`Revoke ${revoking?.name}?`} confirm="Revoke token" busy={busy}
                           onConfirm={revoke} onClose={() => setRevoking(null)}>
                Claude Code, scripts and anything else using it get "unauthorized" from now on.
            </ConfirmDialog>
        </>
    );
};

export default ProfilePage;
