import {useState} from 'react';
import {useLoader} from '../../../hooks/useLoader';
import {useTestRunService} from '../../../services/test.run.service';
import {IBaselineList} from '../../../types/protocol.baseline';
import {ProtocolRef} from '../../../types/test.run.refs';
import {useDebouncedValue} from './useDebouncedValue';

export interface BaselineCandidates {
    list: IBaselineList | null,
    error: unknown,
    loading: boolean,
    retry: () => void,
    text: string,
    setText: (text: string) => void,
    // the search the list was loaded for
    q: string,
    tag: string | null,
    setTag: (tag: string | null) => void,
    anyDevice: boolean,
    setAnyDevice: (on: boolean) => void,
}

// useBaselineCandidates loads the runs a protocol can be compared with,
// searched by text (once typing pauses), tag and device
export const useBaselineCandidates = (open: boolean, ref: ProtocolRef): BaselineCandidates => {
    const service = useTestRunService();
    const [text, setText] = useState('');
    const [tag, setTag] = useState<string | null>(null);
    const [anyDevice, setAnyDevice] = useState(false);
    const q = useDebouncedValue(text.trim(), 300);
    const {projectId, appId, testId, runId, protocolId} = ref;
    const {data, error, loading, retry} = useLoader(
        () => open ? service.baselines({projectId, appId, testId, runId, protocolId}, {q, tag: tag ?? undefined, anyDevice}) : null,
        [service, open, projectId, appId, testId, runId, protocolId, q, tag, anyDevice]);
    return {list: data, error, loading, retry, text, setText, q, tag, setTag, anyDevice, setAnyDevice};
};
