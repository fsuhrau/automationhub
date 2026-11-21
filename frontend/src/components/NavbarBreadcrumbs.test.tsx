import {act, screen} from '@testing-library/react';
import {Route, Routes} from 'react-router-dom';
import {beforeEach, describe, expect, it, vi} from 'vitest';
import NavbarBreadcrumbs from './NavbarBreadcrumbs';
import {renderInApp} from '../app/testing/renderInApp';
import {aDevice, aProject, anApp, SeedHub} from '../app/testing/seedHub';

const getTest = vi.hoisted(() => vi.fn());
vi.mock('../services/test.service', () => ({getTest}));

const mount = async (path: string): Promise<ReturnType<typeof renderInApp>> => {
    const view = renderInApp(
        <SeedHub projects={[aProject('p1', 'Project One')]} apps={[anApp(3, 'Android app')]}
            devices={[aDevice({id: 4, name: 'Pixel 7'}), aDevice({id: 5, name: 'raw name', alias: 'Lab phone'})]}>
            <Routes><Route path="*" element={<NavbarBreadcrumbs/>}/></Routes>
        </SeedHub>, {path});
    await act(async () => undefined);
    return view;
};

const trail = (): string => screen.getByRole('navigation', {name: 'breadcrumb'}).textContent ?? '';

beforeEach(() => {
    getTest.mockReset();
});

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
        getTest.mockResolvedValue({name: 'Smoke test'});
        await mount('/project/p1/app:3/test/11');
        expect(trail()).toContain('Smoke test');
        expect(trail()).toContain('Android app');
        await mount('/project/p1/app:3/test/11/run/2');
        expect(getTest).toHaveBeenCalledTimes(1);
    });

    it('shows the number of a test whose name cannot be loaded', async () => {
        getTest.mockRejectedValue(new Error('gone'));
        await mount('/project/p1/app:3/test/12');
        expect(trail()).toContain('Test 12');
    });

    it('does not update after it is gone', async () => {
        let resolve: (v: { name: string }) => void = () => undefined;
        getTest.mockReturnValue(new Promise(r => {
            resolve = r;
        }));
        const view = await mount('/project/p1/app:3/test/13');
        view.unmount();
        await act(async () => resolve({name: 'late'}));
        expect(screen.queryByRole('navigation')).toBeNull();
    });
});
