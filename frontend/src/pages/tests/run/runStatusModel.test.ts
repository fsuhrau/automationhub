import {describe, expect, it} from 'vitest';
import {TestResultState as S} from '../../../types/test.result.state.enum';
import {device} from './testkit/fixtures';
import {deviceLabel, RESULT_ORDER, resultLabel, resultPalette, STATUS} from './runStatusModel';

describe('resultLabel', () => {
    it.each([
        [S.TestResultFailed, false, 'Failed'],
        [S.TestResultUnstable, false, 'Unstable'],
        [S.TestResultSuccess, false, 'Passed'],
        [S.TestResultOpen, true, 'Running'],
        [S.TestResultOpen, false, 'Not finished'],
    ])('words %s (running %s) as %s', (result, running, label) => {
        expect(resultLabel(result, running)).toBe(label);
    });
});

describe('resultPalette', () => {
    it.each([[S.TestResultFailed, 'error'], [S.TestResultUnstable, 'warning'], [S.TestResultSuccess, 'success'], [S.TestResultOpen, 'info']])(
        'colors %s as %s', (result, color) => {
            expect(resultPalette(result)).toBe(color);
        });
});

describe('deviceLabel', () => {
    it('prefers the alias, else the name, else nothing', () => {
        expect(deviceLabel(device({alias: 'Lab 1'}))).toBe('Lab 1');
        expect(deviceLabel(device({alias: ''}))).toBe('iPhone 12');
        expect(deviceLabel(null)).toBe('');
        expect(deviceLabel(undefined)).toBe('');
    });
});

describe('the status table', () => {
    it('words every run status and orders every result', () => {
        expect(Object.keys(STATUS)).toHaveLength(7);
        expect(RESULT_ORDER).toHaveLength(4);
    });
});
