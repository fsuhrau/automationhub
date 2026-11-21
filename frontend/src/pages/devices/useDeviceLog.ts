import {useCallback, useEffect, useState} from 'react';
import {SseSourceFactory} from '../../hooks/useSseLog';
import {addLines, clearLines, dropped, failed, initialLog, LogState, opened} from './deviceLogModel';

export interface DeviceLog extends LogState {
    clear: () => void,
}

const defaultSource: SseSourceFactory = url => new EventSource(url, {withCredentials: true});

const dataOf = (e: Event): string => (e as MessageEvent<string>).data;

// useDeviceLog follows the log stream at url. The device sends the events
// "lines" (a list of lines) and "failure" ({message}) without ids, so
// useSseLog (which dedupes by id) doesn't fit. The state restarts when the
// url changes and the connection closes on unmount.
export function useDeviceLog(url: string, createSource: SseSourceFactory = defaultSource): DeviceLog {
    const [state, setState] = useState<LogState>(() => initialLog(url));

    useEffect(() => {
        const update = (change: (s: LogState) => LogState): void =>
            setState(prev => change(prev.channel === url ? prev : initialLog(url)));
        const source = createSource(url);
        source.addEventListener('open', () => update(opened));
        source.addEventListener('error', () => update(dropped));
        source.addEventListener('lines', e => {
            const fresh: string[] = JSON.parse(dataOf(e));
            update(s => addLines(s, fresh));
        });
        source.addEventListener('failure', e => {
            const {message}: { message: string } = JSON.parse(dataOf(e));
            update(s => failed(s, message));
        });
        return () => source.close();
    }, [url, createSource]);

    const clear = useCallback(() => setState(clearLines), []);
    return {...(state.channel === url ? state : initialLog(url)), clear};
}
