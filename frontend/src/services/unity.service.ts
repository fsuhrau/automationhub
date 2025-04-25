import axios from 'axios';
import IAppFunctionData from '../types/app.function';

const unityService = axios.create({
    baseURL: 'http://localhost:7109/',
    headers: {
        'Content-type': 'application/json',
    },
});

export const getTestFunctions = (): Promise<IAppFunctionData[]> => {
    return unityService.get('/tests').then(resp => resp.data)
};
