import {useMemo, useState} from 'react';
import {useAsyncAction} from '../../../../hooks/useAsyncAction';
import {useLoader} from '../../../../hooks/useLoader';
import {errorMessage} from '../../../../lib/errorMessage';
import {getAppBundles, uploadBundle, usableBundles} from '../../../../services/app.service';
import {IAppBinaryData, IAppData} from '../../../../types/app';
import {acceptFor, accepts} from '../../../apps/bundlesModel';
import {uploadFailure, withUploads} from './runTestModel';

export interface RunBundles {
    // null while loading
    bundles: IAppBinaryData[] | null,
    bundleId: number | '',
    choose: (id: number | '') => void,
    // 0..1 while the file is sent, then 1 while the master reads it; null when idle
    progress: number | null,
    upload: (file: File) => void,
    error: string | null,
}

// useRunBundles lists the bundles a run can install, the newest chosen by
// default, and uploads a new one. Without needsBundle nothing is loaded.
export const useRunBundles = (projectId: string, app: IAppData, needsBundle: boolean): RunBundles => {
    const accept = acceptFor(app.platform);
    const {data, error: loadError} = useLoader(
        () => needsBundle ? getAppBundles(projectId, app.id) : null, [projectId, app.id, needsBundle]);
    const [uploads, setUploads] = useState<IAppBinaryData[]>([]);
    const [chosen, setChosen] = useState<number | '' | null>(null);
    const [progress, setProgress] = useState<number | null>(null);

    const bundles = useMemo(() => data && withUploads(usableBundles(data), uploads), [data, uploads]);
    const send = async (file: File): Promise<void> => {
        if (!accepts(accept, file.name)) throw new Error(`${app.name} takes ${accept.label} files.`);
        setProgress(0);
        try {
            const bundle = await uploadBundle(file, projectId, app.id, setProgress);
            setUploads(list => [bundle, ...list]);
            setChosen(bundle.id);
        } catch (e) {
            throw uploadFailure(e);
        } finally {
            setProgress(null);
        }
    };
    const {run, error: uploadError} = useAsyncAction(send);
    const error = uploadError ?? (loadError ? errorMessage(loadError) : null);
    return {bundles, bundleId: chosen ?? bundles?.[0]?.id ?? '', choose: setChosen, progress, upload: file => void run(file), error};
};
