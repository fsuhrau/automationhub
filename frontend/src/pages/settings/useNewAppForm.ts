import {useState} from 'react';
import {PlatformType} from '../../types/platform.type.enum';
import {NewAppRequest} from '../../types/settings.requests';
import {NewAppErrors, newAppErrors} from './settingsModel';

export interface NewAppForm {
    name: string,
    setName: (name: string) => void,
    platform: PlatformType,
    setPlatform: (platform: PlatformType) => void,
    identifier: string,
    setIdentifier: (identifier: string) => void,
    // the problems, shown once a submit was tried
    errors: NewAppErrors,
    // marks the submit as tried; true when there is no problem
    validate: () => boolean,
    request: () => NewAppRequest,
}

// useNewAppForm is the state of the "Add an app" form
export function useNewAppForm(): NewAppForm {
    const [name, setName] = useState('');
    const [platform, setPlatform] = useState<PlatformType>(PlatformType.Android);
    const [identifier, setIdentifier] = useState('');
    const [tried, setTried] = useState(false);
    const found = newAppErrors(name, platform, identifier);
    return {
        name, setName, platform, setPlatform, identifier, setIdentifier,
        errors: tried ? found : {},
        validate: () => {
            setTried(true);
            return Object.keys(found).length === 0;
        },
        request: () => ({name: name.trim(), platform, identifier: identifier.trim()}),
    };
}
