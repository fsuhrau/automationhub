import {describe, expect, it} from 'vitest';
import {fireEvent, screen, waitFor, within} from '@testing-library/react';
import {BASE, controlled, loaded, PageWorld, renderPage} from './control/pageKit';

const choose = async (select: string, option: string): Promise<void> => {
    fireEvent.mouseDown(screen.getByRole('combobox', {name: select}));
    fireEvent.click(await within(screen.getByRole('listbox')).findByText(option));
};

const sent = (page: PageWorld, path: string): unknown[] => page.api.to('POST', `${BASE}/${path}`).map(c => c.data);

describe('the control page: the system log', () => {
    const openLog = async (): Promise<PageWorld> => {
        const page = renderPage();
        await loaded();
        fireEvent.click(screen.getByRole('tab', {name: 'System log'}));
        return page;
    };

    it('loads the log on demand, newest last', async () => {
        await openLog();
        expect(screen.getByText(/system log \(logcat on Android\)/)).toBeInTheDocument();
        fireEvent.click(screen.getByRole('button', {name: 'Load'}));
        expect(await screen.findByText('2 lines')).toBeInTheDocument();
        expect(within(screen.getByRole('log')).getAllByText(/App:/)).toHaveLength(2);
        expect(screen.getByRole('button', {name: 'Refresh'})).toBeInTheDocument();
    });

    it('finds text and keeps only warnings and errors', async () => {
        await openLog();
        fireEvent.click(screen.getByRole('button', {name: 'Load'}));
        await screen.findByText('2 lines');
        fireEvent.change(screen.getByLabelText('Find in the system log'), {target: {value: 'hello'}});
        expect(screen.getByText('1 of 2 lines')).toBeInTheDocument();
        fireEvent.change(screen.getByLabelText('Find in the system log'), {target: {value: ''}});
        fireEvent.click(screen.getByText('Warnings and errors'));
        expect(within(screen.getByRole('log')).getByText(/boom/)).toBeInTheDocument();
        expect(within(screen.getByRole('log')).queryByText(/hello/)).not.toBeInTheDocument();
    });

    it('loads the chosen number of lines', async () => {
        const page = await openLog();
        fireEvent.mouseDown(screen.getByRole('combobox'));
        fireEvent.click(await within(screen.getByRole('listbox')).findByText('last 1000 lines'));
        fireEvent.click(screen.getByRole('button', {name: 'Load'}));
        await screen.findByText('2 lines');
        expect(page.api.to('GET', `${BASE}/logs`)[0].config?.params).toEqual({lines: 1000});
    });
});

describe('the control page: the app', () => {
    it('offers the first app that runs on the device and its newest bundle', async () => {
        renderPage();
        await loaded();
        expect(screen.getByRole('combobox', {name: 'App'})).toHaveTextContent('Game');
        await waitFor(() => expect(screen.getByRole('combobox', {name: 'Bundle'})).toHaveTextContent('1.1 (#12)'));
        expect(screen.getByText('Take control to use these.')).toBeInTheDocument();
        expect(screen.getByRole('button', {name: 'Start'})).toBeDisabled();
    });

    it('starts the bundle under control', async () => {
        const page = await controlled();
        await waitFor(() => expect(screen.getByRole('button', {name: 'Start'})).toBeEnabled());
        fireEvent.click(screen.getByRole('button', {name: 'Start'}));
        await waitFor(() => expect(sent(page, 'app/start')).toEqual([{binaryId: 12, clearData: false}]));
        expect(page.api.to('POST', `${BASE}/app/start`)[0].config?.headers).toEqual({'X-Device-Lease': 'tok-1'});
    });
});

describe('the control page: app actions', () => {
    it('installs an older bundle and stops with clearing the data', async () => {
        const page = await controlled();
        await waitFor(() => expect(screen.getByRole('button', {name: 'Install'})).toBeEnabled());
        await choose('Bundle', '1.0');
        fireEvent.click(screen.getByRole('button', {name: 'Install'}));
        await waitFor(() => expect(sent(page, 'app/install')).toHaveLength(1));
        fireEvent.click(screen.getByRole('checkbox'));
        await waitFor(() => expect(screen.getByRole('button', {name: 'Stop'})).toBeEnabled());
        fireEvent.click(screen.getByRole('button', {name: 'Stop'}));
        await waitFor(() => expect(sent(page, 'app/stop')).toHaveLength(1));
        expect(sent(page, 'app/install')).toEqual([{binaryId: 11, clearData: false}]);
        expect(sent(page, 'app/stop')).toEqual([{binaryId: 11, clearData: true}]);
    });

    it('lists the bundles of another app', async () => {
        const page = renderPage();
        await loaded();
        await choose('App', 'Game iOS');
        await waitFor(() => expect(page.api.to('GET', '/p1/app/2/bundles')).toHaveLength(1));
        expect(screen.getByRole('combobox', {name: 'App'})).toHaveTextContent('Game iOS');
    });

    it('points to the bundles page when the app has none', async () => {
        renderPage({answers: {'GET /p1/app/1/bundles': []}});
        await loaded();
        expect(await screen.findByText(/This app has no bundles/)).toBeInTheDocument();
        expect(screen.getByRole('link', {name: 'Bundles'})).toHaveAttribute('href', '/project/p1/app:1/bundles');
    });

    it('shows why an action failed', async () => {
        await controlled({answers: {[`POST ${BASE}/app/start`]: new Error('install first')}});
        await waitFor(() => expect(screen.getByRole('button', {name: 'Start'})).toBeEnabled());
        fireEvent.click(screen.getByRole('button', {name: 'Start'}));
        expect(await screen.findByRole('alert')).toHaveTextContent('install first');
    });
});
