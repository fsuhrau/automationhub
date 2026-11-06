import * as React from 'react';
import {Link as RouterLink, Navigate, useLocation} from 'react-router-dom';
import {
    Alert,
    Box,
    Button,
    Card,
    CircularProgress,
    IconButton,
    InputAdornment,
    Link,
    Stack,
    TextField,
    Typography,
} from '@mui/material';
import CssBaseline from '@mui/material/CssBaseline';
import {GitHub, LoginRounded, VisibilityOffRounded, VisibilityRounded} from '@mui/icons-material';
import {CustomIcon} from '../../components/AppNavbar';
import ColorModeSelect from '../../shared-theme/ColorModeSelect';
import {useAuth} from '../../hooks/AuthProvider';

// AuthPage is the frame of the sign in and sign up pages: the product, a
// card in the middle, the color mode in the corner
export const AuthPage: React.FC<{ title: string, subtitle?: React.ReactNode, children: React.ReactNode }> = ({title, subtitle, children}) =>
    <>
        <CssBaseline enableColorScheme/>
        <ColorModeSelect size="small" sx={{position: 'fixed', top: '1rem', right: '1rem'}} inputProps={{'aria-label': 'Color mode'}}/>
        <Box sx={theme => ({
            minHeight: '100dvh', display: 'flex', alignItems: 'center', justifyContent: 'center', p: 2,
            backgroundImage: 'radial-gradient(ellipse at 50% 30%, hsl(210, 100%, 97%), hsl(0, 0%, 100%))',
            ...theme.applyStyles('dark', {backgroundImage: 'radial-gradient(at 50% 30%, hsla(210, 100%, 16%, 0.5), hsl(220, 30%, 5%))'}),
        })}>
            <Card variant="outlined" sx={{width: '100%', maxWidth: 420, p: {xs: 3, sm: 4}}}>
                <Stack spacing={1} sx={{mb: 3}}>
                    <Stack direction="row" spacing={1} alignItems="center" sx={{mb: 1}}>
                        <CustomIcon/>
                        <Typography variant="subtitle1" sx={{fontWeight: 600}}>Automation Hub</Typography>
                    </Stack>
                    <Typography component="h1" variant="h5">{title}</Typography>
                    {subtitle && <Typography variant="body2" color="text.secondary">{subtitle}</Typography>}
                </Stack>
                {children}
            </Card>
        </Box>
    </>;

const isEmail = (value: string) => /\S+@\S+\.\S+/.test(value);

// SignIn offers the way this master signs people in: the email and password
// form, or the provider's sign in. Afterwards it returns to the page that
// asked for it.
export default function SignIn() {
    const auth = useAuth();
    const location = useLocation();
    const from = (location.state as { from?: string } | null)?.from ?? '/';

    const [email, setEmail] = React.useState('');
    const [password, setPassword] = React.useState('');
    const [showPassword, setShowPassword] = React.useState(false);
    const [tried, setTried] = React.useState(false);
    const [error, setError] = React.useState<string | null>(null);
    const [busy, setBusy] = React.useState(false);

    if (auth.authState === 'checking') {
        return <AuthPage title="Sign in"><Stack alignItems="center"><CircularProgress/></Stack></AuthPage>;
    }
    if (auth.authState === 'authenticated') {
        return <Navigate to={from} replace/>;
    }
    if (auth.authState === 'disabled') {
        return <AuthPage title="No sign in needed" subtitle="This hub runs without accounts: everyone who reaches it can use it.">
            <Button fullWidth variant="contained" component={RouterLink} to={from}>Continue</Button>
        </AuthPage>;
    }

    const method = auth.signIn?.method;
    if ((method === 'github' || method === 'oauth2') && auth.signIn?.url) {
        return <AuthPage title="Sign in" subtitle="Run UI tests of your games on real devices.">
            <Button fullWidth size="large" variant="contained" href={auth.signIn.url}
                    startIcon={method === 'github' ? <GitHub/> : <LoginRounded/>}>
                {method === 'github' ? 'Sign in with GitHub' : 'Sign in with your company account'}
            </Button>
            <Typography variant="caption" color="text.secondary" component="p" sx={{mt: 2, textAlign: 'center'}}>
                Your account is created the first time you sign in.
            </Typography>
        </AuthPage>;
    }

    const emailError = tried && !isEmail(email) ? 'Enter the email address of your account.' : undefined;
    const passwordError = tried && !password ? 'Enter your password.' : undefined;

    const submit = (e: React.FormEvent) => {
        e.preventDefault();
        setTried(true);
        setError(null);
        if (!isEmail(email) || !password) return;
        setBusy(true);
        auth.loginAction({email: email.trim(), password}, message => {
            setError(message);
            setBusy(false);
        }, from);
    };

    return <AuthPage title="Sign in" subtitle="Run UI tests of your games on real devices.">
        <Box component="form" noValidate onSubmit={submit}>
            <Stack spacing={2}>
                {error && <Alert severity="error">{error}</Alert>}
                <TextField label="Email" type="email" name="email" autoComplete="email" autoFocus fullWidth
                           value={email} onChange={e => setEmail(e.target.value)} error={!!emailError} helperText={emailError}/>
                <TextField label="Password" type={showPassword ? 'text' : 'password'} name="password" autoComplete="current-password" fullWidth
                           value={password} onChange={e => setPassword(e.target.value)} error={!!passwordError} helperText={passwordError}
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
                    Sign in
                </Button>
                <Typography variant="caption" color="text.secondary" sx={{textAlign: 'center'}}>
                    Forgot your password? It can't be reset here; ask whoever runs this hub.
                </Typography>
            </Stack>
        </Box>
        <Typography variant="body2" sx={{textAlign: 'center', mt: 3}}>
            No account yet? <Link component={RouterLink} to="/signup" state={location.state}>Create one</Link>
        </Typography>
    </AuthPage>;
}
