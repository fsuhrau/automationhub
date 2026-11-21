import React from 'react';
import {Box, Link, Table, TableBody, TableCell, TableHead, TableRow, Typography} from '@mui/material';
import {Link as RouterLink} from 'react-router-dom';
import {Code, ConfirmDialog, Section} from '../../components/common';
import {useProjectContext} from '../../hooks/ProjectProvider';
import {useNow} from '../../hooks/useNow';
import {routes} from '../../lib/routes';
import IAccessTokenData from '../../types/access.token';
import AccessTokenForm from './sections/AccessTokenForm';
import AccessTokenRow from './sections/AccessTokenRow';
import {useAccessTokens} from './useAccessTokens';
import {useConfirmAction} from './useConfirmAction';

const Description: React.FC = () => <>
    Tokens for scripts, CI and agents to use this project's API, sent in the <code>X-Auth-Token</code> header.
    They only work for this project. For a token that acts as you, use your <Link component={RouterLink} to={routes.profile()}>profile</Link>.
</>;

const CreatedToken: React.FC<{ token: IAccessTokenData }> = ({token}) => {
    const {projectIdentifier} = useProjectContext();
    return <Box sx={{mb: 2}}>
        <Typography variant="body2" sx={{mb: 1}}>Token <b>{token.name}</b> is ready. For example:</Typography>
        <Code label="Copy the command">{`curl -H "X-Auth-Token: ${token.token}" ${window.location.origin}/api/${projectIdentifier}/devices`}</Code>
    </Box>;
};

const NOW_REFRESH_MS = 60_000;
const HEADERS = ['Name', 'Token', 'Created', 'Expires'];

// AccessTokenTab manages the project's tokens for scripts, CI and the MCP
// server; personal tokens live on the profile page
const AccessTokenTab: React.FC = () => {
    const {tokens, created, create, revoke} = useAccessTokens();
    const revoking = useConfirmAction(revoke);
    const now = useNow(NOW_REFRESH_MS);

    return <Section title="Access tokens" description={<Description/>}>
        <AccessTokenForm busy={create.busy} onCreate={create.run}/>
        {created && <CreatedToken token={created}/>}
        {tokens.length === 0
            ? <Typography variant="body2" color="text.secondary">No tokens yet.</Typography>
            : <Box sx={{overflowX: 'auto'}}><Table size="small" aria-label="Access tokens">
                <TableHead><TableRow>{HEADERS.map(h => <TableCell key={h}>{h}</TableCell>)}<TableCell/></TableRow></TableHead>
                <TableBody>{tokens.map(t => <AccessTokenRow key={t.id} token={t} now={now} onRevoke={revoking.ask}/>)}</TableBody>
            </Table></Box>}
        <ConfirmDialog open={!!revoking.target} title={`Revoke ${revoking.target?.name || 'this token'}?`} confirm="Revoke token"
            busy={revoking.busy} onConfirm={revoking.confirm} onClose={revoking.cancel}>
            Anything that uses it (scripts, CI, an MCP client) gets "unauthorized" from now on.
        </ConfirmDialog>
    </Section>;
};

export default AccessTokenTab;
