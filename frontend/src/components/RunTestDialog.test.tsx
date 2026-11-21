import {fireEvent, render, screen, waitFor, within} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {AxiosProgressEvent} from 'axios';
import {MemoryRouter, useLocation} from 'react-router-dom';
import {afterEach, describe, expect, it, vi} from 'vitest';
import {ApiClientProvider} from '../api/ApiClientProvider';
import {createFakeApi, failure, FakeApi, FakeRequest, Reply} from '../pages/tests/definitions/testing/fakeApi';
import {aBundle, aParameter, anApp} from '../pages/tests/definitions/testing/fixtures';
import {IAppData} from '../types/app';
import {PlatformType} from '../types/platform.type.enum';
import RunTestDialog from './RunTestDialog';

const BUNDLES = 'GET /proj/app/1/bundles';
const UPLOAD = 'POST /proj/app/1/upload';
const RUN = 'POST /proj/app/1/test/7/run';
const android = (over: Partial<IAppData> = {}): IAppData => anApp({platform: PlatformType.Android, ...over});
const bundles = [aBundle({id: 1, version: '1.0'}), aBundle({id: 3, version: '1.2'}), aBundle({id: 2, version: '0.9', archivedAt: new Date().toISOString()})];

let api: FakeApi;
let onClose = vi.fn();
const Where: React.FC = () => <p data-testid="location">{useLocation().pathname}</p>;

const show = (app: IAppData, routes: Record<string, Reply> = {}, test: {id: number, name: string} | null = {id: 7, name: 'Smoke'}): void => {
    api = createFakeApi({[BUNDLES]: bundles, [RUN]: {id: 12}, ...routes});
    onClose = vi.fn();
    render(<ApiClientProvider client={api.client}>
        <MemoryRouter><Where/><RunTestDialog projectIdentifier="proj" app={app} test={test} onClose={onClose}/></MemoryRouter>
    </ApiClientProvider>);
};
const start = (): Promise<void> => userEvent.click(screen.getByRole('button', {name: 'Start run'}));
const sent = (): unknown => api.calledWith(RUN)[0].data;
const file = (name: string): File => new File(['x'], name);
const upload = (name = 'new.apk'): Promise<void> => userEvent.upload(fileInput(), file(name));
const uploaded = aBundle({id: 9, version: '2.0'});
const fileInput = (): HTMLInputElement => screen.getByLabelText('Bundle file');

afterEach(() => api.restore());

describe('RunTestDialog bundles', () => {
    it('installs the newest usable bundle by default and opens the new run', async () => {
        show(android());
        expect(await screen.findByText('The test installs this build on its devices first.')).toBeInTheDocument();
        await waitFor(() => expect(screen.getByRole('button', {name: 'Start run'})).toBeEnabled());
        await start();
        await waitFor(() => expect(screen.getByTestId('location')).toHaveTextContent('/project/proj/app:1/test/7/run/12'));
        expect(sent()).toEqual({appBinaryId: 3, params: '', startUrl: ''});
        expect(onClose).toHaveBeenCalled();
    });

    it('installs the bundle that was picked', async () => {
        show(android());
        await waitFor(() => expect(screen.getByRole('button', {name: 'Start run'})).toBeEnabled());
        await userEvent.click(screen.getByRole('combobox'));
        const options = screen.getAllByRole('option');
        expect(options).toHaveLength(2);
        expect(within(options[0]).getByText('1.2 (newest)')).toBeInTheDocument();
        await userEvent.click(options[1]);
        await start();
        await waitFor(() => expect(sent()).toMatchObject({appBinaryId: 1}));
    });

    it('cannot start without a bundle and says to upload one', async () => {
        show(android(), {[BUNDLES]: []});
        expect(await screen.findByText('No bundles yet; upload .apk to run the test.')).toBeInTheDocument();
        expect(screen.getByRole('button', {name: 'Start run'})).toBeDisabled();
        fireEvent.submit(screen.getByRole('button', {name: 'Start run'}).closest('form')!);
        expect(api.calledWith(RUN)).toHaveLength(0);
    });

    it('shows that bundles are loading', async () => {
        show(android(), {[BUNDLES]: () => new Promise(() => undefined)});
        expect(await screen.findByText('Loading bundles…')).toBeInTheDocument();
    });

    it('shows when the bundles cannot be loaded', async () => {
        show(android(), {[BUNDLES]: () => {
            throw failure(500, 'no bundles for you');
        }});
        expect(await screen.findByText('no bundles for you')).toBeInTheDocument();
    });
});

describe('RunTestDialog upload', () => {
    it('uploads a bundle with progress and picks it', async () => {
        let finish: () => void = () => undefined;
        const gate = new Promise<void>(resolve => {
            finish = resolve;
        });
        show(android(), {[UPLOAD]: async (r: FakeRequest) => {
            r.config.onUploadProgress?.({loaded: 1, total: 2} as AxiosProgressEvent);
            r.config.onUploadProgress?.({loaded: 5, total: 0} as AxiosProgressEvent);
            await gate;
            r.config.onUploadProgress?.({loaded: 2, total: 2} as AxiosProgressEvent);
            return uploaded;
        }});
        await screen.findByText('The test installs this build on its devices first.');
        await userEvent.click(screen.getByRole('button', {name: 'Upload'}));
        await upload();
        expect(await screen.findByText('Uploading 50 %')).toBeInTheDocument();
        expect(screen.getByRole('button', {name: 'Start run'})).toBeDisabled();
        finish();
        await waitFor(() => expect(screen.queryByText(/Uploading|Reading the bundle/)).not.toBeInTheDocument());
        await waitFor(() => expect(screen.getByRole('button', {name: 'Start run'})).toBeEnabled());
        await start();
        await waitFor(() => expect(sent()).toMatchObject({appBinaryId: 9}));
    });

    it('says when the master reads the file after the upload', async () => {
        show(android(), {[UPLOAD]: async (r: FakeRequest) => {
            r.config.onUploadProgress?.({loaded: 2, total: 2} as AxiosProgressEvent);
            await new Promise(() => undefined);
        }});
        await screen.findByText('The test installs this build on its devices first.');
        await upload();
        expect(await screen.findByText('Reading the bundle on the master…')).toBeInTheDocument();
    });

});

describe('RunTestDialog upload errors', () => {
    it('refuses a file of another kind', async () => {
        show(android());
        await screen.findByText('The test installs this build on its devices first.');
        await userEvent.setup({applyAccept: false}).upload(fileInput(), file('new.txt'));
        expect(await screen.findByText('Hero takes .apk files.')).toBeInTheDocument();
        expect(api.calledWith(UPLOAD)).toHaveLength(0);
    });

    it.each([
        ['an identical file', failure(409, 'conflict'), 'This exact file was uploaded before; pick it in the list.'],
        ['another failure', failure(500, 'disk full'), 'disk full'],
    ])('shows the error of %s', async (_, error, message) => {
        show(android(), {[UPLOAD]: () => {
            throw error;
        }});
        await screen.findByText('The test installs this build on its devices first.');
        await upload();
        expect(await screen.findByText(message)).toBeInTheDocument();
    });

    it('ignores a selection without a file', async () => {
        show(android());
        await screen.findByText('The test installs this build on its devices first.');
        fireEvent.change(fileInput(), {target: {files: []}});
        expect(api.calledWith(UPLOAD)).toHaveLength(0);
    });
});

describe('RunTestDialog parameters and start', () => {
    const params = [
        aParameter({id: 1, name: 'level', type: {type: 'string', defaultValue: '2'}}),
        aParameter({id: 2, name: 'mode', type: {type: 'option', options: ['easy', 'hard'], defaultValue: 'easy'} as never}),
        aParameter({id: 3, name: 'note', type: {type: 'string', defaultValue: ''}}),
    ];

    it('starts with the defaults, edited text and a picked option, leaving empty ones out', async () => {
        show(anApp({parameter: params}));
        const level = await screen.findByLabelText('level');
        await userEvent.clear(level);
        await userEvent.type(level, '5');
        await userEvent.click(screen.getByRole('combobox', {name: 'mode'}));
        expect(screen.getByRole('option', {name: 'easy (default)'})).toBeInTheDocument();
        await userEvent.click(screen.getByRole('option', {name: 'hard'}));
        await start();
        await waitFor(() => expect(sent()).toEqual({appBinaryId: null, params: 'level=5;mode=hard', startUrl: ''}));
        expect(screen.getByText("Empty parameters aren't passed.")).toBeInTheDocument();
    });

    it('needs the start URL of a web app', async () => {
        show(anApp({platform: PlatformType.Web}));
        expect(await screen.findByRole('button', {name: 'Start run'})).toBeDisabled();
        await userEvent.type(screen.getByLabelText(/Start URL/), 'https://game.test');
        await start();
        await waitFor(() => expect(sent()).toEqual({appBinaryId: null, params: '', startUrl: 'https://game.test'}));
    });

});

describe('RunTestDialog start', () => {
    it('shows the error when the run does not start and lets the user try again', async () => {
        let fail = true;
        show(anApp(), {[RUN]: () => {
            if (fail) throw failure(409, 'no free device');
            return {id: 12};
        }});
        await userEvent.click(await screen.findByRole('button', {name: 'Start run'}));
        expect(await screen.findByText('no free device')).toBeInTheDocument();
        expect(onClose).not.toHaveBeenCalled();
        fail = false;
        await userEvent.click(screen.getByRole('button', {name: 'Start run'}));
        await waitFor(() => expect(onClose).toHaveBeenCalled());
    });

    it('closes on Cancel', async () => {
        show(anApp());
        await userEvent.click(await screen.findByRole('button', {name: 'Cancel'}));
        expect(onClose).toHaveBeenCalled();
    });

    it('renders nothing without a test', () => {
        show(anApp(), {}, null);
        expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
    });
});
