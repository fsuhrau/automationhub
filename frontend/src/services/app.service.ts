import http from '../http-common';
import IAppData from '../types/app';
import { AxiosResponse } from 'axios';

class AppDataService {
    getAll(): Promise<AxiosResponse<IAppData[]>> {
        return http.get('/apps');
    }

    get(id: string): Promise<AxiosResponse<IAppData | undefined>> {
        return http.get(`/app/${id}`);
    }

    create(data: IAppData): Promise<AxiosResponse<IAppData>> {
        return http.post('/app', data);
    }

    update(data: IAppData, id: number): Promise<AxiosResponse<IAppData>> {
        return http.put(`/app/${id}`, data);
    }

    delete(id: number): Promise<AxiosResponse<void>> {
        return http.delete(`/app/${id}`);
    }

    findByName(name: string): Promise<AxiosResponse<IAppData[]>> {
        return http.get(`/apps?name=${name}`);
    }
}

export default new AppDataService();
