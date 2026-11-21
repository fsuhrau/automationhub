import {describe, expect, it} from 'vitest';
import {TestResultState as S} from '../types/test.result.state.enum';
import {getTestStatusChipColor, getTestStatusColor, getTestStatusText} from './TestStatusHelper';

const UNKNOWN = 99 as S;

describe('TestStatusHelper', () => {
    it.each([
        {state: S.TestResultSuccess, text: 'Success', color: 'green', chip: 'success'},
        {state: S.TestResultFailed, text: 'Failed', color: 'red', chip: 'error'},
        {state: S.TestResultOpen, text: 'Open', color: 'gray', chip: 'default'},
        {state: S.TestResultUnstable, text: 'Unstable', color: 'yellow', chip: 'warning'},
    ])('names state $state with text, color and chip color', ({state, text, color, chip}) => {
        expect(getTestStatusText(state)).toBe(text);
        expect(getTestStatusColor(state)).toBe(color);
        expect(getTestStatusChipColor(state)).toBe(chip);
    });

    it('falls back for an unknown state', () => {
        expect(getTestStatusText(UNKNOWN)).toBe('');
        expect(getTestStatusColor(UNKNOWN)).toBe('green');
        expect(getTestStatusChipColor(UNKNOWN)).toBe('default');
    });
});
