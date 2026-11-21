import {describe, expect, it} from 'vitest';
import {buildCrumbs, CrumbNames, parseTestPath} from './breadcrumbs';

const none: CrumbNames = {project: () => undefined, app: () => undefined, device: () => undefined, test: () => undefined};
const named: CrumbNames = {
    project: () => 'Project One',
    app: () => 'Android',
    device: () => 'Pixel',
    test: () => 'Smoke',
};

const labels = (path: string, names = none): string[] => buildCrumbs(path, names).map(c => c.label);
const links = (path: string, names = none): (string | undefined)[] => buildCrumbs(path, names).map(c => c.to);

describe('buildCrumbs outside projects', () => {
    it.each([
        ['/', ['Projects']],
        ['/profile', ['Projects', 'Profile']],
        ['/elsewhere', ['Projects']],
        ['/project', ['Projects']],
    ])('for %s', (path, expected) => {
        expect(labels(path)).toEqual(expected);
    });

    it('links Projects unless it is the current page', () => {
        expect(links('/')).toEqual([undefined]);
        expect(links('/profile')).toEqual(['/', undefined]);
    });
});

describe('buildCrumbs in a project', () => {
    it.each([
        ['/project/p', ['Projects', 'p', 'Overview']],
        ['/project/p/settings', ['Projects', 'p', 'Settings']],
        ['/project/p/devices', ['Projects', 'p', 'Device Pool']],
        ['/project/p/users', ['Projects', 'p']],
        ['/project/p/device', ['Projects', 'p']],
        ['/project/p/device/4', ['Projects', 'p', 'Device Pool', 'Device 4']],
        ['/project/p/device/4/control', ['Projects', 'p', 'Device Pool', 'Device 4', 'Live control']],
        ['/project/p/device/4/edit', ['Projects', 'p', 'Device Pool', 'Device 4', 'Edit']],
        ['/project/p/app:3', ['Projects', 'p', 'App 3']],
        ['/project/p/app:3/results', ['Projects', 'p', 'App 3']],
        ['/project/p/app:3/tests', ['Projects', 'p', 'App 3', 'Tests']],
        ['/project/p/app:3/test', ['Projects', 'p', 'App 3', 'Tests']],
        ['/project/p/app:3/test/new', ['Projects', 'p', 'App 3', 'Tests', 'New test']],
        ['/project/p/app:3/test/5', ['Projects', 'p', 'App 3', 'Tests', 'Test 5']],
        ['/project/p/app:3/test/5/edit', ['Projects', 'p', 'App 3', 'Tests', 'Test 5', 'Edit']],
    ])('for %s', (path, expected) => {
        expect(labels(path)).toEqual(expected);
    });
});

describe('buildCrumbs in an app', () => {
    it.each([
        ['/project/p/app:3/test/5/runs/last', ['Projects', 'p', 'App 3', 'Tests', 'Test 5', 'Last run']],
        ['/project/p/app:3/test/5/runs', ['Projects', 'p', 'App 3', 'Tests', 'Test 5']],
        ['/project/p/app:3/test/5/run', ['Projects', 'p', 'App 3', 'Tests', 'Test 5']],
        ['/project/p/app:3/test/5/run/9', ['Projects', 'p', 'App 3', 'Tests', 'Test 5', 'Run 9']],
        ['/project/p/app:3/test/5/run/9/2', ['Projects', 'p', 'App 3', 'Tests', 'Test 5', 'Run 9', 'Protocol 2']],
        ['/project/p/app:3/bundles', ['Projects', 'p', 'App 3', 'Bundles']],
        ['/project/p/app:3/bundles/compare', ['Projects', 'p', 'App 3', 'Bundles', 'Compare']],
        ['/project/p/app:3/bundle', ['Projects', 'p', 'App 3']],
        ['/project/p/app:3/bundle/7', ['Projects', 'p', 'App 3', 'Bundles', 'Bundle #7']],
    ])('for %s', (path, expected) => {
        expect(labels(path)).toEqual(expected);
    });

    it('uses the names it is given', () => {
        expect(labels('/project/p/app:3/test/5', named)).toEqual(['Projects', 'Project One', 'Android', 'Tests', 'Smoke']);
        expect(labels('/project/p/device/4', named)).toEqual(['Projects', 'Project One', 'Device Pool', 'Pixel']);
    });

    it('links every crumb but the current page, with the routes of the app', () => {
        expect(links('/project/my%20proj/app:3/test/5/run/9')).toEqual([
            '/', '/project/my%20proj', '/project/my%20proj/app:3/tests', '/project/my%20proj/app:3/tests',
            '/project/my%20proj/app:3/test/5', undefined,
        ]);
        expect(labels('/project/my%20proj/settings')[1]).toBe('my proj');
    });
});

describe('parseTestPath', () => {
    it('finds the test of its pages and runs', () => {
        expect(parseTestPath('/project/my%20proj/app:3/test/5/run/9')).toEqual({project: 'my proj', appId: 3, testId: '5'});
    });

    it.each(['/project/p/app:3/tests', '/project/p', '/'])('finds none in %s', path => {
        expect(parseTestPath(path)).toBeNull();
    });
});
