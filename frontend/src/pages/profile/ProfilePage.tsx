import React from 'react';
import {Alert, Box, Link, Stack, Typography} from '@mui/material';
import CssBaseline from '@mui/material/CssBaseline';
import {Link as RouterLink} from 'react-router-dom';
import {ConfirmDialog} from '../../components/common';
import TopBar from '../../components/TopBar';
import {useAuth} from '../../hooks/AuthProvider';
import {useNow} from '../../hooks/useNow';
import {routes} from '../../lib/routes';
import {useConfirmAction} from '../settings/useConfirmAction';
import AccountSection from './sections/AccountSection';
import AppearanceSection from './sections/AppearanceSection';
import ConnectGuide from './sections/ConnectGuide';
import PersonalTokensSection from './sections/PersonalTokensSection';
import {useProfile} from './useProfile';

const NOW_REFRESH_MS = 60_000;

const NoSignIn: React.FC = () => <Alert severity="info" sx={{mb: 3}}>
    This hub runs without sign-in, so there are no accounts or personal tokens: everyone who reaches it can use it.
    For scripts, CI and Claude Code, create a project token in a project's settings under Access tokens.
    {' '}<Link component={RouterLink} to={routes.projects()}>Pick a project</Link>.
</Alert>;

// ProfilePage is about the signed-in person: who they are, how the hub looks
// for them, and the personal tokens that act as them
const ProfilePage: React.FC = () => {
    const auth = useAuth();
    const data = useProfile(auth.authState === 'authenticated');
    const revoking = useConfirmAction(data.revoke);
    const now = useNow(NOW_REFRESH_MS);

    return (
        <>
            <CssBaseline enableColorScheme={true}/>
            <TopBar/>
            <Box sx={{maxWidth: 1000, mx: 'auto', px: 2, py: 4}}>
                <Typography variant="h5" component="h1">Profile</Typography>
                <Typography variant="body2" color="text.secondary" sx={{mb: 3}}>
                    Your account, how the hub looks for you, and the tokens that act as you.
                </Typography>
                {auth.authState === 'disabled' && <NoSignIn/>}
                {data.profileError && <Alert severity="warning" sx={{mb: 3}}>{data.profileError}</Alert>}
                <Stack spacing={3}>
                    {auth.user && <AccountSection profile={data.profile} user={auth.user} onSignOut={() => auth.logOut()}/>}
                    <AppearanceSection/>
                    {data.profile && <>
                        <PersonalTokensSection tokens={data.tokens} created={data.created} onDismissCreated={data.dismissCreated}
                            busy={data.create.busy} onCreate={data.create.run} now={now} onRevoke={revoking.ask}/>
                        <ConnectGuide projects={data.projects} token={data.created} where={window.location}/>
                    </>}
                </Stack>
            </Box>
            <ConfirmDialog open={!!revoking.target} title={`Revoke ${revoking.target?.name}?`} confirm="Revoke token"
                busy={revoking.busy} onConfirm={revoking.confirm} onClose={revoking.cancel}>
                Claude Code, scripts and anything else using it get "unauthorized" from now on.
            </ConfirmDialog>
        </>
    );
};

export default ProfilePage;
