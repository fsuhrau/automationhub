import React, {useState} from 'react';
import {Chip, IconButton, Stack, TableCell, TableRow, Tooltip, Typography} from '@mui/material';
import {DeleteOutlineRounded, VisibilityOffRounded, VisibilityRounded} from '@mui/icons-material';
import {CopyButton} from '../../../components/common';
import {formatDateTime} from '../../../components/DateTime';
import IAccessTokenData from '../../../types/access.token';
import {isExpired, maskToken} from '../settingsModel';

export interface AccessTokenRowProps {
    token: IAccessTokenData,
    now: number,
    onRevoke: (token: IAccessTokenData) => void,
}

const ExpiryCell: React.FC<{ expiresAt: IAccessTokenData['expiresAt'], now: number }> = ({expiresAt, now}) =>
    <TableCell sx={{whiteSpace: 'nowrap'}}>
        {expiresAt
            ? <>{formatDateTime(expiresAt, 'YYYY/MM/DD')}{isExpired(expiresAt, now) && <Chip size="small" color="warning" label="expired" sx={{ml: 1}}/>}</>
            : 'never'}
    </TableCell>;

// AccessTokenRow is one token of the table; its value is masked until shown
const AccessTokenRow: React.FC<AccessTokenRowProps> = ({token, now, onRevoke}) => {
    const [visible, setVisible] = useState(false);
    return <TableRow>
        <TableCell>{token.name || <Typography variant="body2" color="text.secondary" component="span">unnamed</Typography>}</TableCell>
        <TableCell sx={{whiteSpace: 'nowrap'}}>
            <Stack direction="row" alignItems="center" spacing={0.5}>
                <Typography variant="body2" sx={{fontFamily: 'monospace'}}>{visible ? token.token : maskToken(token.token)}</Typography>
                <Tooltip title={visible ? 'Hide' : 'Show'}>
                    <IconButton size="small" aria-label={visible ? 'Hide the token' : 'Show the token'} onClick={() => setVisible(v => !v)}>
                        {visible ? <VisibilityOffRounded fontSize="small"/> : <VisibilityRounded fontSize="small"/>}
                    </IconButton>
                </Tooltip>
                <CopyButton text={token.token} label="Copy the token"/>
            </Stack>
        </TableCell>
        <TableCell sx={{whiteSpace: 'nowrap'}}>{formatDateTime(token.createdAt, 'YYYY/MM/DD')}</TableCell>
        <ExpiryCell expiresAt={token.expiresAt} now={now}/>
        <TableCell align="right">
            <Tooltip title="Revoke">
                <IconButton size="small" aria-label={`Revoke ${token.name}`} onClick={() => onRevoke(token)}>
                    <DeleteOutlineRounded fontSize="small"/>
                </IconButton>
            </Tooltip>
        </TableCell>
    </TableRow>;
};

export default AccessTokenRow;
