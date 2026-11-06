// Breadcrumbs are derived from the URL, so every page shows where it is
// without each page having to say so.

export interface Crumb {
    label: string,
    // absent for the current page
    to?: string,
}

export interface CrumbNames {
    project: (identifier: string) => string | undefined,
    app: (id: number) => string | undefined,
    device: (id: number) => string | undefined,
    test: (appId: number, id: number) => string | undefined,
}

const PROJECTS: Crumb = {label: 'Projects', to: '/'};

export const buildCrumbs = (pathname: string, names: CrumbNames): Crumb[] => {
    const parts = pathname.split('/').filter(p => p.length > 0).map(decodeURIComponent);
    const crumbs: Crumb[] = [];
    const push = (label: string, to: string) => crumbs.push({label, to});

    if (parts.length === 0) {
        return [{label: 'Projects'}];
    }
    if (parts[0] === 'profile') {
        return [PROJECTS, {label: 'Profile'}];
    }
    if (parts[0] !== 'project' || parts.length < 2) {
        return [PROJECTS];
    }

    crumbs.push(PROJECTS);
    const pid = parts[1];
    const base = `/project/${encodeURIComponent(pid)}`;
    push(names.project(pid) ?? pid, base);
    const rest = parts.slice(2);

    if (rest.length === 0) {
        push('Overview', base);
    } else if (rest[0] === 'settings') {
        push('Settings', `${base}/settings`);
    } else if (rest[0] === 'devices') {
        push('Device Pool', `${base}/devices`);
    } else if (rest[0] === 'device' && rest.length >= 2) {
        const id = +rest[1];
        push('Device Pool', `${base}/devices`);
        push(names.device(id) ?? `Device ${rest[1]}`, `${base}/device/${rest[1]}`);
        if (rest[2] === 'control') push('Live control', `${base}/device/${rest[1]}/control`);
        if (rest[2] === 'edit') push('Edit', `${base}/device/${rest[1]}/edit`);
    } else if (rest[0].startsWith('app:')) {
        const appId = +rest[0].substring(4);
        const appBase = `${base}/app:${appId}`;
        push(names.app(appId) ?? `App ${appId}`, `${appBase}/tests`);
        const section = rest[1];
        if (section === 'bundles') {
            push('Bundles', `${appBase}/bundles`);
        } else if (section === 'tests' || section === 'test') {
            push('Tests', `${appBase}/tests`);
        }
        if (section === 'test' && rest.length >= 3) {
            if (rest[2] === 'new') {
                push('New test', `${appBase}/test/new`);
            } else {
                const testId = +rest[2];
                const testBase = `${appBase}/test/${rest[2]}`;
                push(names.test(appId, testId) ?? `Test ${rest[2]}`, testBase);
                if (rest[3] === 'edit') {
                    push('Edit', `${testBase}/edit`);
                } else if (rest[3] === 'runs' && rest[4] === 'last') {
                    push('Last run', `${testBase}/runs/last`);
                } else if (rest[3] === 'run' && rest[4]) {
                    push(`Run ${rest[4]}`, `${testBase}/run/${rest[4]}`);
                    if (rest[5]) push(`Protocol ${rest[5]}`, `${testBase}/run/${rest[4]}/${rest[5]}`);
                }
            }
        }
    }

    // the last crumb is the current page and no link
    const last = crumbs[crumbs.length - 1];
    crumbs[crumbs.length - 1] = {label: last.label};
    return crumbs;
};
