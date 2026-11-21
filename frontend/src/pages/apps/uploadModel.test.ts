import {expect, it} from 'vitest';
import {bundle} from './appsFixtures';
import {CANCELLED, failureState, isActive, rejectedState, sendingState, uploadSeverity, uploadText} from './uploadModel';

it.each([
    [{phase: 'sending', sent: 0.256} as const, 'Uploading 26 %', 'info'],
    [{phase: 'reading'} as const, 'Reading the bundle on the master…', 'info'],
    [{phase: 'done', bundle: bundle(4)} as const, 'Uploaded as version 1.4.0 (#4)', 'success'],
    [{phase: 'done', bundle: bundle(4, {version: ''})} as const, 'Uploaded as version quest-4.apk (#4)', 'success'],
    [{phase: 'exists'} as const, 'This exact file was uploaded before; the list already has it.', 'warning'],
    [{phase: 'failed', message: 'boom'} as const, 'boom', 'error'],
    [CANCELLED, 'Upload cancelled.', 'warning'],
])('describes the state %j', (state, text, severity) => {
    expect(uploadText(state)).toBe(text);
    expect(uploadSeverity(state)).toBe(severity);
});

it('is active while sending or reading only', () => {
    expect(isActive({phase: 'sending', sent: 0})).toBe(true);
    expect(isActive({phase: 'reading'})).toBe(true);
    expect(isActive({phase: 'exists'})).toBe(false);
    expect(isActive(CANCELLED)).toBe(false);
});

it('reads the file once all of it is sent', () => {
    expect(sendingState(0.5)).toEqual({phase: 'sending', sent: 0.5});
    expect(sendingState(1)).toEqual({phase: 'reading'});
});

it('tells how an upload failed', () => {
    expect(failureState(new Error('x'), true)).toEqual(CANCELLED);
    expect(failureState({response: {status: 409}}, false)).toEqual({phase: 'exists'});
    expect(failureState({response: {status: 500, data: {message: 'bad'}}}, false)).toEqual({phase: 'failed', message: 'bad'});
    expect(failureState(null, false)).toEqual({phase: 'failed', message: 'null'});
});

it('rejects a file the app does not take, by the name of the app', () => {
    const accept = {extensions: ['.apk'], label: '.apk'};
    expect(rejectedState('Quest', accept)).toEqual({phase: 'failed', message: 'Quest takes .apk files.'});
    expect(rejectedState(undefined, accept)).toEqual({phase: 'failed', message: 'This app takes .apk files.'});
});
