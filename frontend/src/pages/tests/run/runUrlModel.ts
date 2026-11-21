// The query of the replay link of a run: where the replay starts.
export const replayQuery = (time?: Date, deviceId?: number | null): string => {
    const params = new URLSearchParams();
    if (time) params.set('t', String(new Date(time).valueOf()));
    if (deviceId) params.set('device', String(deviceId));
    const query = params.toString();
    return query ? `?${query}` : '';
};
