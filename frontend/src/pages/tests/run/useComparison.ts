import {useLoader} from '../../../hooks/useLoader';
import {useTestRunService} from '../../../services/test.run.service';
import {IBaseline} from '../../../types/protocol.baseline';
import ITestProtocolData from '../../../types/test.protocol';
import {ProtocolRef} from '../../../types/test.run.refs';
import {Comparison, comparisonOf, pickedBaseline} from './protocolViewModel';

export interface ComparisonState {
    comparison: Comparison | null,
    // the run picked to compare with, once loaded
    baseline: IBaseline | null,
    loading: boolean,
    // the picked run could not be loaded
    failed: boolean,
}

export interface ComparisonQuery {
    vs: string | null,
    vsDevice: string | null,
}

// useComparison finds the run the protocol's performance is compared with:
// the picked one, loaded from the master, or the latest earlier one. It is
// chosen by run so it stays while moving through the tests.
export const useComparison = (ref: ProtocolRef, earlier: ITestProtocolData | null, {vs, vsDevice}: ComparisonQuery): ComparisonState => {
    const service = useTestRunService();
    const {projectId, appId, testId, runId, protocolId} = ref;
    const {data, error, loading} = useLoader(
        () => vs ? service.baseline({projectId, appId, testId, runId, protocolId}, vs, vsDevice) : null,
        [service, projectId, appId, testId, runId, protocolId, vs, vsDevice]);
    const baseline = pickedBaseline(vs, data);
    return {comparison: comparisonOf(vs, baseline, earlier), baseline, loading, failed: !!vs && !!error};
};
