// Settings without React: the new app form, the app parameter form and its
// checks (the same rules the master has), token display, and a node's config.

import {AppParameter, IAppData} from '../../types/app';
import {getPlatformName, PlatformType} from '../../types/platform.type.enum';
import {ParameterRequest} from '../../types/settings.requests';

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
export const parameterRequest = (f: ParameterForm): ParameterRequest => ({
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

// platforms that install a bundle need its identifier
export const needsIdentifier = (p: PlatformType): boolean =>
    p === PlatformType.Android || p === PlatformType.iOS || p === PlatformType.Mac;

export const identifierLabel = (p: PlatformType): string => p === PlatformType.Android ? 'Package name' : 'Bundle identifier';

// the identifier is asked for where the hub starts and stops the app by it
export const asksIdentifier = (p: PlatformType): boolean =>
    needsIdentifier(p) || p === PlatformType.Windows || p === PlatformType.Linux;

export const platformLabel = (p: PlatformType): string => p === PlatformType.Editor ? 'Unity editor' : getPlatformName(p);

export interface NewAppErrors {
    name?: string,
    identifier?: string,
}

export const newAppErrors = (name: string, platform: PlatformType, identifier: string): NewAppErrors => {
    const errors: NewAppErrors = {};
    if (name.trim() === '') errors.name = 'A name is needed.';
    if (needsIdentifier(platform) && identifier.trim() === '') {
        errors.identifier = `The ${identifierLabel(platform).toLowerCase()} is needed to install bundles.`;
    }
    return errors;
};

// sortApps lists apps by platform, then by name
export const sortApps = (apps: IAppData[]): IAppData[] =>
    [...apps].sort((a, b) => a.platform - b.platform || a.name.localeCompare(b.name));

// parameterChoices are the options of a choice parameter, none for text
export const parameterChoices = (p: AppParameter): string[] =>
    p.type.type === 'option' ? (p.type as { options?: string[] }).options ?? [] : [];
