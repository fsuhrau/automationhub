import { AxiosResponse } from 'axios';
import axios from 'axios';
import IAppFunctionData from "../types/app.function";

var unityService = axios.create({
    baseURL: 'http://localhost:7109/',
    headers: {
        'Content-type': 'application/json',
    },
});

export const getTestFunctions = (): Promise<AxiosResponse<IAppFunctionData[]>> => {
    return unityService.get('/tests');
};
