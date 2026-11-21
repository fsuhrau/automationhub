// Fakes for the tests of the control page: a device control API with canned
// answers and a video player that records its input.
import {vi} from 'vitest';
import IDeviceData from '../../../types/device';
import {IAppBinaryData, IAppData} from '../../../types/app';
import {DeviceControlApi} from '../../../types/device.control.api';
import {DeviceLease, DeviceLockStatus} from '../../../types/device.control.lease';
import {PlayerCallbacks, VideoPlayerLike, VideoSupport} from '../../../types/device.control.video';
import {VideoInput} from '../../../services/device.video';
import {PlatformType} from '../../../types/platform.type.enum';

export const FREE: DeviceLockStatus = {locked: false, lockedByTest: false};

export const leaseOf = (over: Partial<DeviceLease> = {}): DeviceLease => ({
    deviceId: '7', token: 'tok-1', owner: 'anna', acquiredAt: '2026-01-01T09:00:00Z', lastUsed: '2026-01-01T09:00:00Z',
    expiresAt: '2099-01-01T10:00:00Z', ...over,
});

export const heldBy = (over: Partial<DeviceLease> = {}): DeviceLockStatus => ({locked: true, lockedByTest: false, lease: leaseOf(over)});

export const DEVICE = {
    id: 7, name: 'Pixel 8', alias: 'Lab phone', os: 'Android', osVersion: '14', platformType: PlatformType.Android, node: {name: 'node-1'},
} as unknown as IDeviceData;

export const APPS = [
    {id: 1, name: 'Game', platform: PlatformType.Android},
    {id: 2, name: 'Game iOS', platform: PlatformType.iOS},
] as unknown as IAppData[];

export const BUNDLES = [
    {id: 11, name: 'game.apk', version: '1.0', createdAt: '2026-01-01T09:00:00Z'},
    {id: 12, name: 'game.apk', version: '1.1', createdAt: '2026-01-02T09:00:00Z'},
    {id: 13, name: 'old.apk', version: '0.9', createdAt: '2025-01-01T09:00:00Z', archivedAt: '2025-06-01T00:00:00Z'},
] as unknown as IAppBinaryData[];

// a fake API where every call succeeds; override what a test needs
export const fakeService = (over: Partial<DeviceControlApi> = {}): DeviceControlApi => ({
    device: vi.fn().mockResolvedValue(DEVICE),
    apps: vi.fn().mockResolvedValue(APPS),
    bundles: vi.fn().mockResolvedValue(BUNDLES),
    lockStatus: vi.fn().mockResolvedValue(FREE),
    lock: vi.fn().mockResolvedValue(leaseOf()),
    release: vi.fn().mockResolvedValue(undefined),
    tap: vi.fn().mockResolvedValue(undefined),
    swipe: vi.fn().mockResolvedValue(undefined),
    text: vi.fn().mockResolvedValue(undefined),
    key: vi.fn().mockResolvedValue(undefined),
    touchElement: vi.fn().mockResolvedValue(undefined),
    hierarchy: vi.fn().mockResolvedValue('<Scene/>'),
    logs: vi.fn().mockResolvedValue([]),
    appAction: vi.fn().mockResolvedValue(undefined),
    ...over,
});

export class FakePlayer implements VideoPlayerLike {
    canSendInput = true;
    closed = false;
    sent: { input: VideoInput, lease: string | null }[] = [];

    constructor(readonly url: string, readonly canvas: HTMLCanvasElement, readonly callbacks: PlayerCallbacks) {
    }

    sendInput(input: VideoInput, lease: string | null): boolean {
        this.sent.push({input, lease});
        return true;
    }

    toVideo(clientX: number, clientY: number): { x: number, y: number, width: number, height: number } {
        return {x: clientX, y: clientY, width: 360, height: 640};
    }

    close(): void {
        this.closed = true;
    }
}

export interface FakeSupport extends VideoSupport {
    players: FakePlayer[],
}

export const fakeSupport = (over: Partial<VideoSupport> = {}): FakeSupport => {
    const players: FakePlayer[] = [];
    return {
        players,
        decodesVideo: true,
        decodesMjpeg: true,
        createPlayer: (url, canvas, callbacks) => {
            const player = new FakePlayer(url, canvas, callbacks);
            players.push(player);
            return player;
        },
        ...over,
    };
};
