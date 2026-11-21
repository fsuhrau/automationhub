import {expect, it} from 'vitest';
import {IUser} from '../../types/user';
import {accountDetails, keychainCommand, mcpCommand, mcpUrl, proxyCommand} from './profileModel';

const user = {name: 'Ann', email: 'ann@x.io', login: 'ann', avatar: ''} as IUser;

it('commands: builds the MCP url with an encoded project', () => {
    expect(mcpUrl('https://hub', 'a b')).toBe('https://hub/api/a%20b/mcp');
});

it('commands: builds the three commands of the guide', () => {
    expect(mcpCommand('https://hub', 'p1', 'tok')).toBe(
        'claude mcp add --transport http hub \\\n  https://hub/api/p1/mcp \\\n  --header "X-Auth-Token: tok"');
    expect(keychainCommand('hub:80')).toBe('security add-generic-password -U -s automationhub -a hub:80 -w');
    expect(proxyCommand('https://hub', 'p1')).toBe('automationhub mcp-proxy --upstream https://hub --project p1');
});

it('accountDetails: uses the session user without a profile; a login equal to the name says nothing', () => {
    expect(accountDetails(null, user)).toEqual({name: 'Ann', email: 'ann@x.io', login: 'ann'});
    expect(accountDetails(null, {...user, login: 'Ann'})).toMatchObject({login: ''});
});

it('accountDetails: prefers the profile', () => {
    const profile = {id: 1, name: 'Anna', login: 'a', email: 'a@x.io'};
    expect(accountDetails(profile, user)).toEqual({name: 'Anna', email: 'a@x.io', login: 'a'});
});

it('accountDetails: falls back to the user for empty profile fields and a missing name', () => {
    const profile = {id: 1, name: 'Anna', login: '', email: ''};
    expect(accountDetails(profile, user)).toMatchObject({email: 'ann@x.io', login: 'ann'});
    expect(accountDetails(null, {...user, name: undefined as unknown as string}).name).toBe('');
});
