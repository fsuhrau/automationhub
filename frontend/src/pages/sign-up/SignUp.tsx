import * as React from 'react';
import {Link as RouterLink, Navigate, useLocation} from 'react-router-dom';
import {Alert, Box, Button, CircularProgress, Link, Stack, TextField, Typography} from '@mui/material';
import {useAuth} from '../../hooks/AuthProvider';
import {routes} from '../../lib/routes';
import {AuthPage} from '../sign-in/AuthPage';
import {AuthStateGate} from '../sign-in/AuthStateGate';
import {PasswordField} from '../sign-in/PasswordField';
import {useAuthSubmit} from '../sign-in/useAuthSubmit';
import {hasErrors, MIN_PASSWORD, returnPath, signUpErrors} from '../sign-in/authModel';

const SignUpForm: React.FC<{ from: string }> = ({from}) => {
    const auth = useAuth();
    const location = useLocation();
    const [name, setName] = React.useState('');
    const [email, setEmail] = React.useState('');
    const [password, setPassword] = React.useState('');
    const errors = signUpErrors({name, email, password});
    const {tried, error, busy, submit} = useAuthSubmit(!hasErrors(errors),
        onError => auth.registerAction({name: name.trim(), email: email.trim(), password}, onError, from));
    const shown = (field: keyof typeof errors): string | undefined => tried ? errors[field] : undefined;

    return <AuthPage title="Create an account" subtitle="To run tests, control devices and get a personal token for Claude Code.">
        <Box component="form" noValidate={true} onSubmit={submit}>
            <Stack spacing={2}>
                {error && <Alert severity="error">{error}</Alert>}
                <TextField label="Name" name="name" autoComplete="name" autoFocus={true} fullWidth={true} value={name} onChange={e => setName(e.target.value)}
                    error={!!shown('name')} helperText={shown('name')}/>
                <TextField label="Email" type="email" name="email" autoComplete="email" fullWidth={true} value={email} onChange={e => setEmail(e.target.value)}
                    error={!!shown('email')} helperText={shown('email')}/>
                <PasswordField value={password} onChange={setPassword} autoComplete="new-password" error={shown('password')}
                    helperText={`At least ${MIN_PASSWORD} characters.`}/>
                <Button type="submit" fullWidth={true} size="large" variant="contained" disabled={busy}
                    startIcon={busy ? <CircularProgress size={16} color="inherit"/> : undefined}>
                    Create account
                </Button>
            </Stack>
        </Box>
        <Typography variant="body2" sx={{textAlign: 'center', mt: 3}}>
            Have an account? <Link component={RouterLink} to={routes.login()} state={location.state}>Sign in</Link>
        </Typography>
    </AuthPage>;
};

const SignUpMethodPage: React.FC<{ from: string }> = ({from}) => {
    const {signIn} = useAuth();
    const location = useLocation();
    // accounts come from GitHub or the company's sign in
    if (signIn && signIn.method !== 'password') return <Navigate to={routes.login()} replace={true} state={location.state}/>;
    return <SignUpForm from={from}/>;
};

// SignUp creates an account with a password, on masters that sign people
// in that way; others send people to their sign in
export default function SignUp(): React.ReactElement {
    const from = returnPath(useLocation().state, routes.projects());
    return <AuthStateGate title="Create an account" disabledTitle="No account needed" from={from}>
        <SignUpMethodPage from={from}/>
    </AuthStateGate>;
}
