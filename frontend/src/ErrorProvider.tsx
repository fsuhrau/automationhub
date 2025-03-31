import { Alert } from '@mui/material';
    import React, { createContext, ReactNode, useContext, useState } from 'react';

    interface ErrorContextType {
        error: string | object | null;
        setError: (message: string | object | null) => void;
    }

    const ErrorContext = createContext<ErrorContextType | undefined>(undefined);

    export const ErrorProvider = ({ children }: { children: ReactNode }) => {
        const [error, setError] = useState<string | object | null>(null);

        React.useEffect(() => {
            if (error) {
                const timer = setTimeout(() => {
                    setError(null);
                }, 5000); // Adjust the duration as needed (5000ms = 5 seconds)

                return () => clearTimeout(timer);
            }
        }, [error]);

        const getErrorMessage = (error: string | object): string => {
            if (typeof error === 'string') {
                return error;
            } else if (error && typeof error === 'object' && 'message' in error) {
                return (error as { message: string }).message;
            }
            return 'An unknown error occurred';
        };

        return (
            <ErrorContext.Provider value={{ error, setError }}>
                {children}
                {error && (
                    <Alert
                        severity="error"
                        color="error"
                        style={{
                            position: 'absolute',
                            bottom: '5%',
                            left: '50%',
                            transform: 'translateX(-50%)',
                            width: '80%',
                            maxWidth: '600px',
                            textAlign: 'center',
                        }}>
                        {getErrorMessage(error)}
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