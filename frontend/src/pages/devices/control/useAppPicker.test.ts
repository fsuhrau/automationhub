import {describe, expect, it, vi} from 'vitest';
import {act, renderHook, RenderHookResult, waitFor} from '@testing-library/react';
import {DeviceControlApi} from '../../../types/device.control.api';
import {PlatformType} from '../../../types/platform.type.enum';
import {APPS, fakeService} from './fakes';
import {ErrorHandler, Runner} from './useGuardedAction';
import {AppPicker, AppPickerInput, useAppPicker} from './useAppPicker';

const run: Runner = async action => {
    await action();
};

interface Setup extends RenderHookResult<AppPicker, unknown> {
    service: DeviceControlApi,
    report: ErrorHandler,
}

const setup = (over: Partial<AppPickerInput> = {}): Setup => {
    const service = fakeService();
    const report = vi.fn();
    const input: AppPickerInput = {service, apps: APPS, platform: PlatformType.Android, menuAppId: null, lease: 'tok', ...over};
    const hook = renderHook(() => useAppPicker(input, run, report));
    return {service, report, ...hook};
};

describe('useAppPicker', () => {
    it('offers the first app that runs on the device and its newest bundle', async () => {
        const {result, service} = setup();
        expect(result.current.appId).toBe(1);
        await waitFor(() => expect(result.current.bundleId).toBe(12));
        expect(service.bundles).toHaveBeenCalledWith(1);
        expect(result.current.bundles.map(b => b.id)).toEqual([12, 11]);
    });

    it('offers the app of the menu when it fits', () => {
        const {result} = setup({menuAppId: 2, platform: PlatformType.iOS});
        expect(result.current.appId).toBe(2);
    });

    it('splits the apps by platform', () => {
        const {result} = setup();
        expect(result.current.fittingApps.map(a => a.id)).toEqual([1]);
        expect(result.current.otherApps.map(a => a.id)).toEqual([2]);
    });

    it('lists all apps as fitting while the platform is unknown', () => {
        const {result} = setup({platform: undefined});
        expect(result.current.fittingApps).toHaveLength(2);
        expect(result.current.otherApps).toHaveLength(0);
    });

    it('has no app and no bundle for a project without apps', () => {
        const {result, service} = setup({apps: []});
        expect(result.current).toMatchObject({appId: '', bundleId: '', bundles: []});
        expect(service.bundles).not.toHaveBeenCalled();
    });
});

describe('useAppPicker choices', () => {
    it('loads the bundles of another app and forgets the bundle choice', async () => {
        const {result, service} = setup();
        await waitFor(() => expect(result.current.bundleId).toBe(12));
        act(() => result.current.chooseBundle(11));
        expect(result.current.bundleId).toBe(11);
        act(() => result.current.chooseApp(2));
        await waitFor(() => expect(service.bundles).toHaveBeenCalledWith(2));
        await waitFor(() => expect(result.current.bundleId).toBe(12));
        expect(result.current.appId).toBe(2);
    });

    it('shows no bundles of the previous app while the new ones load', async () => {
        const {result} = setup();
        await waitFor(() => expect(result.current.bundles).toHaveLength(2));
        act(() => result.current.chooseApp(2));
        expect(result.current.bundles).toEqual([]);
    });
});

describe('useAppPicker actions', () => {
    it('runs an action on the chosen bundle', async () => {
        const {result, service} = setup();
        await waitFor(() => expect(result.current.bundleId).toBe(12));
        act(() => result.current.act('start'));
        expect(service.appAction).toHaveBeenCalledWith('tok', 'start', {binaryId: 12, clearData: false});
    });

    it('clears data only when stopping', async () => {
        const {result, service} = setup();
        await waitFor(() => expect(result.current.bundleId).toBe(12));
        act(() => result.current.setClearData(true));
        act(() => result.current.act('install'));
        act(() => result.current.act('stop'));
        expect(vi.mocked(service.appAction).mock.calls.map(c => c[2])).toEqual([
            {binaryId: 12, clearData: false}, {binaryId: 12, clearData: true},
        ]);
    });

    it('does nothing without a lease or a bundle', async () => {
        const noLease = setup({lease: null});
        await waitFor(() => expect(noLease.result.current.bundleId).toBe(12));
        act(() => noLease.result.current.act('start'));
        const noBundle = setup({apps: []});
        act(() => noBundle.result.current.act('start'));
        expect(noLease.service.appAction).not.toHaveBeenCalled();
        expect(noBundle.service.appAction).not.toHaveBeenCalled();
    });
});

describe('useAppPicker failures', () => {
    it('reports a failure to load the bundles', async () => {
        const failure = new Error('down');
        const service = fakeService({bundles: vi.fn(async () => Promise.reject(failure))});
        const report = vi.fn();
        const input: AppPickerInput = {service, apps: APPS, platform: PlatformType.Android, menuAppId: null, lease: 'tok'};
        renderHook(() => useAppPicker(input, run, report));
        await waitFor(() => expect(report).toHaveBeenCalledWith(failure));
    });
});
