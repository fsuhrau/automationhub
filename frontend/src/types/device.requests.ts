// What a person can set on a device: the alias and the custom parameters
// (replaced as a whole).
export interface DeviceUpdate {
    alias: string,
    customParameter: { id?: number, key: string, value: string }[],
}

// A test started on one device by hand.
export interface DeviceTestRequest {
    testName: string,
    env: string,
}
