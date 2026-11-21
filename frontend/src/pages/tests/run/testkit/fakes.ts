// Fakes of the outside world for the tests of the run and protocol pages:
// the HTTP layer and the hub's event stream.
import {AxiosRequestConfig} from 'axios';
import {act} from '@testing-library/react';
import {ApiClient} from '../../../../api/ApiClient';

export interface FakeRequest {
    // "GET /p1/app/1/test/2/run/3"
    key: string,
    config?: AxiosRequestConfig,
    data?: unknown,
}

export interface FakeApi extends ApiClient {
    requests: FakeRequest[],
}

type Answer = unknown | ((request: FakeRequest) => unknown);

// fakeApi answers requests by their "METHOD url" key; an Error is rejected,
// a function is called with the request, everything else is the body
export const fakeApi = (answers: Record<string, Answer>): FakeApi => {
    const requests: FakeRequest[] = [];
    const respond = <T>(method: string, url: string, extra: Pick<FakeRequest, 'config' | 'data'> = {}): Promise<T> => {
        const request = {key: `${method} ${url}`, ...extra};
        requests.push(request);
        const answer = answers[request.key];
        if (answer === undefined) return Promise.reject(new Error(`no answer for ${request.key}`));
        const body = typeof answer === 'function' ? answer(request) : answer;
        return body instanceof Error ? Promise.reject(body) : Promise.resolve(body as T);
    };
    return {
        requests,
        get: (url, config) => respond('GET', url, {config}),
        post: (url, data, config) => respond('POST', url, {config, data}),
        put: (url, data, config) => respond('PUT', url, {config, data}),
        delete: (url, config) => respond('DELETE', url, {config}),
    };
};

type Listener = (event: MessageEvent) => void;

// FakeSse stands in for the hub's EventSource: pages subscribe through the
// SSEProvider, tests send events with emit
export class FakeSse {
    private readonly listeners = new Map<string, Set<Listener>>();

    // the source the SSEProvider asks for
    readonly source = (): EventSource => ({
        addEventListener: (name: string, listener: Listener) => this.listenersOf(name).add(listener),
        removeEventListener: (name: string, listener: Listener) => this.listenersOf(name).delete(listener),
        close: () => undefined,
    }) as unknown as EventSource;

    emit(name: string, data: unknown): void {
        act(() => this.listenersOf(name).forEach(listener => listener({data: JSON.stringify(data)} as MessageEvent)));
    }

    private listenersOf(name: string): Set<Listener> {
        const known = this.listeners.get(name) ?? new Set<Listener>();
        this.listeners.set(name, known);
        return known;
    }
}
