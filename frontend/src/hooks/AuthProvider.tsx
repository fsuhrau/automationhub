import React, {createContext, ReactNode, useContext, useEffect, useState} from "react";
import {useNavigate} from "react-router-dom";
import {IUser} from "../types/user"


export interface RegisterData {
    name: string,
    email: string,
    password: string,
}

export interface LoginData {
    email: string,
    password: string,
}

// checking: session lookup in flight; anonymous: auth is enabled and nobody
// is logged in; disabled: the master runs without authentication
export type AuthState = 'checking' | 'authenticated' | 'anonymous' | 'disabled';

type AuthContextProps = {
    user: IUser | null,
    authState: AuthState,
    loginAction: (d: LoginData, onError: (error: string) => void) => {},
    registerAction: (d: RegisterData, onError: (error: string) => void) => {},
    logOut: () => {},
};


interface AuthProviderProps {
    children: ReactNode;
}

const AuthContext = createContext<AuthContextProps>({
    user: null,
    authState: 'checking',
    loginAction: async () => {
    },
    registerAction: async () => {
    },
    logOut: async () => {
    },
});

export const AuthProvider: React.FC<AuthProviderProps> = ({children}) => {

    const [user, setUser] = useState<IUser | null>(null);
    const [authState, setAuthState] = useState<AuthState>('checking');
    const navigate = useNavigate();

    useEffect(() => {
        const checkSession = async () => {
            try {
                const response = await fetch("/auth/session", {
                    method: "GET",
                    credentials: "include", // Include cookies in the request
                });
                if (response.status === 401) {
                    setAuthState('anonymous');
                    return;
                }
                const isJson = (response.headers.get('content-type') ?? '').includes('application/json');
                if (!response.ok || !isJson) {
                    // no /auth routes: the master runs without authentication
                    setAuthState('disabled');
                    return;
                }
                const res = await response.json();
                if (res.user) {
                    setUser(res.user);
                    setAuthState('authenticated');
                } else {
                    setAuthState('anonymous');
                }
            } catch (err) {
                // network problem: don't lock the user out, API calls still get 401
                console.error(err);
                setAuthState('disabled');
            }
        };

        if (user == null) {
            checkSession();
        }
    }, [user, navigate]);

    const loginAction = async (data: LoginData, onError: (error: string) => void) => {
        try {
            const response = await fetch("/auth/login", {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                },
                body: JSON.stringify(data),
                credentials: "include", // Include cookies in the request
            });
            const res = await response.json();
            if (res.user) {
                setUser(res.user);
                setAuthState('authenticated');
                navigate("/");
                return;
            }
            onError(res.error ?? "Invalid email or password");
        } catch (err) {
            console.error(err);
            onError("Sign in failed: the server could not be reached.");
        }
    };

    const registerAction = async (data: RegisterData, onError: (error: string) => void) => {
        try {
            const response = await fetch("/auth/register", {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                },
                body: JSON.stringify(data),
                credentials: "include", // Include cookies in the request
            });
            const isJson = (response.headers.get('content-type') ?? '').includes('application/json');
            if (!isJson) {
                // without password auth the master has no /auth routes
                onError("Registration is not available: the master is not configured for password authentication.");
                return;
            }
            const res = await response.json();
            if (res.user) {
                setUser(res.user);
                setAuthState('authenticated');
                navigate("/");
                return;
            }
            onError(res.error ?? `Registration failed (${response.status})`);
        } catch (err) {
            console.error(err);
            onError("Registration failed: the server could not be reached.");
        }
    };

    const logOut = async () => {
        try {
            const res = await fetch("/auth/logout", {
                method: "POST",
                credentials: "include", // Include cookies in the request
            });
            setUser(null);
            setAuthState('anonymous');
            navigate("/login");
        } catch (err) {
            console.error(err);
        }
    };

    return (
        <AuthContext.Provider
            value={{user: user, authState: authState, loginAction: loginAction, registerAction: registerAction, logOut: logOut}}>
            {children}
        </AuthContext.Provider>
    );
};

export const useAuth = (): AuthContextProps => {
    return useContext(AuthContext);
};