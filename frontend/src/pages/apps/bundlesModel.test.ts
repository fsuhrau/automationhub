import {expect, it} from 'vitest';
import {PlatformType} from '../../types/platform.type.enum';
import {bundle} from './appsFixtures';
import {
    acceptFor, accepts, ago, bundleLabel, bundleVersion, downloadUrl, iconUrlOf, isInstallable, joinTags, matches, newestFirst, parseTags, tagsText,
} from './bundlesModel';

it.each([
    [PlatformType.Android, ['.apk']],
    [PlatformType.iOS, ['.ipa', '.zip']],
    [PlatformType.IOSSimulator, ['.ipa', '.zip']],
    [PlatformType.Mac, ['.zip']],
    [PlatformType.Web, ['.apk', '.ipa', '.zip']],
    [undefined, ['.apk', '.ipa', '.zip']],
])('takes the files of platform %s', (platform, extensions) => {
    expect(acceptFor(platform).extensions).toEqual(extensions);
});

it('accepts a file by its extension, in any case', () => {
    const accept = acceptFor(PlatformType.Android);
    expect(accepts(accept, 'Game.APK')).toBe(true);
    expect(accepts(accept, 'game.ipa')).toBe(false);
    expect(accepts({extensions: [], label: 'any'}, 'whatever')).toBe(true);
});

it('splits tags, trims them and drops empty and repeated ones', () => {
    expect(parseTags(' a, b ,,a ')).toEqual(['a', 'b']);
    expect(parseTags(null)).toEqual([]);
    expect(parseTags(undefined)).toEqual([]);
    expect(joinTags([' x', 'y,x', ''])).toBe('x,y');
    expect(tagsText({tags: 'a,b'})).toBe('a, b');
});

it.each([
    ['1.2', true], ['QUEST', true], ['rc', true], ['#3', true], ['3', true], ['hash3', false], ['hash30', true], ['nope', false], ['  ', true],
])('matches a bundle with the text "%s": %s', (text, found) => {
    expect(matches(bundle(3, {tags: 'rc', version: '1.2', name: 'quest.apk'}), text)).toBe(found);
});

it('matches a bundle without version, name or hash only by id and tags', () => {
    const bare = {id: 4, tags: 'x'} as Parameters<typeof matches>[0];
    expect(matches(bare, 'x')).toBe(true);
    expect(matches(bare, 'abcdefg')).toBe(false);
});

it.each([
    [0, 'just now'], [60, '1 minute ago'], [180, '3 minutes ago'], [3600, '1 hour ago'], [86400 * 2, '2 days ago'],
    [86400 * 7, '1 week ago'], [86400 * 40, '1 month ago'], [86400 * 800, '2 years ago'], [-50, 'just now'],
])('says %s seconds ago as "%s"', (seconds, text) => {
    expect(ago(new Date(1_000_000_000_000 - seconds * 1000), 1_000_000_000_000)).toBe(text);
});

it('builds the download url of a file and says when it is there', () => {
    expect(downloadUrl('a b/c.apk')).toBe('/upload/a%20b%2Fc.apk');
    expect(isInstallable({appPath: 'x.apk'})).toBe(true);
    expect(isInstallable({appPath: ''})).toBe(false);
    expect(isInstallable({appPath: 'x.apk', archivedAt: '2020'})).toBe(false);
});

it('names a bundle by version from the file, stored version or name', () => {
    expect(bundleVersion({name: 'n', version: 'v', info: {version: 'i'}})).toBe('i');
    expect(bundleVersion({name: 'n', version: 'v'})).toBe('v');
    expect(bundleVersion({name: 'n', version: ''})).toBe('n');
});

it('labels a bundle with its build when that differs from the version', () => {
    expect(bundleLabel({id: 7, name: 'n', version: '2.0', info: {build: '45'}})).toBe('2.0 (45) · #7');
    expect(bundleLabel({id: 7, name: 'n', version: '2.0', info: {build: '2.0'}})).toBe('2.0 · #7');
    expect(bundleLabel({id: 7, name: '', version: ''})).toBe('– · #7');
});

it('sorts newest first without touching the list', () => {
    const list = [{id: 1}, {id: 3}, {id: 2}];
    expect(newestFirst(list).map(b => b.id)).toEqual([3, 2, 1]);
    expect(list.map(b => b.id)).toEqual([1, 3, 2]);
});

it('serves the icon only when the file had one', () => {
    const app = {projectId: 'p q', appId: 2};
    expect(iconUrlOf(app, {id: 5, info: {icon: 'x.png'}})).toBe('/api/p%20q/app/2/bundle/5/icon');
    expect(iconUrlOf(app, {id: 5})).toBeUndefined();
    expect(iconUrlOf(null, {id: 5, info: {icon: 'x.png'}})).toBeUndefined();
});
