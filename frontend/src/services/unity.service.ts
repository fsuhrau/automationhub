import axios, { AxiosResponse } from 'axios';
import IAppFunctionData from '../types/app.function';

const unityService = axios.create({
    baseURL: 'http://localhost:7109/',
    headers: {
        'Content-type': 'application/json',
    },
});

export const getTestFunctions = (): Promise<AxiosResponse<IAppFunctionData[]>> => {
    return unityService.get('/tests');
};
