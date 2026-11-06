// The bundles page without React: what files an app takes, searching the
// list, tags, and how long ago a bundle was uploaded.
import {PlatformType} from '../../types/platform.type.enum';

export interface Accept {
    // extensions with the dot, lower case; empty means any
    extensions: string[],
    label: string,
}

// acceptFor says which files the master can read for an app's platform
// (apk for Android, ipa or a zipped app for iOS and Mac)
export const acceptFor = (platform: PlatformType | undefined): Accept => {
    switch (platform) {
        case PlatformType.Android:
            return {extensions: ['.apk'], label: '.apk'};
        case PlatformType.iOS:
        case PlatformType.IOSSimulator:
            return {extensions: ['.ipa', '.zip'], label: '.ipa or a zipped .app'};
        case PlatformType.Mac:
            return {extensions: ['.zip'], label: 'a zipped .app'};
        default:
            return {extensions: ['.apk', '.ipa', '.zip'], label: '.apk, .ipa or .zip'};
    }
};

export const accepts = (accept: Accept, fileName: string): boolean => {
    if (accept.extensions.length === 0) return true;
    const name = fileName.toLowerCase();
    return accept.extensions.some(ext => name.endsWith(ext));
};

// parseTags splits the comma separated tags the master stores
export const parseTags = (tags: string | null | undefined): string[] =>
    Array.from(new Set((tags ?? '').split(',').map(t => t.trim()).filter(t => t.length > 0)));

export const joinTags = (tags: string[]): string => parseTags(tags.join(',')).join(',');

export interface Searchable {
    id: number,
    name: string,
    version: string,
    tags: string,
    hash?: string,
}

// matches finds bundles by version, name, tag, id or the start of the hash
export const matches = (b: Searchable, text: string): boolean => {
    const needle = text.trim().toLowerCase();
    if (needle === '') return true;
    return (b.version ?? '').toLowerCase().includes(needle)
        || (b.name ?? '').toLowerCase().includes(needle)
        || parseTags(b.tags).some(t => t.toLowerCase().includes(needle))
        || `#${b.id}` === needle || String(b.id) === needle
        || (needle.length >= 6 && (b.hash ?? '').toLowerCase().startsWith(needle));
};

// ago says how long ago something happened, in the largest fitting unit
export const ago = (date: Date | string, now: number): string => {
    const seconds = Math.max(0, Math.round((now - new Date(date).valueOf()) / 1000));
    const units: [number, string][] = [[86400 * 365, 'year'], [86400 * 30, 'month'], [86400 * 7, 'week'], [86400, 'day'], [3600, 'hour'], [60, 'minute']];
    for (const [size, unit] of units) {
        const n = Math.floor(seconds / size);
        if (n >= 1) return `${n} ${unit}${n === 1 ? '' : 's'} ago`;
    }
    return 'just now';
};
