import {describe, expect, it} from 'vitest';
import {at, entry, protocol} from './testkit/fixtures';
import {
    filterLog, formatRuntime, isError, isScreen, isStep, logAsText, mergeRunnerLog, sourceName, splitMessage, summarize,
} from './logModel';

describe('formatRuntime', () => {
    it.each([
        [undefined, '0:00.000'],
        [null, '0:00.000'],
        [2.345, '0:02.345'],
        [-2.345, '-0:02.345'],
        [-0.0001, '0:00.000'],
        [75.5, '1:15.500'],
        [3725.9, '1:02:05'],
    ])('shows %s as %s', (seconds, text) => {
        expect(formatRuntime(seconds)).toBe(text);
    });
});

describe('entry kinds', () => {
    it('knows errors, steps and screens', () => {
        expect(isError(entry({level: 'error'}))).toBe(true);
        expect(isError(entry())).toBe(false);
        expect(isStep(entry())).toBe(true);
        expect(isScreen(entry({source: 'screen', data: 'a.png'}))).toBe(true);
        expect(isScreen(entry({source: 'screen', data: ''}))).toBe(false);
    });

    it('names the runner for entries without source', () => {
        expect(sourceName('')).toBe('runner');
        expect(sourceName('unity')).toBe('unity');
    });
});

describe('mergeRunnerLog', () => {
    const test = protocol({id: 2, startedAt: at(5)});
    const own = [entry({id: 1, runtime: 1}), entry({id: 2, runtime: 3})];

    it('is the test log without a runner or without runner entries', () => {
        expect(mergeRunnerLog(test, own, null)).toBe(own);
        expect(mergeRunnerLog(test, own, protocol({id: 1, entries: []}))).toBe(own);
    });

    it('puts the runner entries on the test clock as setup and teardown', () => {
        const runner = protocol({id: 1, startedAt: at(0), entries: [entry({id: 10, runtime: 2}), entry({id: 11, runtime: 8.5})]});
        const merged = mergeRunnerLog(test, own, runner);
        expect(merged.map(e => [e.id, e.runtime, e.phase])).toEqual([[10, -3, 'setup'], [1, 1, undefined], [2, 3, undefined], [11, 3.5, 'teardown']]);
    });

    it('lists the test entries first on equal times', () => {
        const runner = protocol({id: 1, startedAt: at(5), entries: [entry({id: 10, runtime: 1})]});
        expect(mergeRunnerLog(test, own, runner).map(e => e.id)).toEqual([1, 10, 2]);
    });

    it('does not shift without start times', () => {
        const runner = protocol({id: 1, startedAt: null as unknown as Date, entries: [entry({id: 10, runtime: 2})]});
        expect(mergeRunnerLog(test, [], runner)[0].runtime).toBe(2);
    });
});

describe('summarize', () => {
    it('groups errors, steps, screens and counts the sources', () => {
        const summary = summarize([
            entry({id: 1}), entry({id: 2, source: '', level: 'error'}), entry({id: 3, source: 'screen', data: 'x'}), entry({id: 4}),
        ]);
        expect(summary.errors.map(e => e.id)).toEqual([2]);
        expect(summary.steps.map(e => e.id)).toEqual([1, 4]);
        expect(summary.screens.map(e => e.id)).toEqual([3]);
        expect(summary.sources).toEqual([{source: 'step', count: 2}, {source: 'runner', count: 1}, {source: 'screen', count: 1}]);
    });
});

describe('filterLog', () => {
    const log = [
        {...entry({id: 1, message: 'Hello World'})},
        {...entry({id: 2, source: 'unity', level: 'error', message: 'boom'})},
        {...entry({id: 3, message: undefined as unknown as string}), phase: 'setup' as const},
    ];
    const filter = (over: Partial<Parameters<typeof filterLog>[1]> = {}): (number | null | undefined)[] =>
        filterLog(log, {hidden: new Set(), errorsOnly: false, text: '', ...over}).map(e => e.id);

    it('keeps everything without filters', () => {
        expect(filter()).toEqual([1, 2, 3]);
    });

    it('filters errors, hidden sources and phases, and text', () => {
        expect(filter({errorsOnly: true})).toEqual([2]);
        expect(filter({hidden: new Set(['unity'])})).toEqual([1, 3]);
        expect(filter({hidden: new Set(['setup'])})).toEqual([1, 2]);
        expect(filter({text: ' WORLD '})).toEqual([1]);
    });
});

describe('splitMessage', () => {
    it.each([
        ['one line  \n', {headline: 'one line', detail: ''}],
        ['head\nat a\nat b', {headline: 'head', detail: 'at a\nat b'}],
        [undefined as unknown as string, {headline: '', detail: ''}],
    ])('splits %j', (text, parts) => {
        expect(splitMessage(text)).toEqual(parts);
    });
});

describe('logAsText', () => {
    it('writes the log as tab separated lines', () => {
        const text = logAsText('Suite/Test', [
            {...entry({runtime: 1.5, message: 'a\nb', data: 'shot.png'})},
            {...entry({source: '', level: 'error', message: undefined as unknown as string, runtime: 2}), phase: 'teardown'},
        ]);
        expect(text).toBe('# Suite/Test\n0:01.500\tstep\tinfo\ta\n\t\t\tb\t[shot.png]\n0:02.000\tteardown/runner\terror\t\n');
    });
});
