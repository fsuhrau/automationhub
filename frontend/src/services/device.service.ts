import http from '../http-common';
import IDeviceData from '../types/device';
import { AxiosResponse } from 'axios';
import ITestRunData from '../types/test.run';

class DeviceDataService {
    getAll(): Promise<AxiosResponse<IDeviceData[]>> {
        return http.get('/devices');
    }

    get(id: string): Promise<AxiosResponse<IDeviceData | undefined>> {
        return http.get(`/device/${id}`);
    }

    create(data: IDeviceData): Promise<AxiosResponse<IDeviceData>> {
        return http.post('/device', data);
    }

    update(data: IDeviceData, id: number): Promise<AxiosResponse<IDeviceData>> {
        return http.put(`/device/${id}`, data);
    }

    delete(id: number): Promise<AxiosResponse<void>> {
        return http.delete(`/device/${id}`);
    }

    findByName(name: string): Promise<AxiosResponse<IDeviceData[]>> {
        return http.get(`/devices?name=${name}`);
    }

    runTests(id: number | null | undefined): Promise<AxiosResponse<ITestRunData>>  {
        return http.post(`/device/${id}/tests`, null);
    }
}

export default new DeviceDataService();
