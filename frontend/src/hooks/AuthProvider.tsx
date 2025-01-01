import React, {useContext, createContext, useState, useEffect, ReactNode, Dispatch} from "react";
import { useNavigate } from "react-router-dom";
import {HubState, HubStateAction} from "../application/HubState";

export interface IUser {
    name: string,
    email: string,
    login: string,
    company: string,
    role: string,
    url: string,
    avatar: string
}

export interface RegisterData {
    name: string,
    email: string,
    password: string,
}

export interface LoginData {
    email: string,
    password: string,
}

type AuthContextProps = {
    user: IUser | null,
    loginAction: (d: LoginData) => {},
    registerAction: (d: RegisterData) => {},
    logOut: () => {},
};


interface AuthProviderProps {
    children: ReactNode;
}

const AuthContext = createContext<AuthContextProps>({
    user: null,
    loginAction: async () => {},
    registerAction: async () => {},
    logOut: async () => {},
});

export const AuthProvider: React.FC<AuthProviderProps> = ({ children }) => {

    const [user, setUser] = useState<IUser|null>(null);
    const navigate = useNavigate();

    useEffect(() => {
        const checkSession = async () => {
            try {
                const response = await fetch("/api/auth/session", {
                    method: "GET",
                    credentials: "include", // Include cookies in the request
                });
                const res = await response.json();
                if (res.user) {
                    setUser(res.user);
                    navigate("/");
                }
            } catch (err) {
                console.error(err);
            }
        };

        if (user == null) {
            checkSession();
        }
    }, [user, navigate]);

    const loginAction = async (data: LoginData) => {
        try {
            const response = await fetch("/api/auth/login", {
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
                navigate("/");
                return;
            }
        } catch (err) {
            console.error(err);
        }
    };

    const registerAction = async (data: RegisterData) => {
        try {
            const response = await fetch("/api/auth/register", {
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
                navigate("/");
                return;
            }
        } catch (err) {
            console.error(err);
        }
    };

    const logOut = async () => {
        try {
            const res = await fetch("/api/auth/logout", {
                method: "POST",
                credentials: "include", // Include cookies in the request
            });
            setUser(null);
            navigate("/login");
        } catch (err) {
            console.error(err);
        }
    };

    return (
        <AuthContext.Provider value={{ user: user, loginAction: loginAction, registerAction: registerAction, logOut: logOut }}>
            {children}
        </AuthContext.Provider>
    );
};

export const useAuth = () : AuthContextProps => {
    return useContext(AuthContext);
};