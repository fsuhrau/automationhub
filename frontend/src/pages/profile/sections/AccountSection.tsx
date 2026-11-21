import React from 'react';
import {Avatar, Box, Button, Stack, Typography} from '@mui/material';
import {LogoutRounded} from '@mui/icons-material';
import {Section} from '../../../components/common';
import {IProfile} from '../../../types/profile.token';
import {IUser} from '../../../types/user';
import {accountDetails} from '../profileModel';

export interface AccountSectionProps {
    profile: IProfile | null,
    user: IUser,
    onSignOut: () => void,
}

// AccountSection says who is signed in, and signs out
const AccountSection: React.FC<AccountSectionProps> = ({profile, user, onSignOut}) => {
    const {name, email, login} = accountDetails(profile, user);
    return <Section title="Account" action={
        <Button variant="outlined" size="small" startIcon={<LogoutRounded/>} onClick={onSignOut}>Sign out</Button>}>
        <Stack direction="row" spacing={2} alignItems="center">
            <Avatar src={user.avatar || undefined} alt={name} sx={{width: 56, height: 56}}>{name[0]}</Avatar>
            <Box sx={{minWidth: 0}}>
                <Typography variant="h6">{name}</Typography>
                {email && <Typography variant="body2" color="text.secondary">{email}</Typography>}
                {login && <Typography variant="caption" color="text.secondary">signed in as {login}</Typography>}
            </Box>
        </Stack>
    </Section>;
};

export default AccountSection;
