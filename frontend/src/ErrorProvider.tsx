import {Alert} from '@mui/material';
import React, {createContext, ReactNode, useContext, useEffect, useState} from 'react';
import {errorMessage} from './lib/errorMessage';

type ErrorValue = string | object | null;

interface ErrorContextType {
    error: ErrorValue;
    setError: (message: ErrorValue) => void;
}

const ErrorContext = createContext<ErrorContextType | undefined>(undefined);

const HIDE_AFTER_MS = 8000;
const NO_MESSAGE = 'An unknown error occurred';
// what errorMessage makes of an object without any message
const OBJECT_TEXT = '[object Object]';

const textOf = (error: string | object): string => {
    const text = errorMessage(error);
    return text === OBJECT_TEXT ? NO_MESSAGE : text;
};

const alertStyle: React.CSSProperties = {
    // fixed, so it shows wherever the page is scrolled
    position: 'fixed',
    zIndex: 1500,
    bottom: '5%',
    left: '50%',
    transform: 'translateX(-50%)',
    width: '80%',
    maxWidth: '600px',
};

// an error hides itself after a while
const useAutoHide = (error: ErrorValue, hide: () => void): void => {
    useEffect(() => {
        if (!error) return;
        const timer = setTimeout(hide, HIDE_AFTER_MS);
        return () => clearTimeout(timer);
    }, [error, hide]);
};

export const ErrorProvider = ({children}: { children: ReactNode }): React.ReactElement => {
    const [error, setError] = useState<ErrorValue>(null);
    const clear = React.useCallback(() => setError(null), []);
    useAutoHide(error, clear);

    return (
        <ErrorContext.Provider value={{error, setError}}>
            {children}
            {error && (
                <Alert severity="error" color="error" role="alert" onClose={clear} style={alertStyle}>
                    {textOf(error)}
                </Alert>
            )}
        </ErrorContext.Provider>
    );
};

export const useError = (): ErrorContextType => {
    const context = useContext(ErrorContext);
    if (context === undefined) {
        throw new Error('useError must be used within an ErrorProvider');
    }
    return context;
};
