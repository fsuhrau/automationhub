import http from '../http-common';
import {IRunSummary} from '../types/test';

// a run on the project overview, with its test and app
export interface IOverviewRun extends IRunSummary {
    testId: number,
    testName: string,
    appId: number,
}

export interface IProjectOverview {
    apps: number,
    tests: number,
    bundles: number,
    bundleStorage: number,
    // tests whose latest run failed, was unstable or didn't run
    failing: IOverviewRun[],
    // latest runs that are still running
    running: IOverviewRun[],
    // the latest runs of the project, newest first
    latest: IOverviewRun[],
}

// OverviewMissing: the master predates the overview and needs a restart
// with the current version
export class OverviewMissing extends Error {
    constructor() {
        super('This master is older than the page. Restart it with the current version to see tests, runs and bundles here.');
    }
}

export const getOverview = (projectId: string): Promise<IProjectOverview> =>
    http.get(`/${encodeURIComponent(projectId)}/overview`).then(resp => {
        // older masters don't have the route and answer with the web page
        if (!resp.data || !Array.isArray(resp.data.latest)) throw new OverviewMissing();
        return resp.data;
    });
