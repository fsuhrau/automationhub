import {useEffect, useState} from 'react';
import {useTestRunService} from '../../../services/test.run.service';
import {RunRef} from '../../../types/test.run.refs';

const POLL_MS = 3000;

// useHeldDevices tells which devices a running run holds. A run that just
// started may not hold its devices yet, so it asks again until it does.
// `initial` is undefined on older masters, which don't say.
export const useHeldDevices = (initial: number[] | undefined, running: boolean, run: RunRef): number[] | undefined => {
    const service = useTestRunService();
    const [held, setHeld] = useState<number[] | undefined>(initial);
    const {projectId, appId, testId, runId} = run;
    const waiting = running && initial !== undefined && !(held && held.length > 0);
    useEffect(() => {
        if (!waiting) return;
        const timer = setInterval(() => {
            service.run({projectId, appId, testId, runId}).then(r => setHeld(r.devices ?? [])).catch(() => undefined);
        }, POLL_MS);
        return () => clearInterval(timer);
    }, [waiting, service, projectId, appId, testId, runId]);
    return held;
};
