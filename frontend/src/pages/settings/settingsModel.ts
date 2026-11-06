// Settings without React: the app parameter form and its checks (the same
// rules the master has), token display, and a node's config.

export type ParameterKind = 'string' | 'option';

export interface ParameterForm {
    name: string,
    kind: ParameterKind,
    defaultValue: string,
    options: string[],
}

export interface ParameterLike {
    id: number,
    name: string,
    type: { type: string, defaultValue?: string, options?: string[] },
}

export const parameterForm = (p: ParameterLike | null): ParameterForm => p
    ? {name: p.name, kind: p.type.type === 'option' ? 'option' : 'string', defaultValue: p.type.defaultValue ?? '', options: [...(p.type.options ?? [])]}
    : {name: '', kind: 'string', defaultValue: '', options: []};

export const tidyOptions = (options: string[]): string[] =>
    Array.from(new Set(options.map(o => o.trim()).filter(o => o.length > 0)));

export interface ParameterErrors {
    name?: string,
    options?: string,
    defaultValue?: string,
}

// parameterErrors checks the form against the app's other parameters
export const parameterErrors = (f: ParameterForm, others: ParameterLike[], selfId: number | null): ParameterErrors => {
    const errors: ParameterErrors = {};
    const name = f.name.trim();
    if (name === '') errors.name = 'A name is needed.';
    else if (others.some(o => o.id !== selfId && o.name.toLowerCase() === name.toLowerCase())) errors.name = 'The app has a parameter with this name.';
    if (f.kind === 'option') {
        const options = tidyOptions(f.options);
        if (options.length === 0) errors.options = 'A choice needs at least one option.';
        else if (f.defaultValue && !options.includes(f.defaultValue)) errors.defaultValue = 'Pick the default from the options.';
    }
    return errors;
};

// parameterRequest is the body the master stores
export const parameterRequest = (f: ParameterForm) => ({
    name: f.name.trim(),
    type: f.kind === 'option'
        ? {type: 'option', defaultValue: f.defaultValue, options: tidyOptions(f.options)}
        : {type: 'string', defaultValue: f.defaultValue},
});

// maskToken shows a token's end only, for recognizing it
export const maskToken = (token: string): string => token.length > 8 ? `••••${token.slice(-6)}` : '••••';

// nodeConfig is what a node's automationhub.yaml needs to connect
export const nodeConfig = (identifier: string, token: string, master: string): string =>
    [`master_url: ${master}`, `identifier: ${identifier}`, 'auth:', '  token:', `    auth_token: ${token}`].join('\n');

export const isExpired = (expiresAt: Date | string | null | undefined, now: number): boolean =>
    !!expiresAt && new Date(expiresAt).valueOf() < now;
