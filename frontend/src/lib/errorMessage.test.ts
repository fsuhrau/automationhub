import {describe, expect, it} from 'vitest';
import {errorMessage} from './errorMessage';

describe('errorMessage', () => {
    it.each([
        ['server message', {response: {data: {message: 'denied'}}, message: 'Request failed'}, 'denied'],
        ['unparsed JSON body', {response: {data: '{"message":"no scene"}'}, message: 'x'}, 'no scene'],
        ['text body without message', {response: {data: 'plain'}, message: 'Request failed'}, 'Request failed'],
        ['empty server message', {response: {data: {message: ''}}, message: 'Request failed'}, 'Request failed'],
        ['null body', {response: {data: null}, message: 'boom'}, 'boom'],
        ['Error', new Error('broken'), 'broken'],
        ['empty Error message', new Error(''), 'Error'],
        ['string', 'plain text', 'plain text'],
        ['number', 42, '42'],
        ['null', null, 'null'],
        ['undefined', undefined, 'undefined'],
    ])('%s', (_name, input, expected) => {
        expect(errorMessage(input)).toBe(expected);
    });
});
