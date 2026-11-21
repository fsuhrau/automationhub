import {useState} from 'react';

export interface NameField {
    name: string,
    setName: (name: string) => void,
    // shown only after a submit was tried
    invalid: boolean,
    // marks the submit as tried; true when the name is filled in
    validate: () => boolean,
    trimmed: string,
    reset: () => void,
}

// useNameField is the state of a required "Name" input of a small form
export function useNameField(): NameField {
    const [name, setName] = useState('');
    const [tried, setTried] = useState(false);
    const trimmed = name.trim();
    return {
        name, setName, trimmed,
        invalid: tried && trimmed === '',
        validate: () => {
            setTried(true);
            return trimmed !== '';
        },
        reset: () => {
            setName('');
            setTried(false);
        },
    };
}
