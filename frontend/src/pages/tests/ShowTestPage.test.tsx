import {act, screen, waitFor, within} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {describe, expect, it} from 'vitest';
import ITestData from '../../types/test';
import {TestExecutionType} from '../../types/test.execution.type.enum';
import {TestType} from '../../types/test.type.enum';
import {UnityTestCategory} from '../../types/unity.test.category.type.enum';
import {Call, fakeApi, failure, FakeApi, Reply} from '../../testing/fakeApi';
import {FakeSse} from '../../testing/fakeSse';
import {aDevice, aFunction, aRunEntry, aTest, anApp} from './definitions/harness/fixtures';
import {renderTestsApp} from './definitions/harness/renderTestsApp';
import {currentLocation} from '../../testing/renderWithProviders';

const PATH = '/project/proj/app:1/test/7';
const RUNS = 'GET /proj/app/1/test/7/runs';
const page = (runs: Reply, test: ITestData = aTest()): {api: FakeApi, sse: FakeSse} => {
    const sse = new FakeSse();
    const api = fakeApi({'GET /proj/app/1/test/7': test, [RUNS]: runs, 'POST /proj/app/1/test/7/run': {id: 12}});
    renderTestsApp(PATH, {api, apps: [anApp()], sse});
    return {api, sse};
};
const withUnity = (testCategoryType: UnityTestCategory, rest: object = {}): ITestData => {
    const test = aTest();
    test.testConfig.unity = {...test.testConfig.unity!, testCategoryType, ...rest};
    return test;
};


describe('ShowTestPage runs', () => {
    it('shows the test with its stats and runs', async () => {
        const bundle = {id: 3, name: 'hero.apk', version: '1.2.0'};
        page({
            runs: [
                aRunEntry({id: 12, failed: 1, passed: 2, parameter: 'level=3;fast', bundle}),
                aRunEntry({id: 11, status: 'running', endedAt: null}),
                aRunEntry({id: 10, bundle: {...bundle, version: ''}}),
            ], more: false,
        });
        expect(await screen.findByRole('heading', {name: 'Smoke'})).toBeInTheDocument();
        const table = await screen.findByRole('table', {name: 'Runs of this test'});
        expect(within(table).getByText('#12')).toBeInTheDocument();
        expect(within(table).getByText('2 passed · 1 failed')).toBeInTheDocument();
        expect(within(table).getByText('1.2.0')).toBeInTheDocument();
        expect(within(table).getByText('level=3 · fast')).toBeInTheDocument();
        expect(within(table).getByText('hero.apk')).toBeInTheDocument();
        expect(within(table).getByText('running')).toBeInTheDocument();
        expect(screen.getByText('Last run')).toBeInTheDocument();
        expect(screen.getByText('Usually takes')).toBeInTheDocument();
        expect(screen.getByText('2 of 2')).toBeInTheDocument();
    });

    it('shows dashes for stats without a verdict', async () => {
        page({runs: [aRunEntry({status: 'running', endedAt: null})], more: false});
        expect(await screen.findByText('not yet')).toBeInTheDocument();
        expect(screen.getAllByText('–').length).toBeGreaterThan(2);
    });

    it('says when the master does not summarize runs', async () => {
        page({runs: [aRunEntry({status: undefined})], more: false});
        expect(await screen.findByText(/doesn't summarize runs yet/)).toBeInTheDocument();
        expect(screen.queryByText('Usually takes')).not.toBeInTheDocument();
    });

});

describe('ShowTestPage paging and errors', () => {
    it('loads older runs on request', async () => {
        const {api} = page(({params}: Call) => (params as {before?: number}).before
            ? {runs: [aRunEntry({id: 5})], more: false}
            : {runs: [aRunEntry({id: 12})], more: true});
        await userEvent.click(await screen.findByRole('button', {name: 'Older runs'}));
        expect(await screen.findByText('#5')).toBeInTheDocument();
        expect(screen.getByText('#12')).toBeInTheDocument();
        expect(screen.queryByRole('button', {name: 'Older runs'})).not.toBeInTheDocument();
        expect(api.called(RUNS)[1].params).toEqual({limit: 25, before: 12});
    });

    it('shows the error and retries', async () => {
        let fail = true;
        page(() => {
            if (fail) throw failure('runs are gone', 500);
            return {runs: [aRunEntry()], more: false};
        });
        expect(await screen.findByText('runs are gone')).toBeInTheDocument();
        fail = false;
        await userEvent.click(screen.getByRole('button', {name: 'Retry'}));
        expect(await screen.findByText('#11')).toBeInTheDocument();
    });

});

describe('ShowTestPage run links', () => {
    it('opens a run from its row and the latest run from the section', async () => {
        page({runs: [aRunEntry({id: 12})], more: false});
        await userEvent.click(await screen.findByText('#12'));
        expect(currentLocation()).toBe('/project/proj/app:1/test/7/run/12');
    });

    it('opens the latest run', async () => {
        page({runs: [aRunEntry()], more: false});
        await userEvent.click(await screen.findByRole('link', {name: 'Latest run'}));
        expect(currentLocation()).toBe('/project/proj/app:1/test/7/runs/last');
    });

    it('opens a run from its row without a click on the id', async () => {
        page({runs: [aRunEntry({id: 12})], more: false});
        const row = (await screen.findByText('Passed', {selector: '.MuiChip-label'})).closest('tr')!;
        await userEvent.click(row);
        expect(currentLocation()).toBe('/project/proj/app:1/test/7/run/12');
    });

    it('reloads the runs when a run finished', async () => {
        const {api, sse} = page({runs: [aRunEntry()], more: false});
        await screen.findByText('#11');
        act(() => sse.emit('test_runs', {projectId: 'proj', appId: 1, testId: 7, testRunId: 12, state: 'started'}));
        await waitFor(() => expect(api.called(RUNS)).toHaveLength(2));
    });
});

describe('ShowTestPage configuration', () => {
    it.each([
        ['all tests', withUnity(UnityTestCategory.RunAllTests), 'All Unity tests of the app.'],
        ['a category', withUnity(UnityTestCategory.RunAllOfCategory, {categories: 'fast'}), 'fast'],
        ['no category', withUnity(UnityTestCategory.RunAllOfCategory), /a category \(none set\)/],
    ])('says what it runs for %s', async (_, test, text) => {
        page({runs: [], more: false}, test);
        expect(await screen.findByText(text)).toBeInTheDocument();
    });

    it('lists the picked tests by class', async () => {
        const functions = [aFunction('Hit'), aFunction('Block'), aFunction('Walk', 'Ns.Move')].map(f => ({...f, testConfigUnityId: 1}));
        page({runs: [], more: false}, withUnity(UnityTestCategory.RunSelectedTestsOnly, {testFunctions: functions}));
        expect(await screen.findByText(/3 picked Unity tests:/)).toBeInTheDocument();
        expect(screen.getByText('Ns.Fight')).toBeInTheDocument();
        expect(screen.getByText('Block')).toBeInTheDocument();
    });

    it('says a single picked test in the singular', async () => {
        page({runs: [], more: false}, withUnity(UnityTestCategory.RunSelectedTestsOnly, {testFunctions: [{...aFunction('Hit'), testConfigUnityId: 1}]}));
        expect(await screen.findByText(/1 picked Unity test:/)).toBeInTheDocument();
    });

    it('names other test types', async () => {
        page({runs: [], more: false}, aTest({}, {type: TestType.Cocos}));
        expect(await screen.findByText('A Cocos test.')).toBeInTheDocument();
    });

});

describe('ShowTestPage devices', () => {
    it('lists the picked devices as links and says how tests spread', async () => {
        const devices = [
            {deviceId: 5, device: aDevice({alias: 'Lab phone'})}, {deviceId: 6, device: aDevice({alias: '', name: 'Pixel'})}, {deviceId: 8, device: null},
        ].map(d => ({...d, testConfigId: 1, createdAt: new Date(), updatedAt: new Date(), deletedAt: new Date()}));
        page({runs: [], more: false}, aTest({}, {allDevices: false, devices, executionType: TestExecutionType.Simultaneously}));
        const link = await screen.findByRole('link', {name: 'Lab phone'});
        expect(link).toHaveAttribute('href', '/project/proj/device/5');
        expect(screen.getByRole('link', {name: 'Pixel'})).toBeInTheDocument();
        expect(screen.getByRole('link', {name: 'Device 8'})).toBeInTheDocument();
        expect(screen.getByText('Every test runs on every device, for results across devices.')).toBeInTheDocument();
    });

    it('says on all devices, or that no device is picked', async () => {
        page({runs: [], more: false});
        expect(await screen.findByText('On all devices.')).toBeInTheDocument();
        expect(screen.getByText('Each test runs on the next free device, for faster results.')).toBeInTheDocument();
    });

    it('warns when no device is picked', async () => {
        page({runs: [], more: false}, aTest({}, {allDevices: false, devices: []}));
        expect(await screen.findByText(/No devices are picked/)).toBeInTheDocument();
    });

    it('opens the editor from the header', async () => {
        page({runs: [], more: false});
        await userEvent.click((await screen.findAllByRole('link', {name: 'Edit'}))[0]);
        expect(currentLocation()).toBe('/project/proj/app:1/test/7/edit');
    });
});

describe('ShowTestPage starting a run', () => {
    it('starts the test from the header with the given parameters', async () => {
        const {api} = page({runs: [aRunEntry()], more: false});
        await userEvent.click(await screen.findByRole('button', {name: 'Run'}));
        await userEvent.click(await screen.findByRole('button', {name: 'Start run'}));
        await waitFor(() => expect(currentLocation()).toBe('/project/proj/app:1/test/7/run/12'));
        expect(api.called('POST /proj/app/1/test/7/run')[0].data).toEqual({appBinaryId: null, params: '', startUrl: ''});
    });

    it('starts a test that never ran from the empty state', async () => {
        page({runs: [], more: false});
        expect(await screen.findByText("This test hasn't run yet.")).toBeInTheDocument();
        await userEvent.click(screen.getByRole('button', {name: 'Run it'}));
        expect(await screen.findByRole('dialog', {name: 'Run Smoke'})).toBeInTheDocument();
        await userEvent.click(screen.getByRole('button', {name: 'Cancel'}));
        await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());
    });

    it('keeps the bundle list out of the way for an editor app', async () => {
        page({runs: [], more: false});
        await userEvent.click(await screen.findByRole('button', {name: 'Run it'}));
        expect(screen.queryByText('Bundle')).not.toBeInTheDocument();
    });
});
