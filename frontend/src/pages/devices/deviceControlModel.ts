// The logic of the device control page that doesn't need React: what a key
// press on the computer sends to the device, picking an element of the scene
// graph by position, the system log, the lease and the default app.
import {PlatformType} from '../../types/platform.type.enum';

// --- keyboard ---

export type DeviceKey = { text: string } | { key: string };

// key codes of keys without a name in the hub (Android KEYCODE_*)
const DPAD: Record<string, string> = {ArrowUp: '19', ArrowDown: '20', ArrowLeft: '21', ArrowRight: '22'};

const NAMED: Record<string, string> = {
    Enter: 'enter',
    Backspace: 'delete',
    Tab: 'tab',
    Escape: 'back',
    Delete: '112', // KEYCODE_FORWARD_DEL
    ...DPAD,
};

export interface KeyLike {
    key: string,
    ctrlKey?: boolean,
    metaKey?: boolean,
    altKey?: boolean,
    isComposing?: boolean,
}

// keyForEvent says what a key press on the computer means on the device:
// a character to type, a named key, or nothing (shortcuts stay with the
// browser).
export const keyForEvent = (e: KeyLike): DeviceKey | null => {
    if (e.ctrlKey || e.metaKey || e.isComposing) {
        return null;
    }
    if (NAMED[e.key]) {
        return {key: NAMED[e.key]};
    }
    // a single character, including AltGr characters like @ and €
    if ([...e.key].length === 1) {
        return {text: e.key};
    }
    return null;
};

// --- scene graph ---

export interface SceneElement {
    getAttribute(name: string): string | null,
    readonly children: ArrayLike<SceneElement>,
}

const rect = (e: SceneElement) => {
    const n = (a: string) => Number(e.getAttribute(a)) || 0;
    return {x: n('X'), y: n('Y'), w: n('RectangleX'), h: n('RectangleY')};
};

// elementAt finds the deepest visible element at a point in scene
// coordinates; later siblings are drawn on top, so they win
export const elementAt = (root: SceneElement, x: number, y: number): SceneElement | null => {
    const inside = (e: SceneElement) => {
        const r = rect(e);
        return r.w > 0 && r.h > 0 && x >= r.x && y >= r.y && x < r.x + r.w && y < r.y + r.h;
    };
    const visit = (e: SceneElement): SceneElement | null => {
        const kids = Array.from(e.children);
        for (let i = kids.length - 1; i >= 0; i--) {
            const hit = visit(kids[i]);
            if (hit) return hit;
        }
        return inside(e) ? e : null;
    };
    return visit(root);
};

// --- system log ---

export type LogLevel = 'V' | 'D' | 'I' | 'W' | 'E' | 'F';

// logLevel reads the level of a logcat line in the threadtime format
// ("11-06 13:51:05.123  1234  5678 E Tag: msg"), the time format
// ("11-06 13:51:05.123 E/Tag( 1234): msg") or the brief one
// ("E/Tag( 1234): msg"); null for other lines
export const logLevel = (line: string): LogLevel | null => {
    const threadtime = line.match(/^\d\d-\d\d\s+[\d:.]+\s+\d+\s+\d+\s+([VDIWEF])\s/);
    if (threadtime) return threadtime[1] as LogLevel;
    const tagged = line.match(/^(?:\d\d-\d\d\s+[\d:.]+\s+)?([VDIWEF])\/[^(]*\(\s*\d+\):/);
    return tagged ? tagged[1] as LogLevel : null;
};

export const isProblem = (level: LogLevel | null) => level === 'E' || level === 'F' || level === 'W';

export const filterLogLines = (lines: string[], text: string, problemsOnly: boolean): string[] => {
    const needle = text.trim().toLowerCase();
    return lines.filter(l => (!problemsOnly || isProblem(logLevel(l))) && (needle === '' || l.toLowerCase().includes(needle)));
};

// --- lease ---

// leaseNote says how long an idle lease lasts, e.g. "ends at 14:12 if idle"
export const leaseNote = (expiresAt: string | undefined, now: number): string | null => {
    if (!expiresAt) return null;
    const end = new Date(expiresAt).valueOf();
    if (isNaN(end)) return null;
    const minutes = Math.round((end - now) / 60000);
    if (minutes <= 0) return 'ends now unless you use the device';
    if (minutes === 1) return 'ends in a minute unless you use the device';
    return `ends in ${minutes} minutes unless you use the device`;
};

// --- apps ---

// app platforms that run on a device platform
export const runsOn = (app: PlatformType, device: PlatformType): boolean =>
    app === device || (app === PlatformType.iOS && device === PlatformType.IOSSimulator);

// defaultApp picks the app to offer for a device: the remembered app when it
// runs there, else the first that does
export const defaultApp = <T extends { id: number, platform: PlatformType }>(apps: T[], device: PlatformType | undefined, remembered: number | null): T | null => {
    const fitting = device === undefined ? apps : apps.filter(a => runsOn(a.platform, device));
    return fitting.find(a => a.id === remembered) ?? fitting[0] ?? null;
};
