import * as React from 'react';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Checkbox from '@mui/material/Checkbox';
import CssBaseline from '@mui/material/CssBaseline';
import FormControlLabel from '@mui/material/FormControlLabel';
import Divider from '@mui/material/Divider';
import FormLabel from '@mui/material/FormLabel';
import FormControl from '@mui/material/FormControl';
import Link from '@mui/material/Link';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';
import Stack from '@mui/material/Stack';
import MuiCard from '@mui/material/Card';
import {styled} from '@mui/material/styles';
import ForgotPassword from './ForgotPassword';
import {AutomationHubIcon, FacebookIcon, GoogleIcon} from './CustomIcons';
import ColorModeSelect from '../../shared-theme/ColorModeSelect';
import {useAuth} from "../../hooks/AuthProvider";
import {useError} from "../../ErrorProvider";

const Card = styled(MuiCard)(({theme}) => ({
    display: 'flex',
    flexDirection: 'column',
    alignSelf: 'center',
    width: '100%',
    padding: theme.spacing(4),
    gap: theme.spacing(2),
    margin: 'auto',
    [theme.breakpoints.up('sm')]: {
        maxWidth: '450px',
    },
    boxShadow:
        'hsla(220, 30%, 5%, 0.05) 0px 5px 15px 0px, hsla(220, 25%, 10%, 0.05) 0px 15px 35px -5px',
    ...theme.applyStyles('dark', {
        boxShadow:
            'hsla(220, 30%, 5%, 0.5) 0px 5px 15px 0px, hsla(220, 25%, 10%, 0.08) 0px 15px 35px -5px',
    }),
}));

const SignInContainer = styled(Stack)(({theme}) => ({
    height: 'calc((1 - var(--template-frame-height, 0)) * 100dvh)',
    minHeight: '100%',
    padding: theme.spacing(2),
    [theme.breakpoints.up('sm')]: {
        padding: theme.spacing(4),
    },
    '&::before': {
        content: '""',
        display: 'block',
        position: 'absolute',
        zIndex: -1,
        inset: 0,
        backgroundImage:
            'radial-gradient(ellipse at 50% 50%, hsl(210, 100%, 97%), hsl(0, 0%, 100%))',
        backgroundRepeat: 'no-repeat',
        ...theme.applyStyles('dark', {
            backgroundImage:
                'radial-gradient(at 50% 50%, hsla(210, 100%, 16%, 0.5), hsl(220, 30%, 5%))',
        }),
    },
}));

export default function SignIn(props: { disableCustomTheme?: boolean }) {

    const auth = useAuth();
    const {error, setError} = useError();

    const [uiState, setUiState] = React.useState<{
        emailError: boolean,
        emailErrorMessage: string,
        passwordError: boolean,
        passwordErrorMessage: string,
        open: boolean
    }>({
        emailError: false,
        emailErrorMessage: '',
        passwordError: false,
        passwordErrorMessage: '',
        open: false
    })


    const handleClickOpen = () => {
        setUiState(prevState => ({...prevState, open: true}));
    };

    const handleClose = () => {
        setUiState(prevState => ({...prevState, open: false}));
    };

    const handleSubmit = (event: React.FormEvent<HTMLFormElement>) => {
        setError(null);

        event.preventDefault();
        if (uiState.emailError || uiState.passwordError) {
            return;
        }

        const data = new FormData(event.currentTarget);
        auth.loginAction({email: data.get('email') as string, password: data.get('password') as string}, error => {
            setError(error);
        })
    };

    const validateInputs = () => {
        const email = document.getElementById('email') as HTMLInputElement;
        const password = document.getElementById('password') as HTMLInputElement;

        let isValid = true;

        if (!email.value || !/\S+@\S+\.\S+/.test(email.value)) {
            setUiState(prevState => ({...prevState, emailError: true, emailErrorMessage: 'Please enter a valid email address.'}));
            isValid = false;
        } else {
            setUiState(prevState => ({...prevState, emailError: false, emailErrorMessage: ''}));
        }

        if (!password.value || password.value.length < 6) {
            setUiState(prevState => ({...prevState, passwordError: true, passwordErrorMessage: 'Password must be at least 6 characters long.'}));
            isValid = false;
        } else {
            setUiState(prevState => ({...prevState, passwordError: false, passwordErrorMessage: ''}));
        }

        return isValid;
    };

    return (
        <>
            <CssBaseline enableColorScheme/>
            <SignInContainer direction="column" justifyContent="space-between">
                <ColorModeSelect sx={{position: 'fixed', top: '1rem', right: '1rem'}}/>
                <Card variant="outlined">
                    <AutomationHubIcon/>
                    <Typography
                        component="h1"
                        variant="h4"
                        sx={{width: '100%', fontSize: 'clamp(2rem, 10vw, 2.15rem)'}}
                    >
                        Sign in
                    </Typography>
                    <Box
                        component="form"
                        onSubmit={handleSubmit}
                        noValidate
                        sx={{
                            display: 'flex',
                            flexDirection: 'column',
                            width: '100%',
                            gap: 2,
                        }}
                    >
                        <FormControl>
                            <FormLabel htmlFor="email">Email</FormLabel>
                            <TextField
                                error={uiState.emailError}
                                helperText={uiState.emailErrorMessage}
                                id="email"
                                type="email"
                                name="email"
                                placeholder="your@email.com"
                                autoComplete="email"
                                autoFocus
                                required
                                fullWidth
                                variant="outlined"
                                color={uiState.emailError ? 'error' : 'primary'}
                            />
                        </FormControl>
                        <FormControl>
                            <FormLabel htmlFor="password">Password</FormLabel>
                            <TextField
                                error={uiState.passwordError}
                                helperText={uiState.passwordErrorMessage}
                                name="password"
                                placeholder="••••••"
                                type="password"
                                id="password"
                                autoComplete="current-password"
                                autoFocus
                                required
                                fullWidth
                                variant="outlined"
                                color={uiState.passwordError ? 'error' : 'primary'}
                            />
                        </FormControl>
                        <FormControlLabel
                            control={<Checkbox value="remember" color="primary"/>}
                            label="Remember me"
                        />
                        <ForgotPassword open={uiState.open} handleClose={handleClose}/>
                        <Button
                            type="submit"
                            fullWidth
                            variant="contained"
                            onClick={validateInputs}
                        >
                            Sign in
                        </Button>
                        <Link
                            component="button"
                            type="button"
                            onClick={handleClickOpen}
                            variant="body2"
                            sx={{alignSelf: 'center'}}
                        >
                            Forgot your password?
                        </Link>
                    </Box>
                    <Divider>or</Divider>
                    <Box sx={{display: 'flex', flexDirection: 'column', gap: 2}}>
                        <Button
                            fullWidth
                            variant="outlined"
                            onClick={() => alert('Sign in with Google')}
                            startIcon={<GoogleIcon/>}
                        >
                            Sign in with Google
                        </Button>
                        <Button
                            fullWidth
                            variant="outlined"
                            onClick={() => alert('Sign in with Facebook')}
                            startIcon={<FacebookIcon/>}
                        >
                            Sign in with Facebook
                        </Button>
                        <Typography sx={{textAlign: 'center'}}>
                            Don&apos;t have an account?{' '}
                            <Link
                                href="/signup"
                                variant="body2"
                                sx={{alignSelf: 'center'}}
                            >
                                Sign up
                            </Link>
                        </Typography>
                    </Box>
                </Card>
            </SignInContainer>
        </>
    );
}
