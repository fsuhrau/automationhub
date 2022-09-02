import http from '../http-common';
import { AxiosResponse } from 'axios';
import IHubStatsData from '../types/hub.stats';

export const getHubStats = (): Promise<AxiosResponse<IHubStatsData>> => {
    return http.get('/stats');
};
