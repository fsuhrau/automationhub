import {ApiClient} from '../api/ApiClient';
import {IProjectOverview} from '../types/overview.project';

// OverviewMissing: the master predates the overview and needs a restart
// with the current version
export class OverviewMissing extends Error {
    constructor() {
        super('This master is older than the page. Restart it with the current version to see tests, runs and bundles here.');
    }
}

export interface OverviewService {
    get: (projectId: string) => Promise<IProjectOverview>,
}

export const overviewService = (api: ApiClient): OverviewService => ({
    get: async projectId => {
        const overview = await api.get<IProjectOverview | null>(`/${encodeURIComponent(projectId)}/overview`);
        // older masters don't have the route and answer with the web page
        if (!overview || !Array.isArray(overview.latest)) throw new OverviewMissing();
        return overview;
    },
});
