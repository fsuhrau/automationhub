// The profile page without React: the commands of the Claude Code guide and
// what the account section shows.
import {IProfile} from '../../types/profile.token';
import {IUser} from '../../types/user';

// where the browser reaches the hub (window.location)
export interface HubLocation {
    origin: string,
    host: string,
}

export const TOKEN_PLACEHOLDER = '<your personal token>';

// The MCP endpoint of a project, as the master serves it.
export const mcpUrl = (origin: string, project: string): string => `${origin}/api/${encodeURIComponent(project)}/mcp`;

export const mcpCommand = (origin: string, project: string, token: string): string =>
    `claude mcp add --transport http hub \\\n  ${mcpUrl(origin, project)} \\\n  --header "X-Auth-Token: ${token}"`;

// Dev containers must not hold the token: it goes into the host's Keychain
// and "automationhub mcp-proxy" adds it on the host (account = master host).
export const keychainCommand = (host: string): string =>
    `security add-generic-password -U -s automationhub -a ${host} -w`;

export const proxyCommand = (origin: string, project: string): string =>
    `automationhub mcp-proxy --upstream ${origin} --project ${project}`;

export interface AccountDetails {
    name: string,
    email: string,
    // the login, only when it says something beyond the name
    login: string,
}

// accountDetails prefers what the profile says over the session's user
export const accountDetails = (profile: IProfile | null, user: IUser): AccountDetails => {
    const name = profile?.name ?? user.name ?? '';
    const login = profile?.login || user.login;
    return {name, email: profile?.email || user.email, login: login && login !== name ? login : ''};
};
