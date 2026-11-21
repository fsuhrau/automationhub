import {useMemo} from 'react';
import http from '../http-common';
import {ApiClient, createApiClient} from '../api/ApiClient';
import {useApiClient} from '../api/ApiClientProvider';
import {AppParameter, IAppBinaryData, IAppBundleDetail, IAppData} from '../types/app';
import {AppRef, BundleRef} from '../types/app.ref';
import {AppFilter} from '../types/app.filter';
import {UploadOptions} from '../types/bundle.upload';

export interface AppService {
    list(projectId: string): Promise<IAppData[]>;

    get(projectId: string, id: string): Promise<IAppData | undefined>;

    create(projectId: string, data: IAppData): Promise<IAppData>;

    update(projectId: string, appId: number, data: IAppData): Promise<IAppData>;

    addParameter(projectId: string, appId: number, data: AppParameter): Promise<AppParameter>;

    // the parameter is saved under its own id
    updateParameter(projectId: string, appId: number, parameter: AppParameter): Promise<AppParameter>;

    removeParameter(projectId: string, appId: number, id: number): Promise<void>;

    find(filter?: AppFilter): Promise<IAppBinaryData[]>;

    bundles(app: AppRef): Promise<IAppBinaryData[]>;

    // one bundle with everything read from its file and the latest runs that used it
    bundle(ref: BundleRef): Promise<IAppBundleDetail>;

    // has the master read a stored bundle file again
    readBundle(ref: BundleRef): Promise<IAppBundleDetail>;

    // removes the file of a bundle and keeps its details, so runs with it
    // stay comparable; uploading the same file again brings it back
    archiveBundle(ref: BundleRef): Promise<IAppBundleDetail>;

    deleteBundle(ref: BundleRef): Promise<void>;

    // sets the comma separated tags of a bundle
    updateTags(ref: BundleRef, tags: string): Promise<IAppBinaryData>;

    // uploads a build; progress is 0..1 while the file is sent, then the master
    // reads the file before it answers with the bundle. An identical file
    // uploaded before is a 409.
    upload(app: AppRef, file: File, options?: UploadOptions): Promise<IAppBinaryData>;
}

const appBase = ({projectId, appId}: AppRef): string => `/${projectId}/app/${appId}`;
const bundleBase = (ref: BundleRef): string => `${appBase(ref)}/bundle/${ref.id}`;

const uploadForm = (file: File): FormData => {
    const form = new FormData();
    form.append('test_target', file);
    return form;
};

export const appService = (api: ApiClient): AppService => ({
    list: projectId => api.get(`/${projectId}/apps`),
    get: (projectId, id) => api.get(`/${projectId}/app/${id}`),
    create: (projectId, data) => api.post(`/${projectId}/app`, data),
    update: (projectId, appId, data) => api.put(`/${projectId}/app/${appId}`, data),
    addParameter: (projectId, appId, data) => api.post(`/${projectId}/app/${appId}/parameter`, data),
    updateParameter: (projectId, appId, parameter) => api.put(`/${projectId}/app/${appId}/parameter/${parameter.id}`, parameter),
    removeParameter: (projectId, appId, id) => api.delete(`/${projectId}/app/${appId}/parameter/${id}`),
    find: filter => api.get('/apps', {params: filter}),
    bundles: app => api.get(`${appBase(app)}/bundles`),
    bundle: ref => api.get(bundleBase(ref)),
    readBundle: ref => api.post(`${bundleBase(ref)}/read`),
    archiveBundle: ref => api.post(`${bundleBase(ref)}/archive`),
    deleteBundle: ref => api.delete(bundleBase(ref)),
    updateTags: (ref, tags) => api.put(bundleBase(ref), {tags}),
    upload: (app, file, {onProgress, signal} = {}) => api.post(`${appBase(app)}/upload`, uploadForm(file), {
        headers: {'Content-Type': 'multipart/form-data'},
        signal,
        onUploadProgress: e => {
            if (e.total) onProgress?.(e.loaded / e.total);
        },
    }),
});

export const useAppService = (): AppService => {
    const api = useApiClient();
    return useMemo(() => appService(api), [api]);
};

// usableBundles are the bundles that still have their file, newest first
export const usableBundles = (list: IAppBinaryData[]): IAppBinaryData[] =>
    list.filter(b => !b.archivedAt).sort((a, b) => b.id - a.id);

export const bundleIconUrl = ({projectId, appId}: AppRef, id: number): string =>
    `/api/${encodeURIComponent(projectId)}/app/${appId}/bundle/${id}/icon`;

const legacy = appService(createApiClient(http));

/** @deprecated legacy wrapper, removed in F13 */
export const getAllApps = (projectId: string): Promise<IAppData[]> => legacy.list(projectId);

/** @deprecated legacy wrapper, removed in F13 */
export const createApp = (projectId: string, data: IAppData): Promise<IAppData> => legacy.create(projectId, data);

/** @deprecated legacy wrapper, removed in F13 */
export const updateApp = (projectId: string, appId: number, data: IAppData): Promise<IAppData> => legacy.update(projectId, appId, data);

/** @deprecated legacy wrapper, removed in F13 */
export const addAppParameter = (projectId: string, appId: number, data: AppParameter): Promise<AppParameter> =>
    legacy.addParameter(projectId, appId, data);

/** @deprecated legacy wrapper, removed in F13 */
// eslint-disable-next-line max-params
export const updateAppParameter = (projectId: string, appId: number, id: number, data: AppParameter): Promise<AppParameter> =>
    legacy.updateParameter(projectId, appId, {...data, id});

/** @deprecated legacy wrapper, removed in F13 */
export const removeAppParameter = (projectId: string, appId: number, id: number): Promise<void> => legacy.removeParameter(projectId, appId, id);

/** @deprecated legacy wrapper, removed in F13 */
export const getAppBundles = (projectId: string, appId: number): Promise<IAppBinaryData[]> => legacy.bundles({projectId, appId});

/** @deprecated legacy wrapper, removed in F13 */
// eslint-disable-next-line max-params
export const uploadBundle = (file: File, projectId: string, appId: number, onProgress: (sent: number) => void): Promise<IAppBinaryData> =>
    legacy.upload({projectId, appId}, file, {onProgress});
