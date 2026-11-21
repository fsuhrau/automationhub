import React, {createContext, ReactNode, useCallback, useContext, useEffect, useMemo, useState} from "react";
import {useNavigate} from "react-router-dom";
import {IUser} from "../types/user";
import {AuthClient, createAuthClient, CredentialsResult, LoginData, RegisterData, SessionResult, SignInMethod} from "../app/authClient";
import {routes} from "../lib/routes";

export type {LoginData, RegisterData, SignInMethod};

// checking: session lookup in flight; anonymous: auth is enabled and nobody
// is logged in; disabled: the master runs without authentication
export type AuthState = 'checking' | 'authenticated' | 'anonymous' | 'disabled';

type OnError = (error: string) => void;

type AuthContextProps = {
    user: IUser | null,
    authState: AuthState,
    // known once a signed out session was checked; older masters don't say
    signIn: SignInMethod | null,
    // signs in and opens `to` (the page that asked for it) or the project list
    loginAction: (d: LoginData, onError: OnError, to?: string) => Promise<void>,
    // creates the account, signs in and opens `to` or the project list
    registerAction: (d: RegisterData, onError: OnError, to?: string) => Promise<void>,
    logOut: () => Promise<void>,
};

interface AuthProviderProps {
    children: ReactNode;
    // tests inject a fake
    client?: AuthClient;
}

const AuthContext = createContext<AuthContextProps>({
    user: null,
    authState: 'checking',
    signIn: null,
    loginAction: async () => undefined,
    registerAction: async () => undefined,
    logOut: async () => undefined,
});

const defaultClient = createAuthClient();

interface Session {
    user: IUser | null,
    authState: AuthState,
    signIn: SignInMethod | null,
}

const CHECKING: Session = {user: null, authState: 'checking', signIn: null};

const sessionOf = (result: SessionResult, previous: Session): Session => {
    if (result.state === 'authenticated') return {...previous, user: result.user, authState: 'authenticated'};
    if (result.state === 'anonymous') return {...previous, authState: 'anonymous', signIn: result.signIn};
    return {...previous, authState: 'disabled'};
};

// looks the session up while nobody is signed in
const useSession = (client: AuthClient): [Session, (s: Session) => void] => {
    const [session, setSession] = useState<Session>(CHECKING);
    const signedIn = session.user !== null;
    useEffect(() => {
        if (signedIn) return;
        let active = true;
        client.session().then(result => {
            if (active) setSession(previous => sessionOf(result, previous));
        });
        return () => {
            active = false;
        };
    }, [signedIn, client]);
    return [session, setSession];
};

export const AuthProvider: React.FC<AuthProviderProps> = ({children, client = defaultClient}) => {
    const [session, setSession] = useSession(client);
    const navigate = useNavigate();

    const submit = useCallback(async (send: Promise<CredentialsResult>, onError: OnError, to: string): Promise<void> => {
        const result = await send;
        if (!result.ok) {
            onError(result.error);
            return;
        }
        setSession({user: result.user, authState: 'authenticated', signIn: null});
        navigate(to, {replace: true});
    }, [navigate, setSession]);

    const logOut = useCallback(async (): Promise<void> => {
        if (!await client.logout()) return;
        setSession({user: null, authState: 'anonymous', signIn: null});
        navigate(routes.login());
    }, [client, navigate, setSession]);

    const value = useMemo((): AuthContextProps => ({
        ...session,
        loginAction: (data, onError, to = routes.projects()) => submit(client.login(data), onError, to),
        registerAction: (data, onError, to = routes.projects()) => submit(client.register(data), onError, to),
        logOut,
    }), [session, client, submit, logOut]);

    return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};

export const useAuth = (): AuthContextProps => useContext(AuthContext);
