import {describe, expect, it} from 'vitest';
import {PlatformType} from '../../types/platform.type.enum';
import {
    defaultApp, deviceTitle, filterLogLines, isProblem, keyForEvent, leaseNote, logLevel, runsOn, unsupportedNotice,
} from './deviceControlModel';

describe('keyForEvent', () => {
    it.each([
        ['Enter', {key: 'enter'}],
        ['Escape', {key: 'back'}],
        ['Delete', {key: '112'}],
        ['ArrowUp', {key: '19'}],
        ['a', {text: 'a'}],
        ['€', {text: '€'}],
    ])('maps %s', (key, expected) => {
        expect(keyForEvent({key})).toEqual(expected);
    });

    it.each([
        ['a shortcut with ctrl', {key: 'c', ctrlKey: true}],
        ['a shortcut with meta', {key: 'c', metaKey: true}],
        ['a composition', {key: 'a', isComposing: true}],
        ['a key without meaning', {key: 'Shift'}],
    ])('ignores %s', (_, event) => {
        expect(keyForEvent(event)).toBeNull();
    });
});

describe('logLevel', () => {
    it.each([
        ['11-06 13:51:05.123  1234  5678 E Tag: msg', 'E'],
        ['11-06 13:51:05.123 W/Tag( 1234): msg', 'W'],
        ['I/Tag( 1234): msg', 'I'],
    ])('reads the level of %s', (line, level) => {
        expect(logLevel(line)).toBe(level);
    });

    it('is null for other lines', () => {
        expect(logLevel('--------- beginning of main')).toBeNull();
    });

    it('counts warnings and errors as problems', () => {
        expect(['E', 'F', 'W'].map(l => isProblem(l as 'E'))).toEqual([true, true, true]);
        expect([isProblem('I'), isProblem(null)]).toEqual([false, false]);
    });
});

describe('filterLogLines', () => {
    const lines = ['I/App( 1): hello', 'E/App( 1): Boom', 'plain'];

    it('keeps everything without filters', () => {
        expect(filterLogLines(lines, ' ', false)).toEqual(lines);
    });

    it('finds text without regard to case', () => {
        expect(filterLogLines(lines, 'boom', false)).toEqual(['E/App( 1): Boom']);
    });

    it('can keep only problems', () => {
        expect(filterLogLines(lines, '', true)).toEqual(['E/App( 1): Boom']);
    });
});

describe('leaseNote', () => {
    const now = Date.parse('2026-01-01T10:00:00Z');
    const at = (minutes: number): string => new Date(now + minutes * 60000).toISOString();

    it.each([
        [undefined, null],
        ['not a date', null],
        [at(0), 'ends now unless you use the device'],
        [at(1), 'ends in a minute unless you use the device'],
        [at(10), 'ends in 10 minutes unless you use the device'],
    ])('says %s', (expiresAt, note) => {
        expect(leaseNote(expiresAt, now)).toBe(note);
    });
});

describe('apps', () => {
    const apps = [
        {id: 1, platform: PlatformType.Android},
        {id: 2, platform: PlatformType.iOS},
        {id: 3, platform: PlatformType.Android},
    ];

    it('knows what runs where', () => {
        expect([
            runsOn(PlatformType.iOS, PlatformType.iOS),
            runsOn(PlatformType.iOS, PlatformType.IOSSimulator),
            runsOn(PlatformType.Android, PlatformType.iOS),
        ]).toEqual([true, true, false]);
    });

    it('offers the remembered app when it fits', () => {
        expect(defaultApp(apps, PlatformType.Android, 3)?.id).toBe(3);
    });

    it('offers the first app that fits otherwise', () => {
        expect(defaultApp(apps, PlatformType.iOS, 3)?.id).toBe(2);
    });

    it('offers the first app for an unknown device and none for no fit', () => {
        expect(defaultApp(apps, undefined, null)?.id).toBe(1);
        expect(defaultApp(apps, PlatformType.Web, null)).toBeNull();
    });
});

describe('deviceTitle', () => {
    it.each([
        [{name: 'Pixel', alias: 'Lab 1'}, 'Lab 1'],
        [{name: 'Pixel', alias: ''}, 'Pixel'],
        [null, 'Device 4'],
    ])('names %j', (device, title) => {
        expect(deviceTitle(device, 4)).toBe(title);
    });
});

describe('unsupportedNotice', () => {
    it('explains a reason', () => {
        expect(unsupportedNotice('WDA failed')).toBe('Live video unavailable, showing screenshots: WDA failed');
    });

    it('stays silent for the plain answer and for none', () => {
        expect(unsupportedNotice('operation not supported by this device')).toBeNull();
        expect(unsupportedNotice()).toBeNull();
    });
});
