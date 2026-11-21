import * as React from 'react';
import {Link as RouterLink, useLocation} from 'react-router-dom';
import {Alert, Box, Button, CircularProgress, Link, Stack, TextField, Typography} from '@mui/material';
import {GitHub, LoginRounded} from '@mui/icons-material';
import {useAuth} from '../../hooks/AuthProvider';
import {routes} from '../../lib/routes';
import {AuthPage} from './AuthPage';
import {AuthStateGate} from './AuthStateGate';
import {PasswordField} from './PasswordField';
import {useAuthSubmit} from './useAuthSubmit';
import {hasErrors, returnPath, signInErrors} from './authModel';

const SUBTITLE = 'Run UI tests of your games on real devices.';

// the sign in of a provider (GitHub or the company's)
const ProviderSignIn: React.FC<{ method: 'github' | 'oauth2', url: string }> = ({method, url}) => (
    <AuthPage title="Sign in" subtitle={SUBTITLE}>
        <Button fullWidth={true} size="large" variant="contained" href={url} startIcon={method === 'github' ? <GitHub/> : <LoginRounded/>}>
            {method === 'github' ? 'Sign in with GitHub' : 'Sign in with your company account'}
        </Button>
        <Typography variant="caption" color="text.secondary" component="p" sx={{mt: 2, textAlign: 'center'}}>
            Your account is created the first time you sign in.
        </Typography>
    </AuthPage>
);

// the email and password form
const PasswordSignIn: React.FC<{ from: string }> = ({from}) => {
    const auth = useAuth();
    const location = useLocation();
    const [email, setEmail] = React.useState('');
    const [password, setPassword] = React.useState('');
    const errors = signInErrors({email, password});
    const {tried, error, busy, submit} = useAuthSubmit(!hasErrors(errors),
        onError => auth.loginAction({email: email.trim(), password}, onError, from));

    return <AuthPage title="Sign in" subtitle={SUBTITLE}>
        <Box component="form" noValidate={true} onSubmit={submit}>
            <Stack spacing={2}>
                {error && <Alert severity="error">{error}</Alert>}
                <TextField label="Email" type="email" name="email" autoComplete="email" autoFocus={true} fullWidth={true} value={email}
                    onChange={e => setEmail(e.target.value)} error={tried && !!errors.email} helperText={tried ? errors.email : undefined}/>
                <PasswordField value={password} onChange={setPassword} autoComplete="current-password" error={tried ? errors.password : undefined}/>
                <Button type="submit" fullWidth={true} size="large" variant="contained" disabled={busy}
                    startIcon={busy ? <CircularProgress size={16} color="inherit"/> : undefined}>
                    Sign in
                </Button>
                <Typography variant="caption" color="text.secondary" sx={{textAlign: 'center'}}>
                    Forgot your password? It can't be reset here; ask whoever runs this hub.
                </Typography>
            </Stack>
        </Box>
        <Typography variant="body2" sx={{textAlign: 'center', mt: 3}}>
            No account yet? <Link component={RouterLink} to={routes.signup()} state={location.state}>Create one</Link>
        </Typography>
    </AuthPage>;
};

const SignInMethodPage: React.FC<{ from: string }> = ({from}) => {
    const {signIn} = useAuth();
    const provider = signIn?.method === 'github' || signIn?.method === 'oauth2' ? signIn.method : null;
    return provider && signIn?.url ? <ProviderSignIn method={provider} url={signIn.url}/> : <PasswordSignIn from={from}/>;
};

// SignIn offers the way this master signs people in: the email and password
// form, or the provider's sign in. Afterwards it returns to the page that
// asked for it.
export default function SignIn(): React.ReactElement {
    const from = returnPath(useLocation().state, routes.projects());
    return <AuthStateGate title="Sign in" disabledTitle="No sign in needed" from={from}>
        <SignInMethodPage from={from}/>
    </AuthStateGate>;
}
