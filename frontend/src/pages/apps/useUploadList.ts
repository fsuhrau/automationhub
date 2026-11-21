import {useCallback, useRef, useState} from 'react';
import {BundleUpload, UploadState} from '../../types/bundle.upload';

export interface UploadList {
    uploads: BundleUpload[],
    // returns the id of the new entry
    add: (file: File, state: UploadState) => number,
    patch: (id: number, state: UploadState) => void,
    dismiss: (id: number) => void,
    clear: () => void,
}

// useUploadList keeps the uploads a page shows, in the order they were added
export const useUploadList = (): UploadList => {
    const [uploads, setUploads] = useState<BundleUpload[]>([]);
    const next = useRef(0);
    const add = useCallback((file: File, state: UploadState): number => {
        const id = next.current++;
        setUploads(list => [...list, {id, file, state}]);
        return id;
    }, []);
    const patch = useCallback((id: number, state: UploadState): void => {
        setUploads(list => list.map(u => u.id === id ? {...u, state} : u));
    }, []);
    const dismiss = useCallback((id: number): void => setUploads(list => list.filter(u => u.id !== id)), []);
    const clear = useCallback((): void => setUploads([]), []);
    return {uploads, add, patch, dismiss, clear};
};
