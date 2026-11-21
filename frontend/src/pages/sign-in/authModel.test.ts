import {describe, expect, it} from 'vitest';
import {hasErrors, isEmail, returnPath, signInErrors, signUpErrors} from './authModel';

describe('isEmail', () => {
    it.each([['a@b.co', true], ['a@b', false], ['no at', false], ['', false]])('%s is %s', (value, expected) => {
        expect(isEmail(value)).toBe(expected);
    });
});

describe('returnPath', () => {
    it('is the page that asked for the sign in, else the fallback', () => {
        expect(returnPath({from: '/secret'}, '/')).toBe('/secret');
        expect(returnPath({}, '/')).toBe('/');
        expect(returnPath(null, '/')).toBe('/');
    });
});

describe('field errors', () => {
    it('are empty for good sign in input', () => {
        expect(hasErrors(signInErrors({email: 'a@b.co', password: 'x'}))).toBe(false);
    });

    it('name every bad sign in field', () => {
        const errors = signInErrors({email: 'x', password: ''});
        expect(errors.email).toContain('email address');
        expect(errors.password).toContain('password');
        expect(hasErrors(errors)).toBe(true);
    });

    it('are empty for good sign up input', () => {
        expect(hasErrors(signUpErrors({name: 'Ada', email: 'a@b.co', password: '12345678'}))).toBe(false);
    });

    it('name every bad sign up field', () => {
        const errors = signUpErrors({name: '  ', email: 'x', password: '1234567'});
        expect(errors.name).toContain('name');
        expect(errors.email).toContain('email');
        expect(errors.password).toBe('At least 8 characters.');
    });
});
