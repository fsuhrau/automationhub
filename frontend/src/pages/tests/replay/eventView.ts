// How the events of a replay are told: filters, wording, tones and the
// marks the controls jump to. No React.
import {TestResultState} from '../../../types/test.result.state.enum';
import {eventIsError, ReplayEvent} from '../replayModel';

export type EventFilter = 'all' | 'tests' | 'errors';

export const FILTERS: [EventFilter, string][] = [['all', 'Everything'], ['tests', 'Tests and steps'], ['errors', 'Errors']];

const FILTER_RULES: Record<EventFilter, (e: ReplayEvent) => boolean> = {
    all: () => true,
    tests: e => e.kind !== 'log' || eventIsError(e),
    errors: eventIsError,
};

export const shownBy = (filter: EventFilter): ((e: ReplayEvent) => boolean) => FILTER_RULES[filter];

export type Tone = 'error' | 'warning' | 'success';

// the tone of a finished test; an open one counts as passed
export const resultTone = (result: TestResultState | undefined): Tone => {
    if (result === TestResultState.TestResultFailed) return 'error';
    return result === TestResultState.TestResultUnstable ? 'warning' : 'success';
};

const RESULT_WORD: Record<Tone, string> = {error: 'failed', warning: 'unstable', success: 'passed'};

export const eventText = (e: ReplayEvent): string => {
    if (e.kind === 'test-start') return `Test started: ${e.text}`;
    if (e.kind === 'test-end') return `Test ${RESULT_WORD[resultTone(e.result)]}: ${e.text}`;
    return e.text;
};

// marks on the timeline: where tests ended and errors happened
export const isMarker = (e: ReplayEvent): boolean => e.kind === 'test-end' || eventIsError(e);

export const markerColor = (e: ReplayEvent): string => `${e.kind === 'test-end' ? resultTone(e.result) : 'error'}.main`;

const NEXT_MARK_GAP_MS = 50;

// nextMarkAfter is the next test start or error after the position
export const nextMarkAfter = (events: ReplayEvent[], position: number): ReplayEvent | undefined =>
    events.find(e => e.time > position + NEXT_MARK_GAP_MS && (e.kind === 'test-start' || eventIsError(e)));

// scrollTargetOf is where a list has to scroll to bring a row into view, or
// null when it is visible already
export interface RowBox {
    rowTop: number,
    rowHeight: number,
    scrollTop: number,
    height: number,
}

export const scrollTargetOf = (b: RowBox): number | null => {
    const hidden = b.rowTop < b.scrollTop || b.rowTop > b.scrollTop + b.height - b.rowHeight;
    return hidden ? Math.max(0, b.rowTop - b.height / 3) : null;
};
