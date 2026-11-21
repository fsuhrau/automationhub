import {expect, it} from 'vitest';
import {fireEvent, screen, waitFor} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {AxiosProgressEvent} from 'axios';
import {IAppBinaryData} from '../../types/app';
import {bundle} from './appsFixtures';
import {Call, failure, fakeApi, FakeApi, renderBundles, signalOf} from './appsHarness';

const UPLOAD = 'POST /demo/app/3/upload';

const apk = (name = 'quest.apk'): File => new File(['x'.repeat(10)], name);

const pick = async (...files: File[]): Promise<void> => {
    await userEvent.upload(await screen.findByLabelText('Bundle files'), files);
};

const progress = (call: Call, loaded: number): void => {
    call.config?.onUploadProgress?.({loaded, total: 100} as AxiosProgressEvent);
};

interface Deferred {
    promise: Promise<IAppBinaryData>,
    resolve: (b: IAppBinaryData) => void,
    reject: (e: unknown) => void,
}

const deferred = (): Deferred => {
    let resolve!: Deferred['resolve'];
    let reject!: Deferred['reject'];
    const promise = new Promise<IAppBinaryData>((res, rej) => {
        resolve = res;
        reject = rej;
    });
    return {promise, resolve, reject};
};

const master = (upload: (call: Call) => unknown, list: () => IAppBinaryData[] = () => []): FakeApi =>
    fakeApi({'GET /demo/app/3/bundles': list, [UPLOAD]: upload});

it('shows the progress of an upload and adds the bundle when it is done', async () => {
    const answer = deferred();
    let stored: IAppBinaryData[] = [];
    renderBundles(master(call => {
        progress(call, 40);
        return answer.promise;
    }, () => stored), 'bundles');
    await pick(apk());
    expect(await screen.findByText('Uploading 40 %')).toBeInTheDocument();
    expect(screen.getByRole('progressbar')).toHaveAttribute('aria-valuenow', '40');
    stored = [bundle(9)];
    answer.resolve(bundle(9));
    expect(await screen.findByText('Uploaded as version 1.9.0 (#9)')).toBeInTheDocument();
    expect(await screen.findByText('just uploaded')).toBeInTheDocument();
});

it('says the master reads the file once all of it is sent', async () => {
    const answer = deferred();
    renderBundles(master(call => {
        progress(call, 100);
        return answer.promise;
    }), 'bundles');
    await pick(apk());
    expect(await screen.findByText('Reading the bundle on the master…')).toBeInTheDocument();
    answer.resolve(bundle(9));
    await screen.findByText(/Uploaded as version/);
});

it('sends the file as a form to the app', async () => {
    const api = master(() => bundle(9));
    renderBundles(api, 'bundles');
    await pick(apk('a.apk'));
    await screen.findByText(/Uploaded as version/);
    const sent = api.calls.find(c => c.url === '/demo/app/3/upload');
    expect((sent?.data as FormData).get('test_target')).toBeInstanceOf(File);
});

it('shows why an upload failed and lets it go', async () => {
    renderBundles(master(() => {
        throw failure('not an apk');
    }), 'bundles');
    await pick(apk());
    expect(await screen.findByText('not an apk')).toBeInTheDocument();
    await userEvent.click(screen.getByRole('button', {name: 'Close'}));
    expect(screen.queryByText('not an apk')).not.toBeInTheDocument();
});

it('tells that an identical file was uploaded before', async () => {
    renderBundles(master(() => {
        throw failure('exists', 409);
    }), 'bundles');
    await pick(apk());
    expect(await screen.findByText(/uploaded before/)).toBeInTheDocument();
});

it('does not send a file the app does not take', async () => {
    const api = master(() => bundle(9));
    renderBundles(api, 'bundles');
    await pick(apk('notes.txt'));
    expect(await screen.findByText('Pixel Quest takes .apk files.')).toBeInTheDocument();
    expect(api.calls.some(c => c.url.endsWith('/upload'))).toBe(false);
});

it('cancels the running upload and the waiting one', async () => {
    const answer = deferred();
    const api = master(call => new Promise((_, reject) => {
        signalOf(call)?.addEventListener('abort', () => reject(new Error('aborted')));
        return answer.promise;
    }));
    renderBundles(api, 'bundles');
    await pick(apk('a.apk'), apk('b.apk'));
    const cancels = await screen.findAllByRole('button', {name: 'Cancel'});
    await userEvent.click(cancels[1]);
    await userEvent.click(cancels[0]);
    await waitFor(() => expect(screen.getAllByText('Upload cancelled.')).toHaveLength(2));
    expect(api.calls.filter(c => c.url.endsWith('/upload'))).toHaveLength(1);
});

it('clears the finished uploads at once', async () => {
    renderBundles(master(() => bundle(9)), 'bundles');
    await pick(apk('a.apk'), apk('b.apk'));
    await userEvent.click(await screen.findByRole('button', {name: 'Clear all'}));
    expect(screen.queryByText(/Uploaded as version/)).not.toBeInTheDocument();
});

it('leaves the upload running list empty after leaving the page', async () => {
    const answer = deferred();
    const api = master(call => {
        signalOf(call)?.addEventListener('abort', () => answer.reject(new Error('left')));
        return answer.promise;
    });
    const {unmount} = renderBundles(api, 'bundles');
    await pick(apk());
    await screen.findByText('Uploading 0 %');
    unmount();
    await expect(answer.promise).rejects.toThrow('left');
});

it('covers the page while files are dragged over it and uploads what is dropped', async () => {
    renderBundles(master(() => bundle(9)), 'bundles');
    const root = (await screen.findByLabelText('Bundle files')).parentElement as HTMLElement;
    const files = [apk()];
    const transfer = {files, items: files.map(file => ({kind: 'file', type: file.type, getAsFile: () => file})), types: ['Files']};
    fireEvent.dragEnter(root, {dataTransfer: transfer});
    expect(await screen.findByText('Drop to upload to Pixel Quest')).toBeInTheDocument();
    fireEvent.drop(root, {dataTransfer: transfer});
    expect(await screen.findByText(/Uploaded as version/)).toBeInTheDocument();
});
