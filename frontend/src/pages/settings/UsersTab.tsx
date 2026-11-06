import React, {useEffect} from 'react';
import {Avatar, Box, Chip, Stack, Table, TableBody, TableCell, TableHead, TableRow, Typography} from '@mui/material';
import {getUsers} from '../../services/settings.service';
import {HubStateActions} from '../../application/HubState';
import {useHubState} from '../../hooks/HubStateProvider';
import {useProjectContext} from '../../hooks/ProjectProvider';
import {useAuth} from '../../hooks/AuthProvider';
import {useError} from '../../ErrorProvider';
import {messageOf, Section} from './parts';

// UsersTab lists the people with an account on this hub; access is not per
// project yet, so it says so instead of offering roles it can't enforce
const UsersTab: React.FC = () => {
    const {state, dispatch} = useHubState();
    const {projectIdentifier} = useProjectContext();
    const {user: me} = useAuth();
    const {setError} = useError();

    useEffect(() => {
        getUsers(projectIdentifier).then(users => dispatch({type: HubStateActions.UsersUpdate, payload: users}))
            .catch(e => setError(messageOf(e)));
    }, [projectIdentifier]);

    const users = state.users ?? [];

    return <Section title="People" description="Everyone with an account on this hub. Every signed-in person can open every project; access per project isn't there yet.">
        {users.length === 0
            ? <Typography variant="body2" color="text.secondary">No accounts: this hub runs without sign-in, so everyone who reaches it can use it.</Typography>
            : <Box sx={{overflowX: 'auto'}}><Table size="small" aria-label="People">
                <TableHead><TableRow><TableCell>Name</TableCell><TableCell>Email</TableCell><TableCell>Role</TableCell></TableRow></TableHead>
                <TableBody>
                    {users.map(u => <TableRow key={u.id}>
                        <TableCell>
                            <Stack direction="row" spacing={1} alignItems="center">
                                <Avatar alt={u.name} sx={{width: 24, height: 24, fontSize: 12}}>{u.name?.[0]}</Avatar>
                                <span>{u.name}</span>
                                {me && me.name === u.name && <Chip size="small" variant="outlined" label="you"/>}
                            </Stack>
                        </TableCell>
                        <TableCell>{u.email}</TableCell>
                        <TableCell>{u.role || '–'}</TableCell>
                    </TableRow>)}
                </TableBody>
            </Table></Box>}
    </Section>;
};

export default UsersTab;
