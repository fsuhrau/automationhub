import {describe, expect, it} from 'vitest';
import {aBundle, aParameter} from '../harness/fixtures';
import {CONFLICT_MESSAGE, defaultParams, paramsString, readyToStart, ReadyState, uploadFailure, withUploads} from './runTestModel';

const ready = (over: Partial<ReadyState> = {}): ReadyState =>
    ({needsBundle: true, bundleId: 3, needsUrl: false, startUrl: '', uploading: false, starting: false, ...over});

describe('parameters', () => {
    it('defaultParams uses the defaults and empty text without one', () => {
        const params = defaultParams([
            aParameter({name: 'a', type: {type: 'string', defaultValue: 'x'}}),
            aParameter({id: 2, name: 'b', type: {type: 'string'} as never}),
        ]);
        expect(params).toEqual({a: 'x', b: ''});
    });

    it('paramsString joins the filled parameters', () => {
        expect(paramsString({a: '1', b: '', c: '3'})).toBe('a=1;c=3');
        expect(paramsString({})).toBe('');
    });
});

describe('readyToStart', () => {
    it.each([
        ['everything given', ready(), true],
        ['no bundle needed', ready({needsBundle: false, bundleId: ''}), true],
        ['no bundle chosen', ready({bundleId: ''}), false],
        ['url needed and empty', ready({needsUrl: true, startUrl: '  '}), false],
        ['url needed and given', ready({needsUrl: true, startUrl: 'https://x'}), true],
        ['uploading', ready({uploading: true}), false],
        ['starting', ready({starting: true}), false],
    ])('%s', (_, state, expected) => {
        expect(readyToStart(state)).toBe(expected);
    });
});

describe('uploads', () => {
    it('uploadFailure turns a 409 into the conflict message and keeps other errors', () => {
        expect(uploadFailure({response: {status: 409}})).toEqual(new Error(CONFLICT_MESSAGE));
        const other = {response: {status: 500}};
        expect(uploadFailure(other)).toBe(other);
    });

    it('withUploads puts the uploads first without doubles', () => {
        const listed = [aBundle({id: 1}), aBundle({id: 2})];
        expect(withUploads(listed, [aBundle({id: 2}), aBundle({id: 9})]).map(b => b.id)).toEqual([2, 9, 1]);
    });
});
