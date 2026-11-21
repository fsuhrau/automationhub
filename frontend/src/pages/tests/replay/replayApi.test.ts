import {expect, it, vi} from 'vitest';
import {ApiClient} from '../../../api/ApiClient';
import {fetchReplayRun} from './replayApi';

it('fetchReplayRun gets the run of the app test through the client', async () => {
    const get = vi.fn().mockResolvedValue({testRun: {id: 7}});
    const api = {get} as unknown as ApiClient;
    const result = await fetchReplayRun(api, {projectId: 'p1', appId: 2, testId: '3', runId: '7'});
    expect(get).toHaveBeenCalledWith('/p1/app/2/test/3/run/7');
    expect(result).toEqual({testRun: {id: 7}});
});
