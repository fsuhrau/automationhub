import {useEffect, useState} from 'react';
import {useNavigate} from 'react-router-dom';
import {useError} from '../../../ErrorProvider';
import {useAsyncAction} from '../../../hooks/useAsyncAction';
import {routes} from '../../../lib/routes';
import {useTestRunService} from '../../../services/test.run.service';
import ITestRunData from '../../../types/test.run';
import {TestRef} from '../../../types/test.run.refs';

export interface RunActions {
    rerun: () => void,
    cancel: () => void,
    // the cancel request was sent; the run ends on its own after that
    cancelling: boolean,
}

// shows the failure of an action in the page's error banner
const useReportError = (error: string | null): void => {
    const {setError} = useError();
    useEffect(() => {
        if (error) setError(error);
    }, [error, setError]);
};

// useRunActions starts the same test again and cancels the run
export const useRunActions = (run: ITestRunData, test: TestRef): RunActions => {
    const service = useTestRunService();
    const navigate = useNavigate();
    const [cancelled, setCancelled] = useState(false);
    const {projectId, appId, testId} = test;
    const again = useAsyncAction(async () => {
        const started = await service.execute(test, {appBinaryId: run.appBinaryId, params: run.parameter, startUrl: run.startUrl});
        navigate(routes.project(projectId).app(appId).test(testId).run(started.id as number).show());
    });
    const stop = useAsyncAction(async () => {
        await service.cancel({...test, runId: run.id as number});
        setCancelled(true);
    });
    useReportError(again.error);
    useReportError(stop.error);
    return {rerun: () => void again.run(), cancel: () => void stop.run(), cancelling: stop.busy || cancelled};
};
