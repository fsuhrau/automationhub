import {useEffect, useState} from 'react';
import {appendNew, Identified} from '../lib/appendNew';

// the part of EventSource the hook uses, so tests can fake it
export type SseSource = Pick<EventSource, 'addEventListener' | 'close'>;
export type SseSourceFactory = (channel: string) => SseSource;

export interface SseLogOptions {
    // name of the event that carries the entries (JSON: one entry or a list)
    event?: string,
    // keep only the last `limit` entries
    limit?: number,
    createSource?: SseSourceFactory,
}

export interface SseLog<T> {
    entries: T[],
    connected: boolean,
}

const defaultSource: SseSourceFactory = channel => new EventSource(channel, {withCredentials: true});

const parseEntries = <T>(e: Event): T[] => {
    const data: T | T[] = JSON.parse((e as MessageEvent<string>).data);
    return Array.isArray(data) ? data : [data];
};

interface LogState<T> extends SseLog<T> {
    channel: string,
}

const emptyState = <T>(channel: string): LogState<T> => ({channel, entries: [], connected: false});

// an update that belongs to another channel than the state restarts it
const apply = <T>(prev: LogState<T>, channel: string, change: Partial<SseLog<T>>): LogState<T> =>
    ({...(prev.channel === channel ? prev : emptyState<T>(channel)), ...change});

// useSseLog collects the entries an SSE channel (an URL) sends. Entries
// whose id is already known are ignored; the list restarts when the channel
// changes and the connection closes on unmount.
export function useSseLog<T extends Identified>(channel: string, options: SseLogOptions = {}): SseLog<T> {
    const {event = 'message', limit, createSource = defaultSource} = options;
    const [state, setState] = useState<LogState<T>>(() => emptyState<T>(channel));

    useEffect(() => {
        const set = (change: Partial<SseLog<T>>): void => setState(prev => apply(prev, channel, change));
        const source = createSource(channel);
        source.addEventListener('open', () => set({connected: true}));
        source.addEventListener('error', () => set({connected: false}));
        source.addEventListener(event, e => setState(prev => {
            const {entries} = prev.channel === channel ? prev : emptyState<T>(channel);
            return apply(prev, channel, {entries: appendNew(entries, parseEntries<T>(e), limit)});
        }));
        return () => source.close();
    }, [channel, event, limit, createSource]);

    return state.channel === channel ? state : emptyState<T>(channel);
}
