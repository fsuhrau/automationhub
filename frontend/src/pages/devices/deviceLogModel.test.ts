import {describe, expect, it} from 'vitest';
import {
    addLines, chipHint, clearLines, dropped, emptyLogText, failed, initialLog, logChip, opened,
} from './deviceLogModel';

const start = initialLog('url');

describe('device log state', () => {
    it('starts connecting without lines', () => {
        expect(start).toEqual({channel: 'url', lines: [], status: 'connecting', failure: null});
    });

    it('goes live with new lines and forgets a failure', () => {
        const live = addLines(failed(start, 'broken'), ['a', 'b']);
        expect(live).toMatchObject({lines: ['a', 'b'], status: 'live', failure: null});
    });

    it('keeps the last lines up to the limit', () => {
        expect(addLines(addLines(start, ['a', 'b']), ['c'], 2).lines).toEqual(['b', 'c']);
    });

    it('fails with a message', () => {
        expect(failed(start, 'no adb')).toMatchObject({status: 'failing', failure: 'no adb'});
    });

    it('moves between connecting and live on open and drop, and leaves other states', () => {
        const live = opened(start);
        expect(live.status).toBe('live');
        expect(dropped(live).status).toBe('connecting');
        const failing = failed(start, 'x');
        expect(opened(failing)).toBe(failing);
        expect(dropped(failing)).toBe(failing);
        expect(dropped(start)).toBe(start);
        expect(opened(live)).toBe(live);
    });

    it('clears the lines', () => {
        expect(clearLines(addLines(start, ['a'])).lines).toEqual([]);
    });
});

describe('logChip', () => {
    it.each([
        ['connecting', 'Connecting…', 'default'],
        ['live', 'Live', 'success'],
        ['failing', 'No log', 'error'],
    ] as const)('shows %s', (status, label, color) => {
        expect(logChip({...start, status}, null)).toEqual({label, color});
    });

    it('counts the lines that came while paused', () => {
        const state = addLines(start, ['a', 'b', 'c']);
        expect(logChip(state, ['a'])).toEqual({label: 'Paused · 2 new', color: 'default'});
    });
});

describe('texts', () => {
    it('hints at the failure, or at what live means', () => {
        expect(chipHint(failed(start, 'boom'))).toBe('boom');
        expect(chipHint(opened(start))).toBe('New log lines appear as the device writes them');
        expect(chipHint(start)).toBe('');
    });

    it('says why the log area is empty', () => {
        expect(emptyLogText(0, 'connecting')).toBe('Waiting for log lines…');
        expect(emptyLogText(0, 'failing')).toBe('The device log could not be read.');
        expect(emptyLogText(5, 'live')).toBe('No line matches.');
    });
});
