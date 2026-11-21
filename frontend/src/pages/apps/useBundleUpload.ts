import {useCallback, useEffect, useRef} from 'react';
import {useAppService} from '../../services/app.service';
import {IAppBinaryData} from '../../types/app';
import {AppRef} from '../../types/app.ref';
import {BundleUpload} from '../../types/bundle.upload';
import {Accept, accepts} from './bundlesModel';
import {CANCELLED, failureState, isActive, rejectedState, sendingState} from './uploadModel';
import {useUploadList} from './useUploadList';

export interface BundleUploadOptions {
    // null until the app is known; nothing is uploaded then
    app: AppRef | null,
    accept: Accept,
    appName?: string,
    onUploaded: (bundle: IAppBinaryData) => void,
}

export interface BundleUploads {
    uploads: BundleUpload[],
    // true while a file is on its way
    busy: boolean,
    upload: (files: File[]) => void,
    cancel: (id: number) => void,
    dismiss: (id: number) => void,
    clear: () => void,
}

interface Job {
    id: number,
    file: File,
    signal: AbortSignal,
}

// useBundleUpload uploads files one after the other (the master reads each
// one) and reports progress, the result or the failure per file. A running
// or waiting upload can be cancelled; leaving the page cancels all of them.
export const useBundleUpload = ({app, accept, appName, onUploaded}: BundleUploadOptions): BundleUploads => {
    const service = useAppService();
    const list = useUploadList();
    const queue = useRef<Promise<void>>(Promise.resolve());
    const aborts = useRef(new Map<number, AbortController>());
    useEffect(() => {
        const pending = aborts.current;
        return () => pending.forEach(controller => controller.abort());
    }, []);

    const send = async (job: Job, target: AppRef): Promise<void> => {
        if (job.signal.aborted) return list.patch(job.id, CANCELLED);
        try {
            const bundle = await service.upload(target, job.file, {
                signal: job.signal,
                onProgress: sent => list.patch(job.id, sendingState(sent)),
            });
            list.patch(job.id, {phase: 'done', bundle});
            onUploaded(bundle);
        } catch (e) {
            list.patch(job.id, failureState(e, job.signal.aborted));
        }
    };

    const enqueue = (file: File, target: AppRef): void => {
        const controller = new AbortController();
        const id = list.add(file, {phase: 'sending', sent: 0});
        aborts.current.set(id, controller);
        queue.current = queue.current.then(() => send({id, file, signal: controller.signal}, target));
    };

    const addFile = (file: File): void => {
        if (!app) return;
        if (accepts(accept, file.name)) enqueue(file, app);
        else list.add(file, rejectedState(appName, accept));
    };

    const cancel = useCallback((id: number): void => aborts.current.get(id)?.abort(), []);
    return {
        uploads: list.uploads, busy: list.uploads.some(u => isActive(u.state)),
        upload: files => files.forEach(addFile), cancel, dismiss: list.dismiss, clear: list.clear,
    };
};
