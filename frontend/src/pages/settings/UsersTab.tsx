import React from 'react';
import {Avatar, Box, Chip, Stack, Table, TableBody, TableCell, TableHead, TableRow, Typography} from '@mui/material';
import {HubStateActions} from '../../application/HubState';
import {Section} from '../../components/common';
import {useHubState} from '../../hooks/HubStateProvider';
import {useProjectContext} from '../../hooks/ProjectProvider';
import {useAuth} from '../../hooks/AuthProvider';
import {useSettingsService} from '../../services/settings.service';
import {IUserData} from '../../types/user';
import {useHubLoad} from './useHubLoad';

const DESCRIPTION = "Everyone with an account on this hub. Every signed-in person can open every project; access per project isn't there yet.";

const UserRow: React.FC<{ user: IUserData, isMe: boolean }> = ({user, isMe}) =>
    <TableRow>
        <TableCell>
            <Stack direction="row" spacing={1} alignItems="center">
                <Avatar alt={user.name} sx={{width: 24, height: 24, fontSize: 12}}>{user.name?.[0]}</Avatar>
                <span>{user.name}</span>
                {isMe && <Chip size="small" variant="outlined" label="you"/>}
            </Stack>
        </TableCell>
        <TableCell>{user.email}</TableCell>
        <TableCell>{user.role || '–'}</TableCell>
    </TableRow>;

// UsersTab lists the people with an account on this hub; access is not per
// project yet, so it says so instead of offering roles it can't enforce
const UsersTab: React.FC = () => {
    const {state} = useHubState();
    const {projectIdentifier} = useProjectContext();
    const {user: me} = useAuth();
    const service = useSettingsService();
    useHubLoad(() => service.users(projectIdentifier), payload => ({type: HubStateActions.UsersUpdate, payload}), [projectIdentifier]);
    const users = state.users ?? [];

    return <Section title="People" description={DESCRIPTION}>
        {users.length === 0
            ? <Typography variant="body2" color="text.secondary">No accounts: this hub runs without sign-in, so everyone who reaches it can use it.</Typography>
            : <Box sx={{overflowX: 'auto'}}><Table size="small" aria-label="People">
                <TableHead><TableRow><TableCell>Name</TableCell><TableCell>Email</TableCell><TableCell>Role</TableCell></TableRow></TableHead>
                <TableBody>{users.map(u => <UserRow key={u.id} user={u} isMe={!!me && me.name === u.name}/>)}</TableBody>
            </Table></Box>}
    </Section>;
};

export default UsersTab;
