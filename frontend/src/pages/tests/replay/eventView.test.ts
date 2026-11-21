import {expect, it} from 'vitest';
import {TestResultState} from '../../../types/test.result.state.enum';
import {ReplayEvent} from '../replayModel';
import {eventText, isMarker, markerColor, nextMarkAfter, resultTone, scrollTargetOf, shownBy} from './eventView';

const ev = (over: Partial<ReplayEvent>): ReplayEvent => ({key: 'k', time: 0, kind: 'log', level: 'info', text: 'x', ...over});

const log = ev({});
const error = ev({level: 'error'});
const step = ev({kind: 'step'});

it.each([
    ['all', [true, true, true]],
    ['tests', [false, true, true]],
    ['errors', [false, true, false]],
] as const)('%s', (filter, expected) => {
    expect([log, error, step].map(shownBy(filter))).toEqual(expected);
});

it.each([
    [TestResultState.TestResultFailed, 'error', 'failed'],
    [TestResultState.TestResultUnstable, 'warning', 'unstable'],
    [TestResultState.TestResultSuccess, 'success', 'passed'],
    [TestResultState.TestResultOpen, 'success', 'passed'],
    [undefined, 'success', 'passed'],
] as const)('result %s is %s', (result, tone, word) => {
    expect(resultTone(result)).toBe(tone);
    expect(eventText(ev({kind: 'test-end', result, text: 'T'}))).toBe(`Test ${word}: T`);
});

it('result wording and tones: words a test start and leaves other texts', () => {
    expect(eventText(ev({kind: 'test-start', text: 'T'}))).toBe('Test started: T');
    expect(eventText(ev({text: 'plain'}))).toBe('plain');
});

it('markers: marks test ends and errors, colored by result', () => {
    const end = ev({kind: 'test-end', result: TestResultState.TestResultUnstable});
    expect([end, ev({level: 'error'}), ev({})].map(isMarker)).toEqual([true, true, false]);
    expect(markerColor(end)).toBe('warning.main');
    expect(markerColor(ev({level: 'error'}))).toBe('error.main');
});

it('markers: finds the next test start or error beyond a small gap', () => {
    const events = [ev({key: 'a', time: 100, kind: 'test-start'}), ev({key: 'b', time: 120, level: 'error'}), ev({key: 'c', time: 200, kind: 'step'})];
    expect(nextMarkAfter(events, 0)?.key).toBe('a');
    expect(nextMarkAfter(events, 60)?.key).toBe('b');
    expect(nextMarkAfter(events, 80)).toBeUndefined();
});

const box = {rowHeight: 20, scrollTop: 100, height: 300};

it('scrollTargetOf: is null while the row is in view', () => {
    expect(scrollTargetOf({...box, rowTop: 150})).toBeNull();
});

it('scrollTargetOf: puts a hidden row a third down the list', () => {
    expect(scrollTargetOf({...box, rowTop: 50})).toBe(0);
    expect(scrollTargetOf({...box, rowTop: 900})).toBe(800);
    expect(scrollTargetOf({...box, rowTop: 395})).toBe(295);
});
