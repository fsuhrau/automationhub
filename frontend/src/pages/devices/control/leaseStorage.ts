// The lease token of a device is kept for the browser tab, so a reload keeps
// the control. Storage can be unavailable; then the lease only lives in memory.
const key = (deviceId: number): string => `device-lease-${deviceId}`;

export const loadLease = (deviceId: number): string | null => {
    try {
        return sessionStorage.getItem(key(deviceId));
    } catch {
        return null;
    }
};

export const storeLease = (deviceId: number, token: string | null): void => {
    try {
        if (token) {
            sessionStorage.setItem(key(deviceId), token);
        } else {
            sessionStorage.removeItem(key(deviceId));
        }
    } catch {
        // storage unavailable
    }
};
