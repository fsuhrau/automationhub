import {expect, it} from 'vitest';
import {AppParameter, IAppData} from '../../types/app';
import {PlatformType} from '../../types/platform.type.enum';
import {
    asksIdentifier, ParameterForm, identifierLabel, isExpired, maskToken, needsIdentifier, newAppErrors, nodeConfig, parameterChoices,
    parameterErrors, parameterForm, parameterRequest, platformLabel, sortApps, tidyOptions,
} from './settingsModel';

const choice: AppParameter = {id: 2, name: 'Market', type: {type: 'option', options: ['eu', 'us'], defaultValue: 'eu'}};
const text: AppParameter = {id: 1, name: 'Level', type: {type: 'string', defaultValue: '3'}};

it('parameterForm: is empty text for a new parameter', () => {
    expect(parameterForm(null)).toEqual({name: '', kind: 'string', defaultValue: '', options: []});
});

it('parameterForm: copies a choice parameter', () => {
    expect(parameterForm(choice)).toEqual({name: 'Market', kind: 'option', defaultValue: 'eu', options: ['eu', 'us']});
});

it('parameterForm: copies a text parameter without options and tolerates a missing default', () => {
    expect(parameterForm({id: 3, name: 'X', type: {type: 'string'}})).toEqual({name: 'X', kind: 'string', defaultValue: '', options: []});
    expect(parameterForm(text).kind).toBe('string');
});

it('tidyOptions: trims and drops blanks and doubles', () => {
    expect(tidyOptions([' a ', '', 'a', 'b'])).toEqual(['a', 'b']);
});

const form = (change: object): ParameterForm => ({...parameterForm(null), name: 'New', ...change});

it.each([
    ['a missing name', form({name: ' '}), {name: 'A name is needed.'}],
    ['a name of a sibling, ignoring case', form({name: 'level'}), {name: 'The app has a parameter with this name.'}],
    ['a choice without options', form({kind: 'option'}), {options: 'A choice needs at least one option.'}],
    ['a default outside the options', form({kind: 'option', options: ['a'], defaultValue: 'b'}), {defaultValue: 'Pick the default from the options.'}],
])('parameterErrors: reports %s', (_, f, expected) => {
    expect(parameterErrors(f, [text], null)).toEqual(expected);
});

it('parameterErrors: accepts a valid choice, and the own name when editing', () => {
    expect(parameterErrors(form({kind: 'option', options: ['a'], defaultValue: 'a'}), [text], null)).toEqual({});
    expect(parameterErrors(form({name: 'Level'}), [text], 1)).toEqual({});
});

it('parameterRequest: is the body of a choice with tidy options', () => {
    const f = {name: ' Market ', kind: 'option' as const, defaultValue: 'eu', options: ['eu', ' eu', 'us']};
    expect(parameterRequest(f)).toEqual({name: 'Market', type: {type: 'option', defaultValue: 'eu', options: ['eu', 'us']}});
});

it('parameterRequest: is the body of a text', () => {
    expect(parameterRequest({name: 'L', kind: 'string', defaultValue: '3', options: []})).toEqual({name: 'L', type: {type: 'string', defaultValue: '3'}});
});

it('token and node helpers: masks all but the end of a long token', () => {
    expect(maskToken('abcdefghijkl')).toBe('••••ghijkl');
    expect(maskToken('short')).toBe('••••');
});

it('token and node helpers: writes the node config', () => {
    expect(nodeConfig('n1', 'tok', 'hub:80')).toBe('master_url: hub:80\nidentifier: n1\nauth:\n  token:\n    auth_token: tok');
});

it.each([
    [null, false], [undefined, false], ['2020-01-01', true], ['2999-01-01', false],
])('token helpers: isExpired(%s) is %s', (value, expired) => {
    expect(isExpired(value, Date.parse('2024-01-01'))).toBe(expired);
});

it('app form helpers: knows which platforms need an identifier', () => {
    expect(needsIdentifier(PlatformType.Android)).toBe(true);
    expect(needsIdentifier(PlatformType.Web)).toBe(false);
    expect(asksIdentifier(PlatformType.Linux)).toBe(true);
    expect(asksIdentifier(PlatformType.Editor)).toBe(false);
});

it('app form helpers: labels the identifier and platform', () => {
    expect(identifierLabel(PlatformType.Android)).toBe('Package name');
    expect(identifierLabel(PlatformType.iOS)).toBe('Bundle identifier');
    expect(platformLabel(PlatformType.Editor)).toBe('Unity editor');
    expect(platformLabel(PlatformType.Mac)).toBe('Mac');
});

it('app form helpers: checks a new app', () => {
    expect(newAppErrors('', PlatformType.Android, '')).toEqual({
        name: 'A name is needed.', identifier: 'The package name is needed to install bundles.',
    });
    expect(newAppErrors('A', PlatformType.Web, '')).toEqual({});
});

it('app form helpers: sorts apps by platform, then name', () => {
    const app = (name: string, platform: PlatformType): IAppData => ({name, platform}) as IAppData;
    const sorted = sortApps([app('b', PlatformType.Android), app('z', PlatformType.iOS), app('a', PlatformType.Android)]);
    expect(sorted.map(a => a.name)).toEqual(['z', 'a', 'b']);
});

it('app form helpers: lists the choices of a choice parameter only', () => {
    expect(parameterChoices(choice)).toEqual(['eu', 'us']);
    expect(parameterChoices(text)).toEqual([]);
    expect(parameterChoices({id: 4, name: 'Y', type: {type: 'option', defaultValue: ''} as AppParameter['type']})).toEqual([]);
});
