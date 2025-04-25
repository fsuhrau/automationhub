import http from '../http-common';
import IHubStatsData from '../types/hub.stats';

export const getHubStats = (projectId: string): Promise<IHubStatsData> => {
    return http.get('/stats').then(resp => resp.data)
};
