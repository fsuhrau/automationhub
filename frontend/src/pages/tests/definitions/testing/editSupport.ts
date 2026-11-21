import {screen} from '@testing-library/react';
import ITestData from '../../../../types/test';
import {createFakeApi, FakeApi, Reply} from './fakeApi';
import {FakeEventSource} from './fakeEventSource';
import {aTest, anApp} from './fixtures';
import {renderTestsApp} from './renderTestsApp';

export const TEST_PATH = '/project/proj/app:1/test/7';
export const NEW_PATH = '/project/proj/app:1/test/new';

export const DEVICES = 'GET /proj/devices';
export const UPDATE = 'PUT /proj/app/1/test/7';
export const CREATE = 'POST /proj/app/1/test';
export const UNITY = 'GET /tests';

export interface Opened {
    api: FakeApi,
    sse: FakeEventSource,
}

// openEditor renders the edit page of a test (or the new test page without
// one) with a fake master that also answers the pages saving leads to
export const openEditor = (test: ITestData | null, routes: Record<string, Reply> = {}): Opened => {
    const sse = new FakeEventSource();
    const api = createFakeApi({
        [DEVICES]: [],
        [UNITY]: [],
        'GET /proj/app/1/tests': [],
        'GET /proj/app/1/test/7': test ?? {},
        'GET /proj/app/1/test/7/runs': {runs: [], more: false},
        [UPDATE]: test ?? {},
        [CREATE]: {id: 21},
        'GET /proj/app/1/test/21': aTest({id: 21}),
        'GET /proj/app/1/test/21/runs': {runs: [], more: false},
        ...routes,
    });
    renderTestsApp(test ? `${TEST_PATH}/edit` : NEW_PATH, {api, apps: [anApp()], sse});
    return {api, sse};
};

export const location = (): string | null => screen.getByTestId('location').textContent;
