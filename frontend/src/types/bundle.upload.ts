import {IAppBinaryData} from './app';

// Where an upload of a bundle file stands, from the first byte to the answer
export type UploadState =
    | { phase: 'sending', sent: number }
    | { phase: 'reading' }
    | { phase: 'done', bundle: IAppBinaryData }
    | { phase: 'exists' }
    | { phase: 'failed', message: string }
    | { phase: 'cancelled' };

export interface BundleUpload {
    id: number,
    file: File,
    state: UploadState,
}

export interface UploadOptions {
    // 0..1 while the file is sent
    onProgress?: (sent: number) => void,
    // aborts the request
    signal?: AbortSignal,
}
