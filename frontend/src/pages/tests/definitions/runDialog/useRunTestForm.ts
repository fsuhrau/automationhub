import {useState} from 'react';
import {useNavigate} from 'react-router-dom';
import {useAsyncAction} from '../../../../hooks/useAsyncAction';
import {routes} from '../../../../lib/routes';
import {IAppData} from '../../../../types/app';
import {PlatformType} from '../../../../types/platform.type.enum';
import {useTestService} from '../useTestService';
import {defaultParams, paramsString, readyToStart} from './runTestModel';
import {RunBundles, useRunBundles} from './useRunBundles';

interface RunTarget {
    projectIdentifier: string,
    app: IAppData,
    test: { id: number },
}

export interface RunTestForm {
    needsBundle: boolean,
    needsUrl: boolean,
    bundles: RunBundles,
    params: Record<string, string>,
    setParam: (name: string, value: string) => void,
    startUrl: string,
    setStartUrl: (url: string) => void,
    ready: boolean,
    starting: boolean,
    error: string | null,
    start: () => void,
}

// useRunTestForm is the state of the run dialog: the bundle to install, the
// app's parameters and, for web apps, the start URL. start runs the test and
// opens the new run.
export const useRunTestForm = ({projectIdentifier, app, test}: RunTarget, onClose: () => void): RunTestForm => {
    const navigate = useNavigate();
    const service = useTestService();
    const needsBundle = app.platform !== PlatformType.Editor && app.platform !== PlatformType.Web;
    const needsUrl = app.platform === PlatformType.Web;
    const bundles = useRunBundles(projectIdentifier, app, needsBundle);
    const [params, setParams] = useState(() => defaultParams(app.parameter));
    const [startUrl, setStartUrl] = useState('');

    const {run, busy, error} = useAsyncAction(async (): Promise<void> => {
        const started = await service.execute({projectId: projectIdentifier, appId: app.id}, test.id, {
            appBinaryId: bundles.bundleId === '' ? null : bundles.bundleId,
            params: paramsString(params),
            startUrl: needsUrl ? startUrl : '',
        });
        onClose();
        navigate(routes.project(projectIdentifier).app(app.id).test(test.id).run(String(started.id)).show());
    });
    const ready = readyToStart({needsBundle, bundleId: bundles.bundleId, needsUrl, startUrl, uploading: bundles.progress !== null, starting: busy});
    return {
        needsBundle, needsUrl, bundles, params, startUrl, setStartUrl, ready, starting: busy, error: error ?? bundles.error,
        setParam: (name, value) => setParams(v => ({...v, [name]: value})),
        start: () => void run(),
    };
};
