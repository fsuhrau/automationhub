import {useState} from 'react';
import {AppParameter} from '../../types/app';
import {ParameterErrors, ParameterForm, parameterErrors, parameterForm, tidyOptions} from './settingsModel';

export interface ParameterFormState {
    form: ParameterForm,
    set: (change: Partial<ParameterForm>) => void,
    // what is typed into the option input, not yet added
    option: string,
    setOption: (option: string) => void,
    // adds the typed option; commas separate several
    addOption: () => void,
    // the options without blanks and doubles
    options: string[],
    // the problems, shown once a submit was tried
    errors: ParameterErrors,
    // marks the submit as tried; true when there is no problem
    validate: () => boolean,
}

// useParameterForm is the state of the form that adds or edits a parameter
export function useParameterForm(parameter: AppParameter | null, siblings: AppParameter[]): ParameterFormState {
    const [form, setForm] = useState<ParameterForm>(() => parameterForm(parameter));
    const [option, setOption] = useState('');
    const [tried, setTried] = useState(false);
    const found = parameterErrors(form, siblings, parameter?.id ?? null);
    const set = (change: Partial<ParameterForm>): void => setForm(f => ({...f, ...change}));
    return {
        form, set, option, setOption,
        addOption: () => {
            set({options: tidyOptions([...form.options, ...option.split(',')])});
            setOption('');
        },
        options: tidyOptions(form.options),
        errors: tried ? found : {},
        validate: () => {
            setTried(true);
            return Object.keys(found).length === 0;
        },
    };
}
