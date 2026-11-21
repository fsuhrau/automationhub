import {expect, it} from 'vitest';
import {act, renderHook, RenderHookResult, waitFor} from '@testing-library/react';
import {bundle} from './appsFixtures';
import {Call, fakeApi, FakeApi} from '../../testing/fakeApi';
import {wrapperOf} from '../../testing/renderWithProviders';
import {signalOf} from './appsHarness';
import {acceptFor} from './bundlesModel';
import {BundleUploadOptions, BundleUploads, useBundleUpload} from './useBundleUpload';
import {PlatformType} from '../../types/platform.type.enum';

const accept = acceptFor(PlatformType.Android);

interface Setup extends RenderHookResult<BundleUploads, unknown> {
    api: FakeApi,
    uploaded: number[],
}

const setup = (upload: (call: Call) => unknown, over: Partial<BundleUploadOptions> = {}): Setup => {
    const api = fakeApi({'POST /p/app/1/upload': upload});
    const uploaded: number[] = [];
    const hook = renderHook(() => useBundleUpload({
        app: {projectId: 'p', appId: 1}, accept, onUploaded: b => uploaded.push(b.id), ...over,
    }), {wrapper: wrapperOf({api})});
    return {...hook, api, uploaded};
};

it('uploads files one after the other', async () => {
    const order: string[] = [];
    const {result, uploaded} = setup(call => {
        order.push(((call.data as FormData).get('test_target') as File).name);
        return bundle(order.length);
    });
    act(() => result.current.upload([new File(['a'], 'a.apk'), new File(['b'], 'b.apk')]));
    await waitFor(() => expect(uploaded).toEqual([1, 2]));
    expect(order).toEqual(['a.apk', 'b.apk']);
    expect(result.current.busy).toBe(false);
});

it('uploads nothing while the app is unknown', () => {
    const {result, api} = setup(() => bundle(1), {app: null});
    act(() => result.current.upload([new File(['a'], 'a.apk')]));
    expect(result.current.uploads).toEqual([]);
    expect(api.calls).toEqual([]);
});

it('does not start an upload that was cancelled while it waited', async () => {
    let release: (() => void) | undefined;
    const {result, api} = setup(() => new Promise(resolve => {
        release = () => resolve(bundle(1));
    }));
    act(() => result.current.upload([new File(['a'], 'a.apk'), new File(['b'], 'b.apk')]));
    act(() => result.current.cancel(result.current.uploads[1].id));
    await waitFor(() => expect(release).toBeDefined());
    release?.();
    await waitFor(() => expect(result.current.uploads[1].state.phase).toBe('cancelled'));
    expect(api.calls).toHaveLength(1);
});

it('cancelling an upload that is gone does nothing', () => {
    const {result} = setup(() => bundle(1));
    expect(() => act(() => result.current.cancel(99))).not.toThrow();
});

it('aborts the running upload when the page is left', async () => {
    let signal: AbortSignal | undefined;
    const {result, unmount} = setup(call => {
        signal = signalOf(call);
        return new Promise(() => undefined);
    });
    act(() => result.current.upload([new File(['a'], 'a.apk')]));
    await waitFor(() => expect(signal).toBeDefined());
    unmount();
    expect(signal?.aborted).toBe(true);
});
