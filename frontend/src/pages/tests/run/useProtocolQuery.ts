import {useSearchParams} from 'react-router-dom';
import {IBaselineCandidate} from '../../../types/protocol.baseline';
import {TabKey, tabFromQuery} from './protocolViewModel';

export interface ProtocolQuery {
    tab: TabKey,
    setTab: (tab: TabKey) => void,
    // the run picked to compare with, and its device when it is another one
    vs: string | null,
    vsDevice: string | null,
    // picks a run to compare with, null for the latest earlier run
    setBaseline: (candidate: IBaselineCandidate | null, protocolDeviceId: number | null | undefined) => void,
    // the query to keep when moving to another test of the run
    keep: () => string,
}

// useProtocolQuery keeps the tab and the comparison in the url
export const useProtocolQuery = (): ProtocolQuery => {
    const [query, setQuery] = useSearchParams();
    const tab = tabFromQuery(query.get('t'));
    const vs = query.get('vs');
    const vsDevice = query.get('vsDevice');
    const update = (change: (next: URLSearchParams) => void, replace = false): void => setQuery(q => {
        const next = new URLSearchParams(q);
        change(next);
        return next;
    }, {replace});
    return {
        tab, vs, vsDevice,
        setTab: t => update(next => t === 'overview' ? next.delete('t') : next.set('t', t), true),
        setBaseline: (c, deviceId) => update(next => {
            next.delete('vs');
            next.delete('vsDevice');
            if (!c) return;
            next.set('vs', String(c.testRunId));
            if (c.deviceId && c.deviceId !== deviceId) next.set('vsDevice', String(c.deviceId));
        }),
        keep: () => {
            const keep = new URLSearchParams();
            if (tab !== 'overview') keep.set('t', tab);
            if (vs) keep.set('vs', vs);
            if (vsDevice) keep.set('vsDevice', vsDevice);
            const q = keep.toString();
            return q ? `?${q}` : '';
        },
    };
};
