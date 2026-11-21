import {IOverviewRun} from './overview.run';

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
