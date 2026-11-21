import * as React from 'react';
import {Link as RouterLink, Navigate} from 'react-router-dom';
import {Button, CircularProgress, Stack} from '@mui/material';
import {useAuth} from '../../hooks/AuthProvider';
import {AuthPage} from './AuthPage';

interface AuthStateGateProps {
    // the page's title while the session is looked up
    title: string,
    // the title where no accounts are needed
    disabledTitle: string,
    // where to go when there is nothing to do here
    from: string,
    children: React.ReactNode,
}

// AuthStateGate lets the sign in and sign up forms show only when someone has
// to sign in: the session may still be checked, exist already, or not be needed.
export const AuthStateGate: React.FC<AuthStateGateProps> = ({title, disabledTitle, from, children}) => {
    const {authState} = useAuth();
    if (authState === 'checking') {
        return <AuthPage title={title}><Stack alignItems="center"><CircularProgress/></Stack></AuthPage>;
    }
    if (authState === 'authenticated') {
        return <Navigate to={from} replace={true}/>;
    }
    if (authState === 'disabled') {
        return <AuthPage title={disabledTitle} subtitle="This hub runs without accounts: everyone who reaches it can use it.">
            <Button fullWidth={true} variant="contained" component={RouterLink} to={from}>Continue</Button>
        </AuthPage>;
    }
    return <>{children}</>;
};
