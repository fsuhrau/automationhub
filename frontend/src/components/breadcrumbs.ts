// Breadcrumbs are derived from the URL, so every page shows where it is
// without each page having to say so. The links are built with lib/routes.ts,
// so they stay in step with the route table.
import {AppRoutes, ProjectRoutes, routes, TestRoutes} from '../lib/routes';

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

// what a page of a project knows: its routes, the path after the project and the names
interface Scope {
    project: ProjectRoutes,
    rest: string[],
    names: CrumbNames,
}

const crumb = (label: string, to: string): Crumb => ({label, to});

const PROJECTS: Crumb = crumb('Projects', routes.projects());

const deviceCrumbs = ({project, rest, names}: Scope): Crumb[] => {
    const [, id, action] = rest;
    if (id === undefined) return [];
    const device = project.device(id);
    const crumbs = [crumb('Device Pool', project.devices()), crumb(names.device(+id) ?? `Device ${id}`, device.show())];
    if (action === 'control') crumbs.push(crumb('Live control', device.control()));
    if (action === 'edit') crumbs.push(crumb('Edit', device.edit()));
    return crumbs;
};

const runCrumbs = (test: TestRoutes, [, , , , runId, protocolId]: string[]): Crumb[] => {
    if (runId === undefined) return [];
    const run = test.run(runId);
    const crumbs = [crumb(`Run ${runId}`, run.show())];
    if (protocolId !== undefined) crumbs.push(crumb(`Protocol ${protocolId}`, run.protocol(protocolId)));
    return crumbs;
};

// what follows the test's name: its edit page, the last run or one run
const testPageCrumbs = (test: TestRoutes, rest: string[]): Crumb[] => {
    const [, , , page, sub] = rest;
    if (page === 'edit') return [crumb('Edit', test.edit())];
    if (page === 'runs' && sub === 'last') return [crumb('Last run', test.lastRun())];
    return page === 'run' ? runCrumbs(test, rest) : [];
};

const testCrumbs = (app: AppRoutes, appId: number, {rest, names}: Scope): Crumb[] => {
    const id = rest[2];
    if (id === 'new') return [crumb('New test', app.newTest())];
    const test = app.test(id);
    return [crumb(names.test(appId, +id) ?? `Test ${id}`, test.show()), ...testPageCrumbs(test, rest)];
};

const bundleCrumbs = (app: AppRoutes, rest: string[]): Crumb[] => {
    const [, section, id] = rest;
    const bundles = crumb('Bundles', app.bundles());
    if (section === 'bundles') return id === 'compare' ? [bundles, crumb('Compare', app.compareBundles())] : [bundles];
    return [bundles, crumb(`Bundle #${id}`, app.bundle(id))];
};

// what follows the app's name: tests or bundles
const sectionCrumbs = (app: AppRoutes, appId: number, scope: Scope): Crumb[] => {
    const [, section, id] = scope.rest;
    if (section === 'bundles' || (section === 'bundle' && id)) return bundleCrumbs(app, scope.rest);
    if (section === 'tests') return [crumb('Tests', app.tests())];
    if (section !== 'test') return [];
    const tests = crumb('Tests', app.tests());
    return id ? [tests, ...testCrumbs(app, appId, scope)] : [tests];
};

const appCrumbs = (scope: Scope): Crumb[] => {
    const appId = +scope.rest[0].substring('app:'.length);
    const app = scope.project.app(appId);
    return [crumb(scope.names.app(appId) ?? `App ${appId}`, app.tests()), ...sectionCrumbs(app, appId, scope)];
};

// the crumbs after the project's name, by the first part of the path
const inProject = (scope: Scope): Crumb[] => {
    const {project, rest} = scope;
    if (rest.length === 0) return [crumb('Overview', project.overview())];
    if (rest[0] === 'settings') return [crumb('Settings', project.settings())];
    if (rest[0] === 'devices') return [crumb('Device Pool', project.devices())];
    if (rest[0] === 'device') return deviceCrumbs(scope);
    return rest[0].startsWith('app:') ? appCrumbs(scope) : [];
};

const projectCrumbs = (identifier: string, rest: string[], names: CrumbNames): Crumb[] => {
    const project = routes.project(identifier);
    return [PROJECTS, crumb(names.project(identifier) ?? identifier, project.overview()), ...inProject({project, rest, names})];
};

// the last crumb is the current page and no link
const withCurrentPage = (crumbs: Crumb[]): Crumb[] => [...crumbs.slice(0, -1), {label: crumbs[crumbs.length - 1].label}];

export const buildCrumbs = (pathname: string, names: CrumbNames): Crumb[] => {
    const parts = pathname.split('/').filter(p => p.length > 0).map(decodeURIComponent);
    if (parts.length === 0) return [{label: 'Projects'}];
    if (parts[0] === 'profile') return [PROJECTS, {label: 'Profile'}];
    if (parts[0] !== 'project' || parts.length < 2) return [PROJECTS];
    return withCurrentPage(projectCrumbs(parts[1], parts.slice(2), names));
};

export interface TestPath {
    project: string,
    appId: number,
    testId: string,
}

// the test a path points into (its pages and its runs), if any
export const parseTestPath = (pathname: string): TestPath | null => {
    const match = pathname.match(/^\/project\/([^/]+)\/app:(\d+)\/test\/(\d+)/);
    return match ? {project: decodeURIComponent(match[1]), appId: +match[2], testId: match[3]} : null;
};
