import http from '../http-common';
import ITestData from '../types/test';
import { AxiosResponse } from 'axios';
import ICreateTestData from '../types/request.create.test';

class TestDataService {
    getAll(): Promise<AxiosResponse<ITestData[]>> {
        return http.get('/tests');
    }

    get(id: string): Promise<AxiosResponse<ITestData>> {
        return http.get(`/test/${id}`);
    }

    create(data: ICreateTestData): Promise<AxiosResponse<ITestData>> {
        return http.post('/test', data);
    }

    update(data: ITestData, id: string): Promise<AxiosResponse<ITestData>> {
        return http.put(`/test/${id}`, data);
    }

    delete(id: string): Promise<AxiosResponse<void>> {
        return http.delete(`/test/${id}`);
    }

    findByName(name: string): Promise<AxiosResponse<ITestData[]>> {
        return http.get(`/tests?name=${name}`);
    }

    executeTest(id: number | null | undefined, appid: number): Promise<AxiosResponse<unknown>> {
        return http.post(`/test/${id}/run`, {
            AppID: appid,
        });
    }
}

export default new TestDataService();