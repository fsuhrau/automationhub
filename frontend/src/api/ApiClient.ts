import type {AxiosInstance, AxiosRequestConfig} from 'axios';

// What features need from the HTTP layer. Every call resolves with the
// response body (`response.data`), not the axios response, and rejects with
// the axios error. A 401 with a redirect url is handled by the interceptor
// of the axios instance (src/http-common.ts), not here.
export interface ApiClient {
    get<T>(url: string, config?: AxiosRequestConfig): Promise<T>;

    post<T>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<T>;

    put<T>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<T>;

    delete<T>(url: string, config?: AxiosRequestConfig): Promise<T>;
}

const body = async <T, >(response: Promise<{ data: T }>): Promise<T> => (await response).data;

export const createApiClient = (http: AxiosInstance): ApiClient => ({
    get: <T, >(url: string, config?: AxiosRequestConfig) => body(http.get<T>(url, config)),
    post: <T, >(url: string, data?: unknown, config?: AxiosRequestConfig) => body(http.post<T>(url, data, config)),
    put: <T, >(url: string, data?: unknown, config?: AxiosRequestConfig) => body(http.put<T>(url, data, config)),
    delete: <T, >(url: string, config?: AxiosRequestConfig) => body(http.delete<T>(url, config)),
});
