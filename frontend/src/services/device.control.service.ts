import http from '../http-common';

export const LEASE_HEADER = 'X-Device-Lease';

export interface DeviceLease {
    deviceId: string,
    token?: string,
    owner: string,
    acquiredAt: string,
    lastUsed: string,
    expiresAt: string,
}

export interface DeviceLockStatus {
    locked: boolean,
    lease?: DeviceLease,
    lockedByTest: boolean,
}

const base = (projectId: string, deviceId: number | string) => `/${encodeURIComponent(projectId)}/device/${encodeURIComponent(String(deviceId))}`;
const withLease = (token: string) => ({headers: {[LEASE_HEADER]: token}});

export const streamUrl = (projectId: string, deviceId: number | string, fps: number = 2): string =>
    `/api${base(projectId, deviceId)}/stream?fps=${fps}`;

export const screenshotUrl = (projectId: string, deviceId: number | string): string =>
    `/api${base(projectId, deviceId)}/screenshot?t=${Date.now()}`;

export const getLockStatus = (projectId: string, deviceId: number): Promise<DeviceLockStatus> =>
    http.get(`${base(projectId, deviceId)}/lock`).then(r => r.data);

export const lockDevice = (projectId: string, deviceId: number): Promise<DeviceLease> =>
    http.post(`${base(projectId, deviceId)}/lock`).then(r => r.data);

export const releaseDevice = (projectId: string, deviceId: number, token: string): Promise<void> =>
    http.delete(`${base(projectId, deviceId)}/lock`, withLease(token)).then(() => undefined);

export const tap = (projectId: string, deviceId: number, token: string, x: number, y: number): Promise<void> =>
    http.post(`${base(projectId, deviceId)}/tap`, {x, y}, withLease(token)).then(() => undefined);

export const swipe = (projectId: string, deviceId: number, token: string, fromX: number, fromY: number, toX: number, toY: number, durationMs: number): Promise<void> =>
    http.post(`${base(projectId, deviceId)}/swipe`, {fromX, fromY, toX, toY, durationMs}, withLease(token)).then(() => undefined);

export const inputText = (projectId: string, deviceId: number, token: string, text: string): Promise<void> =>
    http.post(`${base(projectId, deviceId)}/text`, {text}, withLease(token)).then(() => undefined);

export const pressKey = (projectId: string, deviceId: number, token: string, key: string): Promise<void> =>
    http.post(`${base(projectId, deviceId)}/key`, {key}, withLease(token)).then(() => undefined);

export const touchElement = (projectId: string, deviceId: number, token: string, ref: { elementId?: string, xpath?: string }): Promise<void> =>
    http.post(`${base(projectId, deviceId)}/element/touch`, ref, withLease(token)).then(() => undefined);

export const getHierarchy = (projectId: string, deviceId: number): Promise<string> =>
    http.get(`${base(projectId, deviceId)}/hierarchy`, {responseType: 'text', transformResponse: r => r}).then(r => r.data);

export const getDeviceLogs = (projectId: string, deviceId: number, lines: number = 300): Promise<string[]> =>
    http.get(`${base(projectId, deviceId)}/logs`, {params: {lines}}).then(r => r.data.lines);

export type AppAction = 'install' | 'start' | 'stop';

export const appAction = (projectId: string, deviceId: number, token: string, action: AppAction, binaryId: number, clearData: boolean = false): Promise<void> =>
    http.post(`${base(projectId, deviceId)}/app/${action}`, {binaryId, clearData}, withLease(token)).then(() => undefined);
