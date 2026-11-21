import React from 'react';
import {Box, Chip, IconButton, Table, TableBody, TableCell, TableHead, TableRow, Tooltip, Typography} from '@mui/material';
import {DeleteOutlineRounded} from '@mui/icons-material';
import {formatDateTime} from '../../../components/DateTime';
import {IPersonalToken} from '../../../types/profile.token';
import {ago} from '../../apps/bundlesModel';
import {isExpired} from '../../settings/settingsModel';

export interface PersonalTokensTableProps {
    tokens: IPersonalToken[],
    now: number,
    onRevoke: (token: IPersonalToken) => void,
}

const NOWRAP = {whiteSpace: 'nowrap'} as const;
const HEADERS = ['Name', 'Token', 'Last used', 'Created', 'Expires'];

const LastUsed: React.FC<{ lastUsedAt: string | null, now: number }> = ({lastUsedAt, now}) => lastUsedAt
    ? <Tooltip title={formatDateTime(lastUsedAt)}><span>{ago(lastUsedAt, now)}</span></Tooltip>
    : <Typography variant="body2" color="text.secondary" component="span">never</Typography>;

const Expiry: React.FC<{ expiresAt: string | null, now: number }> = ({expiresAt, now}) => expiresAt
    ? <>{formatDateTime(expiresAt, 'YYYY/MM/DD')}{isExpired(expiresAt, now) && <Chip size="small" color="warning" label="expired" sx={{ml: 1}}/>}</>
    : <>never</>;

const TokenRow: React.FC<{ token: IPersonalToken } & Omit<PersonalTokensTableProps, 'tokens'>> = ({token, now, onRevoke}) =>
    <TableRow>
        <TableCell>{token.name}</TableCell>
        <TableCell sx={{fontFamily: 'monospace', ...NOWRAP}}>ahp_…{token.hint}</TableCell>
        <TableCell sx={NOWRAP}><LastUsed lastUsedAt={token.lastUsedAt} now={now}/></TableCell>
        <TableCell sx={NOWRAP}>{formatDateTime(token.createdAt, 'YYYY/MM/DD')}</TableCell>
        <TableCell sx={NOWRAP}><Expiry expiresAt={token.expiresAt} now={now}/></TableCell>
        <TableCell align="right">
            <Tooltip title="Revoke">
                <IconButton size="small" aria-label={`Revoke ${token.name}`} onClick={() => onRevoke(token)}>
                    <DeleteOutlineRounded fontSize="small"/>
                </IconButton>
            </Tooltip>
        </TableCell>
    </TableRow>;

const PersonalTokensTable: React.FC<PersonalTokensTableProps> = ({tokens, ...rest}) =>
    <Box sx={{overflowX: 'auto'}}><Table size="small" aria-label="Personal tokens">
        <TableHead><TableRow>{HEADERS.map(h => <TableCell key={h}>{h}</TableCell>)}<TableCell/></TableRow></TableHead>
        <TableBody>{tokens.map(t => <TokenRow key={t.id} token={t} {...rest}/>)}</TableBody>
    </Table></Box>;

export default PersonalTokensTable;
