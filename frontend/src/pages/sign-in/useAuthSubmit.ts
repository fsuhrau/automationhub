import {FormEvent, useState} from 'react';

type OnError = (message: string) => void;

export interface AuthSubmit {
    // a field error shows only after the first try
    tried: boolean,
    error: string | null,
    busy: boolean,
    submit: (e: FormEvent) => void,
}

// the state of a sign in or sign up form: sends only valid input, and keeps
// the busy flag and the error of the master's answer
export const useAuthSubmit = (valid: boolean, send: (onError: OnError) => void): AuthSubmit => {
    const [tried, setTried] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [busy, setBusy] = useState(false);

    const submit = (e: FormEvent): void => {
        e.preventDefault();
        setTried(true);
        setError(null);
        if (!valid) return;
        setBusy(true);
        send(message => {
            setError(message);
            setBusy(false);
        });
    };
    return {tried, error, busy, submit};
};
