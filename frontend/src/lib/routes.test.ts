import {readFileSync} from 'node:fs';
import {describe, expect, it} from 'vitest';
import {buildCrumbs, CrumbNames} from '../components/breadcrumbs';
import {routes} from './routes';

const p = routes.project('my proj');
const base = '/project/my%20proj';
const a = p.app(3);
const t = a.test(5);

// every path pattern of the route table, with the builder that makes it
const BUILDERS: Record<string, string> = {
    '/login': routes.login(),
    '/signup': routes.signup(),
    '/': routes.projects(),
    'profile': routes.profile(),
    ':project_identifier': p.overview(),
    // the layout route of the app pages, it has no page of its own
    ':project_identifier/:appId': a.tests(),
    ':project_identifier/:appId/home': a.home(),
    'bundles': a.bundles(),
    'bundles/compare': a.compareBundles(),
    'bundle/:bundleId': a.bundle(7),
    'tests': a.tests(),
    'test/:testId': t.show(),
    'test/new': a.newTest(),
    'test/:testId/edit': t.edit(),
    'test/:testId/run/:runId/replay': t.run(9).replay(),
    'test/:testId/run/:runId/:protocolId': t.run(9).protocol(2),
    'test/:testId/run/:runId': t.run(9).show(),
    'test/:testId/runs/last': t.lastRun(),
    'results': a.results(),
    'performance': a.performance(),
    ':project_identifier/settings': p.settings(),
    ':project_identifier/users': p.users(),
    ':project_identifier/device/:deviceId/control': p.device(4).control(),
    ':project_identifier/device/:deviceId/edit': p.device(4).edit(),
    ':project_identifier/device/:deviceId': p.device(4).show(),
    ':project_identifier/devices': p.devices(),
};

const appPatterns = (): string[] => {
    const source = readFileSync(`${__dirname}/../app/routes.tsx`, 'utf8');
    return [...source.matchAll(/<Route\s+path=\{?["']([^"']+)["']/g)].map(m => m[1]);
};

describe('routes of App.tsx', () => {
    it('has a builder for every path pattern of the route table', () => {
        const patterns = appPatterns().filter(x => !['project'].includes(x));
        expect(patterns.length).toBeGreaterThan(20);
        expect(patterns.filter(x => !(x in BUILDERS))).toEqual([]);
    });

});

const EXPECTED: [string, string, string][] = [
    ['login', routes.login(), '/login'],
    ['signup', routes.signup(), '/signup'],
    ['projects', routes.projects(), '/'],
    ['profile', routes.profile(), '/profile'],
    ['overview', p.overview(), base],
    ['settings', p.settings(), `${base}/settings`],
    ['settings tab', p.settings('apps'), `${base}/settings?tab=apps`],
    ['users', p.users(), `${base}/users`],
    ['devices', p.devices(), `${base}/devices`],
    ['device', p.device(4).show(), `${base}/device/4`],
    ['device control', p.device(4).control(), `${base}/device/4/control`],
    ['device edit', p.device(4).edit(), `${base}/device/4/edit`],
    ['app home', a.home(), `${base}/app:3/home`],
    ['tests', a.tests(), `${base}/app:3/tests`],
    ['new test', a.newTest(), `${base}/app:3/test/new`],
    ['test', t.show(), `${base}/app:3/test/5`],
    ['test edit', t.edit(), `${base}/app:3/test/5/edit`],
    ['last run', t.lastRun(), `${base}/app:3/test/5/runs/last`],
    ['run', t.run(9).show(), `${base}/app:3/test/5/run/9`],
    ['replay', t.run(9).replay(), `${base}/app:3/test/5/run/9/replay`],
    ['protocol', t.run(9).protocol(2), `${base}/app:3/test/5/run/9/2`],
    ['bundles', a.bundles(), `${base}/app:3/bundles`],
    ['compare', a.compareBundles(), `${base}/app:3/bundles/compare`],
    ['bundle', a.bundle(7), `${base}/app:3/bundle/7`],
    ['results', a.results(), `${base}/app:3/results`],
    ['performance', a.performance(), `${base}/app:3/performance`],
];

describe('route builders', () => {
    it.each(EXPECTED)('build %s', (_name, actual, expected) => {
        expect(actual).toBe(expected);
    });
});

describe('breadcrumbs parse the built paths', () => {
    const names: CrumbNames = {project: () => undefined, app: () => undefined, device: () => undefined, test: () => undefined};
    const labels = (path: string): string[] => buildCrumbs(path, names).map(c => c.label);

    it('for the device pages', () => {
        expect(labels(p.device(4).control())).toEqual(['Projects', 'my proj', 'Device Pool', 'Device 4', 'Live control']);
        expect(labels(p.device(4).edit()).at(-1)).toBe('Edit');
    });

    it('for the test pages', () => {
        expect(labels(t.run(9).protocol(2)).slice(-3)).toEqual(['Test 5', 'Run 9', 'Protocol 2']);
        expect(labels(t.lastRun()).at(-1)).toBe('Last run');
        expect(labels(a.newTest()).at(-1)).toBe('New test');
    });

    it('for the bundle pages', () => {
        expect(labels(a.compareBundles()).slice(-2)).toEqual(['Bundles', 'Compare']);
        expect(labels(a.bundle(7)).at(-1)).toBe('Bundle #7');
    });
});
