// The live control lease of a device: who holds it and until when.
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
