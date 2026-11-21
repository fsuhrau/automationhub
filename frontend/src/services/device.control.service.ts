import type {AxiosRequestConfig} from 'axios';
import {ApiClient} from '../api/ApiClient';
import {appService} from './app.service';
import {deviceService} from './device.service';
import {AppAction, AppRequest, DeviceControlApi, DeviceTarget, ElementRef, Point, Swipe} from '../types/device.control.api';

export const LEASE_HEADER = 'X-Device-Lease';

const base = (projectId: string, deviceId: number | string): string => `/${encodeURIComponent(projectId)}/device/${encodeURIComponent(String(deviceId))}`;
const withLease = (token: string): AxiosRequestConfig => ({headers: {[LEASE_HEADER]: token}});

export const streamUrl = (projectId: string, deviceId: number | string, fps: number = 2): string =>
    `/api${base(projectId, deviceId)}/stream?fps=${fps}`;

// server-sent events: "lines" (JSON array of new log lines) and "failure"
export const logStreamUrl = (projectId: string, deviceId: number | string, lines: number = 300): string =>
    `/api${base(projectId, deviceId)}/logs/stream?lines=${lines}`;

export const screenshotUrl = (projectId: string, deviceId: number | string): string =>
    `/api${base(projectId, deviceId)}/screenshot?t=${Date.now()}`;

// deviceControlService talks to the hub about one device. The calls that
// change the device send the lease token in the lease header.
export const deviceControlService = (api: ApiClient, {projectId, deviceId}: DeviceTarget): DeviceControlApi => {
    const device = base(projectId, deviceId);
    const devices = deviceService(api);
    const apps = appService(api);
    const send = (path: string, token: string, body?: unknown): Promise<void> =>
        api.post<unknown>(`${device}${path}`, body, withLease(token)).then(() => undefined);
    return {
        device: () => devices.get(projectId, String(deviceId)),
        apps: () => apps.list(projectId),
        bundles: (appId: number) => apps.bundles({projectId, appId}),
        lockStatus: () => api.get(`${device}/lock`),
        lock: () => api.post(`${device}/lock`),
        release: (token: string) => api.delete<unknown>(`${device}/lock`, withLease(token)).then(() => undefined),
        tap: (token: string, at: Point) => send('/tap', token, at),
        swipe: (token: string, {from, to, durationMs}: Swipe) =>
            send('/swipe', token, {fromX: from.x, fromY: from.y, toX: to.x, toY: to.y, durationMs}),
        text: (token: string, text: string) => send('/text', token, {text}),
        key: (token: string, key: string) => send('/key', token, {key}),
        touchElement: (token: string, ref: ElementRef) => send('/element/touch', token, ref),
        hierarchy: () => api.get<string>(`${device}/hierarchy`, {responseType: 'text', transformResponse: r => r}),
        logs: (lines: number) => api.get<{ lines: string[] }>(`${device}/logs`, {params: {lines}}).then(r => r.lines),
        appAction: (token: string, action: AppAction, {binaryId, clearData = false}: AppRequest) =>
            send(`/app/${action}`, token, {binaryId, clearData}),
    };
};
