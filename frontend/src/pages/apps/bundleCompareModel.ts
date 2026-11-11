// Comparing two bundles without React: which attributes changed from the
// older bundle (a) to the newer one (b), which entries of lists like
// permissions or frameworks came and went, and where the size moved.

import {IAppBinaryData} from '../../types/app';
import {BundleInfo, SizeEntry} from '../../types/bundle.info';
import {byteFormat} from '../tests/value_formatter';
import {apiLevel, commonName, entitlementValue, orientationName, PROFILE_KINDS, shortPermission, usageName} from './bundleDetailModel';

export type Change = 'same' | 'changed' | 'added' | 'removed';

// ValueRow is one attribute of both bundles
export interface ValueRow {
    kind: 'value',
    label: string,
    a?: string,
    b?: string,
    change: Change,
    // a as a number and b minus a, for numbers like sizes
    from?: number,
    delta?: number,
    // how to show the delta
    unit?: 'bytes' | 'count',
}

// ListRow is a list attribute, like the permissions: what only one of the
// bundles has
export interface ListRow {
    kind: 'list',
    label: string,
    added: string[],
    removed: string[],
    same: number,
    change: Change,
}

export type Row = ValueRow | ListRow;

export interface Section {
    title: string,
    rows: Row[],
}

// SizeRow is one part of the bundle contents in both bundles
export interface SizeRow {
    path: string,
    a?: number,
    b?: number,
    // b minus a; unknown when a list is partial and one side lacks the entry
    delta?: number,
    change: Change,
}

export interface Comparison {
    sections: Section[],
    // the parts of the contents, biggest change first
    parts: SizeRow[],
    // the largest files of both, biggest change first
    files: SizeRow[],
    changes: number,
}

const present = (v: string | undefined): v is string => v !== undefined && v !== '';

const changeOf = (a: string | undefined, b: string | undefined): Change =>
    !present(a) && !present(b) ? 'same'
        : !present(a) ? 'added'
            : !present(b) ? 'removed'
                : a === b ? 'same' : 'changed';

const value = (label: string, a: string | undefined, b: string | undefined): ValueRow =>
    ({kind: 'value', label, a: present(a) ? a : undefined, b: present(b) ? b : undefined, change: changeOf(a, b)});

const number = (label: string, a: number | undefined, b: number | undefined, unit: 'bytes' | 'count'): ValueRow => {
    const show = (n: number | undefined) => n === undefined ? undefined : unit === 'bytes' ? byteFormat(n) : n.toLocaleString();
    return {
        kind: 'value', label, a: show(a), b: show(b), unit, from: a,
        change: a === undefined && b === undefined ? 'same' : a === undefined ? 'added' : b === undefined ? 'removed' : a === b ? 'same' : 'changed',
        delta: a !== undefined && b !== undefined ? b - a : undefined,
    };
};

const flag = (label: string, a: boolean | undefined, b: boolean | undefined): ValueRow =>
    value(label, a === undefined ? undefined : a ? 'yes' : 'no', b === undefined ? undefined : b ? 'yes' : 'no');

// list compares two lists as sets; a missing list is an empty one
export const list = (label: string, a: string[] | undefined, b: string[] | undefined): ListRow => {
    const inA = new Set(a ?? []);
    const inB = new Set(b ?? []);
    const added = [...inB].filter(x => !inA.has(x)).sort();
    const removed = [...inA].filter(x => !inB.has(x)).sort();
    return {
        kind: 'list', label, added, removed, same: [...inA].filter(x => inB.has(x)).length,
        change: added.length === 0 && removed.length === 0 ? 'same' : 'changed',
    };
};

// signedBytes is a size change with its sign, like "+1.20MB"
export const signedBytes = (delta: number): string =>
    delta === 0 ? '±0' : `${delta > 0 ? '+' : '−'}${byteFormat(Math.abs(delta))}`;

// signedCount is a count change with its sign, like "+3"
export const signedCount = (delta: number): string =>
    delta === 0 ? '±0' : `${delta > 0 ? '+' : '−'}${Math.abs(delta).toLocaleString()}`;

// percent is the change relative to a, or undefined when a is 0
export const percent = (a: number | undefined, delta: number): string | undefined => {
    if (!a) return undefined;
    const p = delta / a * 100;
    if (Math.abs(p) < 0.05) return undefined;
    return `${p > 0 ? '+' : '−'}${Math.abs(p).toFixed(1)} %`;
};

// sizes compares two size lists by path. A complete list (the parts) misses
// an entry when it isn't there; a partial one (the largest files) only when
// it is smaller, so the change of such an entry is unknown.
export const sizes = (a: SizeEntry[] | undefined, b: SizeEntry[] | undefined, partial = false): SizeRow[] => {
    const byPath = new Map<string, SizeRow>();
    for (const e of a ?? []) byPath.set(e.path, {path: e.path, a: e.size, delta: 0, change: 'removed'});
    for (const e of b ?? []) {
        const row = byPath.get(e.path);
        if (row) {
            row.b = e.size;
            row.delta = e.size - (row.a ?? 0);
            row.change = row.delta === 0 ? 'same' : 'changed';
        } else {
            byPath.set(e.path, {path: e.path, b: e.size, delta: e.size, change: 'added'});
        }
    }
    for (const row of byPath.values()) {
        if (row.change === 'removed') row.delta = -(row.a ?? 0);
        if (partial && (row.change === 'added' || row.change === 'removed')) row.delta = undefined;
    }
    const weight = (r: SizeRow) => r.delta === undefined ? -1 : Math.abs(r.delta);
    return [...byPath.values()].sort((x, y) => weight(y) - weight(x) || x.path.localeCompare(y.path));
};

const engine = (info?: BundleInfo) => info?.engine && [info.engine.name, info.engine.version].filter(Boolean).join(' ');

const date = (d: string | undefined) => d ? new Date(d).toISOString().slice(0, 10) : undefined;

// compareBundles compares bundle a with bundle b; a is usually the older one
export const compareBundles = (a: IAppBinaryData, b: IAppBinaryData): Comparison => {
    const ia = a.info, ib = b.info;
    const sections: Section[] = [];

    sections.push({
        title: 'App', rows: [
            value('Name', ia?.displayName || a.name, ib?.displayName || b.name),
            value('Identifier', ia?.identifier || a.identifier, ib?.identifier || b.identifier),
            value('Version', ia?.version || a.version, ib?.version || b.version),
            value('Build', ia?.build, ib?.build),
            value('Format', ia?.format, ib?.format),
            value('Platform', a.platform, b.platform),
            value('Engine', engine(ia), engine(ib)),
            value('Scripting backend', ia?.engine?.backend, ib?.engine?.backend),
            list('Tags', a.tags ? a.tags.split(',') : [], b.tags ? b.tags.split(',') : []),
        ],
    });

    sections.push({
        title: 'Size', rows: [
            number('File size', a.size || undefined, b.size || undefined, 'bytes'),
            number('Installed size', ia?.contents?.size, ib?.contents?.size, 'bytes'),
            number('Compressed', ia?.contents?.compressed, ib?.contents?.compressed, 'bytes'),
            number('Files', ia?.contents?.files, ib?.contents?.files, 'count'),
        ],
    });

    const aa = ia?.android, ab = ib?.android;
    if (aa || ab) {
        const permissions = (p?: typeof aa) => p?.permissions?.map(x => shortPermission(x.name) + (x.maxSdk ? ` (up to API ${x.maxSdk})` : ''));
        const features = (p?: typeof aa) => p?.features?.map(f => `${f.name.replace(/^android\.hardware\./, '')}${f.required ? '' : ' (optional)'}`);
        const level = (l?: string) => l ? apiLevel(l) : undefined;
        sections.push({
            title: 'Android', rows: [
                value('Package', aa?.packageName, ab?.packageName),
                value('Version name', aa?.versionName, ab?.versionName),
                value('Version code', aa?.versionCode, ab?.versionCode),
                value('Min SDK', level(aa?.minSdk), level(ab?.minSdk)),
                value('Target SDK', level(aa?.targetSdk), level(ab?.targetSdk)),
                value('Max SDK', level(aa?.maxSdk), level(ab?.maxSdk)),
                value('Compile SDK', level(aa?.compileSdk), level(ab?.compileSdk)),
                value('Launch activity', aa?.launchActivity, ab?.launchActivity),
                value('OpenGL ES', aa?.gles, ab?.gles),
                flag('Debuggable', aa?.debuggable, ab?.debuggable),
                flag('Test only', aa?.testOnly, ab?.testOnly),
                flag('Android TV', aa?.leanback, ab?.leanback),
                list('Permissions', permissions(aa), permissions(ab)),
                list('Hardware features', features(aa), features(ab)),
                list('ABIs', aa?.abis, ab?.abis),
                list('Locales', aa?.locales, ab?.locales),
                list('Screens', aa?.screens, ab?.screens),
                list('Densities', aa?.densities, ab?.densities),
            ],
        });
    }

    const pa = ia?.apple, pb = ib?.apple;
    if (pa || pb) {
        const usage = (p?: typeof pa) => p?.usageDescriptions?.map(u => `${usageName(u.key)}: ${u.text}`);
        const xcode = (p?: typeof pa) => p?.xcode && `${p.xcode}${p.xcodeBuild ? ` (${p.xcodeBuild})` : ''}`;
        sections.push({
            title: pa?.platform === 'macosx' || pb?.platform === 'macosx' ? 'macOS' : 'iOS', rows: [
                value('Bundle name', pa?.bundleName, pb?.bundleName),
                value('Executable', pa?.executable, pb?.executable),
                value('Minimum OS', pa?.minOs, pb?.minOs),
                value('SDK', pa?.sdk, pb?.sdk),
                value('Xcode', xcode(pa), xcode(pb)),
                value('Built on macOS', pa?.buildMachineOs, pb?.buildMachineOs),
                flag('File sharing', pa?.fileSharing, pb?.fileSharing),
                list('Devices', pa?.deviceFamily, pb?.deviceFamily),
                list('Orientations', pa?.orientations?.map(orientationName), pb?.orientations?.map(orientationName)),
                list('Requires', pa?.requiredCapabilities, pb?.requiredCapabilities),
                list('Background modes', pa?.backgroundModes, pb?.backgroundModes),
                list('URL schemes', pa?.urlSchemes, pb?.urlSchemes),
                list('Permissions', usage(pa), usage(pb)),
                list('Extensions', pa?.extensions, pb?.extensions),
                list('Frameworks', pa?.frameworks, pb?.frameworks),
            ],
        });
        const fa = pa?.profile, fb = pb?.profile;
        if (fa || fb) {
            const kind = (k?: string) => k ? PROFILE_KINDS[k] ?? k : undefined;
            const devices = (p?: typeof fa) => p ? p.allDevices ? 'every device' : `${p.devices?.length ?? 0}` : undefined;
            const entitlements = (p?: typeof fa) => Object.entries(p?.entitlements ?? {}).map(([k, v]) => `${k} = ${entitlementValue(v)}`);
            sections.push({
                title: 'Provisioning profile', rows: [
                    value('Name', fa?.name, fb?.name),
                    value('Kind', kind(fa?.kind), kind(fb?.kind)),
                    value('Team', fa?.teamName, fb?.teamName),
                    value('UUID', fa?.uuid, fb?.uuid),
                    value('Expires', date(fa?.expiresAt), date(fb?.expiresAt)),
                    value('Devices', devices(fa), devices(fb)),
                    list('Device UDIDs', fa?.devices, fb?.devices),
                    list('Entitlements', fa ? entitlements(fa) : undefined, fb ? entitlements(fb) : undefined),
                ],
            });
        }
    }

    if (ia?.signing || ib?.signing) {
        const certs = (i?: BundleInfo) => i?.signing?.certificates?.map(c => `${commonName(c.subject)} (${c.sha1.slice(0, 8)}…)${c.debug ? ', debug key' : ''}`);
        const until = (i?: BundleInfo) => i?.signing?.certificates?.map(c => date(c.notAfter)).filter(present).sort()[0];
        sections.push({
            title: 'Signing', rows: [
                list('Signature schemes', ia?.signing?.schemes, ib?.signing?.schemes),
                list('Certificates', certs(ia), certs(ib)),
                value('First certificate expiry', until(ia), until(ib)),
            ],
        });
    }

    sections.push({
        title: 'Warnings', rows: [list('Warnings', ia?.warnings, ib?.warnings)],
    });

    const parts = sizes(ia?.contents?.groups, ib?.contents?.groups);
    const files = sizes(ia?.contents?.largest, ib?.contents?.largest, true);
    const changes = sections.reduce((n, s) => n + s.rows.filter(r => r.change !== 'same').length, 0);
    return {sections, parts, files, changes};
};

// onlyChanges drops the rows and sections without a change
export const onlyChanges = (sections: Section[]): Section[] =>
    sections.map(s => ({...s, rows: s.rows.filter(r => r.change !== 'same')})).filter(s => s.rows.length > 0);

// hideEmpty drops rows that neither bundle has
export const hideEmpty = (sections: Section[]): Section[] =>
    sections.map(s => ({
        ...s, rows: s.rows.filter(r => r.kind === 'value' ? r.a !== undefined || r.b !== undefined : r.same + r.added.length + r.removed.length > 0),
    })).filter(s => s.rows.length > 0);

// previousOf is the bundle uploaded before the given one, for the default
// comparison; bundles are compared to the one before them
export const previousOf = (bundles: IAppBinaryData[], id: number): IAppBinaryData | undefined =>
    bundles.filter(b => b.id < id).sort((x, y) => y.id - x.id)[0];
