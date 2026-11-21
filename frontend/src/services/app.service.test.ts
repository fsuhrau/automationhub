import {expect, it, vi} from 'vitest';
import {AxiosProgressEvent} from 'axios';
import {ApiClient} from '../api/ApiClient';
import {bundle} from '../pages/apps/appsFixtures';
import {AppParameter, IAppData} from '../types/app';
import {ParameterRequest} from '../types/settings.requests';
import {appService, bundleIconUrl, usableBundles} from './app.service';

interface Seen {
    method: string,
    url: string,
    data?: unknown,
    config?: unknown,
}

// answers every request with the request itself, so a test sees what was asked
const echo = (seen: Seen[]): ApiClient => {
    const reply = <T, >(request: Seen): Promise<T> => {
        seen.push(request);
        return Promise.resolve({method: request.method, url: request.url} as T);
    };
    return {
        get: (url, config) => reply({method: 'get', url, config}),
        post: (url, data, config) => reply({method: 'post', url, data, config}),
        put: (url, data, config) => reply({method: 'put', url, data, config}),
        delete: (url, config) => reply({method: 'delete', url, config}),
    };
};

interface UploadRequest {
    url: string,
    data: FormData,
    config: { headers: unknown, signal: AbortSignal, onUploadProgress: (e: AxiosProgressEvent) => void },
}

const ref = {projectId: 'p', appId: 3, id: 7};

type Service = ReturnType<typeof appService>;

it.each([
    {call: (s: Service) => s.list('p'), method: 'get', url: '/p/apps'},
    {call: (s: Service) => s.get('p', '3'), method: 'get', url: '/p/app/3'},
    {call: (s: Service) => s.create('p', {name: 'n', platform: 1, identifier: 'i'}), method: 'post', url: '/p/app'},
    {call: (s: Service) => s.update('p', 3, {} as IAppData), method: 'put', url: '/p/app/3'},
    {call: (s: Service) => s.addParameter('p', 3, {} as ParameterRequest), method: 'post', url: '/p/app/3/parameter'},
    {call: (s: Service) => s.updateParameter('p', 3, {id: 9} as AppParameter), method: 'put', url: '/p/app/3/parameter/9'},
    {call: (s: Service) => s.removeParameter('p', 3, 9), method: 'delete', url: '/p/app/3/parameter/9'},
    {call: (s: Service) => s.find({name: 'x'}), method: 'get', url: '/apps'},
    {call: (s: Service) => s.bundles(ref), method: 'get', url: '/p/app/3/bundles'},
    {call: (s: Service) => s.bundle(ref), method: 'get', url: '/p/app/3/bundle/7'},
    {call: (s: Service) => s.readBundle(ref), method: 'post', url: '/p/app/3/bundle/7/read'},
    {call: (s: Service) => s.archiveBundle(ref), method: 'post', url: '/p/app/3/bundle/7/archive'},
    {call: (s: Service) => s.deleteBundle(ref), method: 'delete', url: '/p/app/3/bundle/7'},
    {call: (s: Service) => s.updateTags(ref, 'a,b'), method: 'put', url: '/p/app/3/bundle/7'},
])('asks the master for $method $url', async ({call, method, url}) => {
    const seen: Seen[] = [];
    await call(appService(echo(seen)));
    expect(seen).toHaveLength(1);
    expect(seen[0]).toMatchObject({method, url});
});

it('sends the tags and the parameter as the body', async () => {
    const seen: Seen[] = [];
    const service = appService(echo(seen));
    await service.updateTags(ref, 'a,b');
    await service.updateParameter('p', 3, {id: 9, name: 'n'} as AppParameter);
    await service.find({name: 'x'});
    expect(seen[0].data).toEqual({tags: 'a,b'});
    expect(seen[1].data).toEqual({id: 9, name: 'n'});
    expect(seen[2].config).toEqual({params: {name: 'x'}});
});

it('uploads the file as a multipart form with the signal and reports the progress', async () => {
    const seen: Seen[] = [];
    const onProgress = vi.fn();
    const controller = new AbortController();
    await appService(echo(seen)).upload({projectId: 'p', appId: 3}, new File(['x'], 'a.apk'), {onProgress, signal: controller.signal});
    const {url, data, config} = seen[0] as unknown as UploadRequest;
    expect(url).toBe('/p/app/3/upload');
    expect((data.get('test_target') as File).name).toBe('a.apk');
    expect(config.headers).toEqual({'Content-Type': 'multipart/form-data'});
    expect(config.signal).toBe(controller.signal);
    config.onUploadProgress({loaded: 25, total: 100} as AxiosProgressEvent);
    config.onUploadProgress({loaded: 25} as AxiosProgressEvent);
    expect(onProgress).toHaveBeenCalledTimes(1);
    expect(onProgress).toHaveBeenCalledWith(0.25);
});

it('uploads without caring about progress', async () => {
    const seen: Seen[] = [];
    await appService(echo(seen)).upload({projectId: 'p', appId: 3}, new File(['x'], 'a.apk'));
    const {onUploadProgress} = seen[0].config as { onUploadProgress: (e: AxiosProgressEvent) => void };
    expect(() => onUploadProgress({loaded: 1, total: 2} as AxiosProgressEvent)).not.toThrow();
});

it('keeps the bundles that still have a file, newest first', () => {
    const list = [bundle(1), bundle(3, {archivedAt: '2020'}), bundle(2)];
    expect(usableBundles(list).map(b => b.id)).toEqual([2, 1]);
});

it('serves the icon of a bundle under the api', () => {
    expect(bundleIconUrl({projectId: 'a b', appId: 1}, 2)).toBe('/api/a%20b/app/1/bundle/2/icon');
});
