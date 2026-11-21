import {describe, expect, it} from 'vitest';
import {OverviewMissing, overviewService} from './overview.service';
import {fakeApi} from '../testing/fakeApi';
import {IProjectOverview} from '../types/overview.project';

const overview: IProjectOverview = {apps: 1, tests: 2, bundles: 3, bundleStorage: 4, failing: [], running: [], latest: []};

describe('overviewService', () => {
    it('loads the overview of the project, encoding the identifier', async () => {
        const api = fakeApi({'GET /my%20proj/overview': overview});
        expect(await overviewService(api).get('my proj')).toBe(overview);
    });

    it.each([
        ['an answer that is not an overview (the web page of an older master)', '<html/>'],
        ['an empty answer', null],
        ['an overview without its latest runs', {apps: 1}],
    ])('says the master is too old for %s', async (_name, answer) => {
        await expect(overviewService(fakeApi({'GET /p/overview': answer})).get('p')).rejects.toBeInstanceOf(OverviewMissing);
    });

    it('passes on a failing request', async () => {
        await expect(overviewService(fakeApi({})).get('p')).rejects.toThrow('no fake answer');
    });

    it('explains what to do about an old master', () => {
        expect(new OverviewMissing().message).toContain('Restart it');
    });
});
