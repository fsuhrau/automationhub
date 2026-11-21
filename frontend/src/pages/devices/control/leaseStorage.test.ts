import {afterEach, describe, expect, it, vi} from 'vitest';
import {loadLease, storeLease} from './leaseStorage';

afterEach(() => {
    sessionStorage.clear();
    vi.restoreAllMocks();
});

describe('leaseStorage', () => {
    it('keeps a token per device', () => {
        storeLease(1, 'a');
        storeLease(2, 'b');
        expect([loadLease(1), loadLease(2), loadLease(3)]).toEqual(['a', 'b', null]);
    });

    it('forgets a token', () => {
        storeLease(1, 'a');
        storeLease(1, null);
        expect(loadLease(1)).toBeNull();
    });

    it('lives without storage', () => {
        vi.spyOn(Storage.prototype, 'getItem').mockImplementation(() => {
            throw new Error('denied');
        });
        vi.spyOn(Storage.prototype, 'setItem').mockImplementation(() => {
            throw new Error('denied');
        });
        expect(loadLease(1)).toBeNull();
        expect(() => storeLease(1, 'a')).not.toThrow();
    });
});
