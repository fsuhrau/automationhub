import http from '../http-common';
import { AxiosResponse } from 'axios';
import ITestRunData from '../types/test.run';

class TestRunDataService {
    getAll(test_id: string): Promise<AxiosResponse<ITestRunData[]>> {
        return http.get(`/test/${test_id}/runs`);
    }

    getLast(test_id: string): Promise<AxiosResponse<ITestRunData>> {
        return http.get(`/test/${test_id}/runs/last`);
    }

    getRun(test_id: string, id: string): Promise<AxiosResponse<ITestRunData>> {
        return http.get(`/test/${test_id}/run/${id}`);
    }
}

export default new TestRunDataService();
