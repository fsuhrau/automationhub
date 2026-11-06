import * as React from 'react';
import {Link as RouterLink, Navigate, useLocation} from 'react-router-dom';
import {Alert, Box, Button, CircularProgress, IconButton, InputAdornment, Link, Stack, TextField, Typography} from '@mui/material';
import {VisibilityOffRounded, VisibilityRounded} from '@mui/icons-material';
import {useAuth} from '../../hooks/AuthProvider';
import {AuthPage} from '../sign-in/SignIn';

// the master's rule for new passwords
const MIN_PASSWORD = 8;
const isEmail = (value: string) => /\S+@\S+\.\S+/.test(value);

// SignUp creates an account with a password, on masters that sign people
// in that way; others send people to their sign in
export default function SignUp() {
    const auth = useAuth();
    const location = useLocation();
    const from = (location.state as { from?: string } | null)?.from ?? '/';

    const [name, setName] = React.useState('');
    const [email, setEmail] = React.useState('');
    const [password, setPassword] = React.useState('');
    const [showPassword, setShowPassword] = React.useState(false);
    const [tried, setTried] = React.useState(false);
    const [error, setError] = React.useState<string | null>(null);
    const [busy, setBusy] = React.useState(false);

    if (auth.authState === 'checking') {
        return <AuthPage title="Create an account"><Stack alignItems="center"><CircularProgress/></Stack></AuthPage>;
    }
    if (auth.authState === 'authenticated') {
        return <Navigate to={from} replace/>;
    }
    if (auth.authState === 'disabled') {
        return <AuthPage title="No account needed" subtitle="This hub runs without accounts: everyone who reaches it can use it.">
            <Button fullWidth variant="contained" component={RouterLink} to={from}>Continue</Button>
        </AuthPage>;
    }
    if (auth.signIn && auth.signIn.method !== 'password') {
        // accounts come from GitHub or the company's sign in
        return <Navigate to="/login" replace state={location.state}/>;
    }

    const nameError = tried && !name.trim() ? 'Enter your name; others see it on devices you control.' : undefined;
    const emailError = tried && !isEmail(email) ? 'Enter an email address; you sign in with it.' : undefined;
    const passwordError = tried && password.length < MIN_PASSWORD ? `At least ${MIN_PASSWORD} characters.` : undefined;

    const submit = (e: React.FormEvent) => {
        e.preventDefault();
        setTried(true);
        setError(null);
        if (!name.trim() || !isEmail(email) || password.length < MIN_PASSWORD) return;
        setBusy(true);
        auth.registerAction({name: name.trim(), email: email.trim(), password}, message => {
            setError(message);
            setBusy(false);
        }, from);
    };

    return <AuthPage title="Create an account" subtitle="To run tests, control devices and get a personal token for Claude Code.">
        <Box component="form" noValidate onSubmit={submit}>
            <Stack spacing={2}>
                {error && <Alert severity="error">{error}</Alert>}
                <TextField label="Name" name="name" autoComplete="name" autoFocus fullWidth value={name} onChange={e => setName(e.target.value)}
                           error={!!nameError} helperText={nameError}/>
                <TextField label="Email" type="email" name="email" autoComplete="email" fullWidth value={email} onChange={e => setEmail(e.target.value)}
                           error={!!emailError} helperText={emailError}/>
                <TextField label="Password" type={showPassword ? 'text' : 'password'} name="password" autoComplete="new-password" fullWidth
                           value={password} onChange={e => setPassword(e.target.value)} error={!!passwordError}
                           helperText={passwordError ?? `At least ${MIN_PASSWORD} characters.`}
                           slotProps={{
                               input: {
                                   endAdornment: <InputAdornment position="end">
                                       <IconButton size="small" edge="end" aria-label={showPassword ? 'Hide the password' : 'Show the password'}
                                                   onClick={() => setShowPassword(v => !v)}>
                                           {showPassword ? <VisibilityOffRounded fontSize="small"/> : <VisibilityRounded fontSize="small"/>}
                                       </IconButton>
                                   </InputAdornment>,
                               },
                           }}/>
                <Button type="submit" fullWidth size="large" variant="contained" disabled={busy}
                        startIcon={busy ? <CircularProgress size={16} color="inherit"/> : undefined}>
                    Create account
                </Button>
            </Stack>
        </Box>
        <Typography variant="body2" sx={{textAlign: 'center', mt: 3}}>
            Have an account? <Link component={RouterLink} to="/login" state={location.state}>Sign in</Link>
        </Typography>
    </AuthPage>;
}
