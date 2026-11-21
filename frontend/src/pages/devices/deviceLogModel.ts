// The state of a live device log: the lines it kept and whether the stream
// works. Every change is a pure function so the hook only wires events to it.

// the most lines kept in the browser
export const MAX_LINES = 3000;

export type StreamState = 'connecting' | 'live' | 'failing';

export interface LogState {
    // the stream the state belongs to
    channel: string,
    lines: string[],
    status: StreamState,
    failure: string | null,
}

export const initialLog = (channel: string): LogState => ({channel, lines: [], status: 'connecting', failure: null});

export const addLines = (state: LogState, fresh: string[], limit: number = MAX_LINES): LogState =>
    ({...state, status: 'live', failure: null, lines: state.lines.concat(fresh).slice(-limit)});

export const failed = (state: LogState, message: string): LogState => ({...state, status: 'failing', failure: message});

// the browser reconnects by itself after a dropped connection
export const dropped = (state: LogState): LogState => state.status === 'live' ? {...state, status: 'connecting'} : state;

export const opened = (state: LogState): LogState => state.status === 'connecting' ? {...state, status: 'live'} : state;

export const clearLines = (state: LogState): LogState => ({...state, lines: []});

export interface LogChip {
    label: string,
    color: 'default' | 'success' | 'error',
}

const CHIPS: Record<StreamState, LogChip> = {
    live: {label: 'Live', color: 'success'},
    failing: {label: 'No log', color: 'error'},
    connecting: {label: 'Connecting…', color: 'default'},
};

// paused: the shown lines stay (frozen), new ones are still collected
export const logChip = (state: LogState, frozen: string[] | null): LogChip =>
    frozen ? {label: `Paused · ${state.lines.length - frozen.length} new`, color: 'default'} : CHIPS[state.status];

export const chipHint = (state: LogState): string =>
    state.failure ?? (state.status === 'live' ? 'New log lines appear as the device writes them' : '');

// what the empty log area says
export const emptyLogText = (shownCount: number, status: StreamState): string => {
    if (shownCount > 0) return 'No line matches.';
    return status === 'failing' ? 'The device log could not be read.' : 'Waiting for log lines…';
};
