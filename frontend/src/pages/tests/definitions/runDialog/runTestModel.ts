// Starting a run without React: the form's parameters as the master takes
// them, when the form may be submitted, and the uploaded bundles.
import {AppParameter, IAppBinaryData} from '../../../../types/app';

export const CONFLICT_MESSAGE = 'This exact file was uploaded before; pick it in the list.';

// defaultParams has a value for every parameter of the app: its default
export const defaultParams = (parameters: AppParameter[]): Record<string, string> =>
    Object.fromEntries(parameters.map(p => [p.name, p.type.defaultValue ?? '']));

// paramsString is "name=value;name=value"; empty parameters aren't passed
export const paramsString = (params: Record<string, string>): string =>
    Object.entries(params).filter(([, v]) => v.length > 0).map(([k, v]) => `${k}=${v}`).join(';');

export interface ReadyState {
    needsBundle: boolean,
    bundleId: number | '',
    needsUrl: boolean,
    startUrl: string,
    // a bundle is being uploaded
    uploading: boolean,
    starting: boolean,
}

export const readyToStart = (s: ReadyState): boolean =>
    (!s.needsBundle || s.bundleId !== '') && (!s.needsUrl || s.startUrl.trim() !== '') && !s.uploading && !s.starting;

// uploadFailure is what to show for a failed upload: an identical file is a 409
export const uploadFailure = (e: unknown): unknown =>
    (e as { response?: { status?: number } }).response?.status === 409 ? new Error(CONFLICT_MESSAGE) : e;

// withUploads puts the uploaded bundles (newest first) before the listed ones
export const withUploads = (listed: IAppBinaryData[], uploads: IAppBinaryData[]): IAppBinaryData[] =>
    [...uploads, ...listed.filter(b => !uploads.some(u => u.id === b.id))];
