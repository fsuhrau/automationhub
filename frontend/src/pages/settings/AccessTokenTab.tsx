import React, {useEffect, useState} from 'react';
import {
    Box,
    Button,
    Chip,
    IconButton,
    Link,
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
import {DeleteOutlineRounded, VisibilityOffRounded, VisibilityRounded} from '@mui/icons-material';
import {LocalizationProvider, DatePicker} from '@mui/x-date-pickers';
import {AdapterDayjs} from '@mui/x-date-pickers/AdapterDayjs';
import {Dayjs} from 'dayjs';
import {Link as RouterLink} from 'react-router-dom';
import {createAccessToken, deleteAccessToken, getAccessTokens} from '../../services/settings.service';
import {HubStateActions} from '../../application/HubState';
import {useHubState} from '../../hooks/HubStateProvider';
import {useProjectContext} from '../../hooks/ProjectProvider';
import {useError} from '../../ErrorProvider';
import {formatDateTime} from '../../components/DateTime';
import IAccessTokenData from '../../types/access.token';
import {Code, ConfirmDialog, CopyButton, messageOf, Section} from './parts';
import {isExpired, maskToken} from './settingsModel';

// AccessTokenTab manages the project's tokens for scripts, CI and the MCP
// server; personal tokens live on the profile page
const AccessTokenTab: React.FC = () => {
    const {state, dispatch} = useHubState();
    const {projectIdentifier} = useProjectContext();
    const {setError} = useError();
    const [name, setName] = useState('');
    const [expiresAt, setExpiresAt] = useState<Dayjs | null>(null);
    const [tried, setTried] = useState(false);
    const [busy, setBusy] = useState(false);
    const [shown, setShown] = useState<Set<number>>(new Set());
    const [revoking, setRevoking] = useState<IAccessTokenData | null>(null);
    const [created, setCreated] = useState<IAccessTokenData | null>(null);
    const now = Date.now();

    useEffect(() => {
        getAccessTokens(projectIdentifier).then(list => dispatch({type: HubStateActions.AccessTokensUpdate, payload: list}))
            .catch(e => setError(messageOf(e)));
    }, [projectIdentifier]);

    const create = () => {
        setTried(true);
        if (!name.trim()) return;
        setBusy(true);
        createAccessToken(projectIdentifier, {name: name.trim(), expiresAt}).then(token => {
            dispatch({type: HubStateActions.AccessTokenAdd, payload: token});
            setCreated(token);
            setName('');
            setExpiresAt(null);
            setTried(false);
        }).catch(e => setError(messageOf(e))).finally(() => setBusy(false));
    };

    const revoke = () => {
        if (!revoking) return;
        setBusy(true);
        deleteAccessToken(projectIdentifier, revoking.id).then(() => {
            dispatch({type: HubStateActions.AccessTokenDelete, payload: revoking.id});
            if (created?.id === revoking.id) setCreated(null);
            setRevoking(null);
        }).catch(e => setError(messageOf(e))).finally(() => setBusy(false));
    };

    const tokens = state.accessTokens ?? [];

    return <Section title="Access tokens" description={<>
        Tokens for scripts, CI and agents to use this project's API, sent in the <code>X-Auth-Token</code> header.
        They only work for this project. For a token that acts as you, use your <Link component={RouterLink} to="/profile">profile</Link>.
    </>}>
        <Box component="form" noValidate onSubmit={e => {
            e.preventDefault();
            create();
        }} sx={{mb: 2}}>
            <Stack direction={{xs: 'column', sm: 'row'}} spacing={1} alignItems={{sm: 'flex-start'}}>
                <TextField size="small" label="Name" placeholder="e.g. CI nightly" value={name} onChange={e => setName(e.target.value)}
                           error={tried && !name.trim()} helperText={tried && !name.trim() ? 'A name is needed.' : 'Says where the token is used.'}
                           sx={{minWidth: 240}}/>
                <LocalizationProvider dateAdapter={AdapterDayjs}>
                    <DatePicker label="Expires (optional)" value={expiresAt} onChange={setExpiresAt} disablePast
                                slotProps={{textField: {size: 'small', helperText: 'Empty: until revoked'}, field: {clearable: true}}}/>
                </LocalizationProvider>
                <Button type="submit" variant="contained" disabled={busy}>Create token</Button>
            </Stack>
        </Box>

        {created && <Box sx={{mb: 2}}>
            <Typography variant="body2" sx={{mb: 1}}>Token <b>{created.name}</b> is ready. For example:</Typography>
            <Code label="Copy the command">{`curl -H "X-Auth-Token: ${created.token}" ${window.location.origin}/api/${projectIdentifier}/devices`}</Code>
        </Box>}

        {tokens.length === 0
            ? <Typography variant="body2" color="text.secondary">No tokens yet.</Typography>
            : <Box sx={{overflowX: 'auto'}}><Table size="small" aria-label="Access tokens">
                <TableHead><TableRow>
                    <TableCell>Name</TableCell><TableCell>Token</TableCell><TableCell>Created</TableCell><TableCell>Expires</TableCell><TableCell/>
                </TableRow></TableHead>
                <TableBody>
                    {tokens.map(t => {
                        const visible = shown.has(t.id);
                        const expired = isExpired(t.expiresAt, now);
                        return <TableRow key={t.id}>
                            <TableCell>{t.name || <Typography variant="body2" color="text.secondary" component="span">unnamed</Typography>}</TableCell>
                            <TableCell sx={{whiteSpace: 'nowrap'}}>
                                <Stack direction="row" alignItems="center" spacing={0.5}>
                                    <Typography variant="body2" sx={{fontFamily: 'monospace'}}>{visible ? t.token : maskToken(t.token)}</Typography>
                                    <Tooltip title={visible ? 'Hide' : 'Show'}>
                                        <IconButton size="small" aria-label={visible ? 'Hide the token' : 'Show the token'}
                                                    onClick={() => setShown(prev => {
                                                        const next = new Set(prev);
                                                        if (next.has(t.id)) next.delete(t.id); else next.add(t.id);
                                                        return next;
                                                    })}>
                                            {visible ? <VisibilityOffRounded fontSize="small"/> : <VisibilityRounded fontSize="small"/>}
                                        </IconButton>
                                    </Tooltip>
                                    <CopyButton text={t.token} label="Copy the token"/>
                                </Stack>
                            </TableCell>
                            <TableCell sx={{whiteSpace: 'nowrap'}}>{formatDateTime(t.createdAt, 'YYYY/MM/DD')}</TableCell>
                            <TableCell sx={{whiteSpace: 'nowrap'}}>
                                {t.expiresAt
                                    ? <>{formatDateTime(t.expiresAt, 'YYYY/MM/DD')}{expired && <Chip size="small" color="warning" label="expired" sx={{ml: 1}}/>}</>
                                    : 'never'}
                            </TableCell>
                            <TableCell align="right">
                                <Tooltip title="Revoke">
                                    <IconButton size="small" aria-label={`Revoke ${t.name}`} onClick={() => setRevoking(t)}><DeleteOutlineRounded fontSize="small"/></IconButton>
                                </Tooltip>
                            </TableCell>
                        </TableRow>;
                    })}
                </TableBody>
            </Table></Box>}

        <ConfirmDialog open={!!revoking} title={`Revoke ${revoking?.name || 'this token'}?`} confirm="Revoke token" busy={busy}
                       onConfirm={revoke} onClose={() => setRevoking(null)}>
            Anything that uses it (scripts, CI, an MCP client) gets "unauthorized" from now on.
        </ConfirmDialog>
    </Section>;
};

export default AccessTokenTab;
