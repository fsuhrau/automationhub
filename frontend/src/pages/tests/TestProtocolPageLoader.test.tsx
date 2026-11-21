import {describe, expect, it} from 'vitest';
import {screen, waitFor} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {fakeApi} from '../../testing/fakeApi';
import {FakeSse} from '../../testing/fakeSse';
import {protocol, run} from './run/harness/fixtures';
import {renderAt} from './run/harness/render';
import {base, login, openProtocol, url} from './run/harness/protocolPage';
import {currentLocation} from '../../testing/renderWithProviders';

describe('the protocol page loader', () => {
    it('sends the page of a wrapping runner to the test it wrapped', async () => {
        const runner = protocol({id: 5, testName: 'Game.Tests/Login'});
        const wrapped = login({id: 1, parentTestProtocolId: 5});
        openProtocol(runner, [wrapped], {at: '/project/p1/app:1/test/2/run/3/5?t=log'});
        await waitFor(() => expect(currentLocation()).toBe('/project/p1/app:1/test/2/run/3/1?t=log'));
    });

    it('shows the error of a protocol that cannot be loaded and retries', async () => {
        const api = fakeApi({[`GET ${base}/1`]: new Error('protocol is gone'), [`GET ${base}`]: {testRun: run(), nextRunId: null, prevRunId: null}});
        renderAt(url, {api, sse: new FakeSse()});
        expect(await screen.findByText('protocol is gone')).toBeInTheDocument();
        await userEvent.click(screen.getByRole('button', {name: 'Retry'}));
        await waitFor(() => expect(api.calls.filter(r => r.key === `GET ${base}/1`)).toHaveLength(2));
    });

    it('says a protocol is not in the run when the master answers with none', async () => {
        const api = fakeApi({[`GET ${base}/1`]: run({protocols: []}), [`GET ${base}`]: {testRun: run(), nextRunId: null, prevRunId: null}});
        renderAt(url, {api, sse: new FakeSse()});
        expect(await screen.findByText('protocol 1 not found in run 3')).toBeInTheDocument();
    });

    it('keeps loading while the answer is another protocol', async () => {
        const api = fakeApi({[`GET ${base}/1`]: run({protocols: [protocol({id: 9})]}), [`GET ${base}`]: {testRun: run(), nextRunId: null, prevRunId: null}});
        renderAt(url, {api, sse: new FakeSse()});
        await waitFor(() => expect(api.calls).toHaveLength(2));
        expect(screen.queryByRole('alert')).not.toBeInTheDocument();
        expect(screen.queryByRole('heading', {level: 1})).not.toBeInTheDocument();
    });

    it('shows the page without the rest of the run', async () => {
        const api = fakeApi({[`GET ${base}/1`]: run({protocols: [login()]}), [`GET ${base}`]: new Error('run missing')});
        renderAt(url, {api, sse: new FakeSse()});
        expect(await screen.findByRole('heading', {level: 1})).toHaveTextContent('Login');
        expect(screen.queryByText(/Test 1 of/)).not.toBeInTheDocument();
    });
});
