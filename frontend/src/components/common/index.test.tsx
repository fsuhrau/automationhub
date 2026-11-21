import {describe, expect, it} from 'vitest';
import {Code, ConfirmDialog, CopyButton, Section} from './index';
import * as parts from '../../pages/settings/parts';

describe('shared kit exports', () => {
    it('re-exports the shared kit', () => {
        expect(parts.Section).toBe(Section);
        expect(parts.Code).toBe(Code);
        expect(parts.CopyButton).toBe(CopyButton);
        expect(parts.ConfirmDialog).toBe(ConfirmDialog);
        expect(parts.messageOf(new Error('m'))).toBe('m');
    });
});
