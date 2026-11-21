import {act, RenderResult, screen} from '@testing-library/react';
import {Route, Routes} from 'react-router-dom';
import {describe, expect, it} from 'vitest';
import NavbarBreadcrumbs from './NavbarBreadcrumbs';
import {fakeApi, FakeApi} from '../testing/fakeApi';
import {aDevice, aProject, anApp} from '../testing/fixtures';
import {renderWithProviders} from '../testing/renderWithProviders';

const hub = {
    projects: [aProject('p1', 'Project One')],
    apps: [anApp(3, 'Android app')],
    devices: [aDevice({id: 4, name: 'Pixel 7'}), aDevice({id: 5, name: 'raw name', alias: 'Lab phone'})],
};

const mount = async (path: string, api: FakeApi = fakeApi()): Promise<RenderResult> => {
    const view = renderWithProviders(<Routes><Route path="*" element={<NavbarBreadcrumbs/>}/></Routes>, {url: path, api, hub});
    await act(async () => undefined);
    return view;
};

const trail = (): string => screen.getByRole('navigation', {name: 'breadcrumb'}).textContent ?? '';

describe('NavbarBreadcrumbs', () => {
    it('names the project, the app and the device from the hub state', async () => {
        await mount('/project/p1/device/4');
        expect(trail()).toContain('Project One');
        expect(trail()).toContain('Pixel 7');
    });

    it('prefers the alias of a device', async () => {
        await mount('/project/p1/device/5/control');
        expect(trail()).toContain('Lab phone');
        expect(trail()).not.toContain('raw name');
    });

    it('falls back to numbers for devices and apps that are not known', async () => {
        await mount('/project/p1/device/99');
        expect(trail()).toContain('Device 99');
    });

    it('looks up the name of a test once', async () => {
        const api = fakeApi({'GET /p1/app/3/test/11': {name: 'Smoke test'}});
        await mount('/project/p1/app:3/test/11', api);
        expect(trail()).toContain('Smoke test');
        expect(trail()).toContain('Android app');
        await mount('/project/p1/app:3/test/11/run/2', api);
        expect(api.called('GET /p1/app/3/test/11')).toHaveLength(1);
    });

    it('shows the number of a test whose name cannot be loaded', async () => {
        await mount('/project/p1/app:3/test/12', fakeApi({'GET /p1/app/3/test/12': new Error('gone')}));
        expect(trail()).toContain('Test 12');
    });

    it('does not update after it is gone', async () => {
        let resolve: (v: { name: string }) => void = () => undefined;
        const pending = new Promise<{ name: string }>(r => {
            resolve = r;
        });
        const view = await mount('/project/p1/app:3/test/13', fakeApi({'GET /p1/app/3/test/13': () => pending}));
        view.unmount();
        await act(async () => resolve({name: 'late'}));
        expect(screen.queryByRole('navigation')).toBeNull();
    });
});
