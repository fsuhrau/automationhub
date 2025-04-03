export enum DeviceStateType {
    None = 0,
    Unknown,
    Shutdown,
    RemoteDisconnected,
    Booted,
    Locked,
    NodeDisconnected,
}

export const deviceState = (state: DeviceStateType): string => {

    switch (state) {
        case DeviceStateType.None:
            return 'None';
        case DeviceStateType.Unknown:
            return 'Unknown';
        case DeviceStateType.Shutdown:
            return 'Shutdown';
        case DeviceStateType.RemoteDisconnected:
            return 'Disconnected';
        case DeviceStateType.Booted:
            return 'Booted';
        case DeviceStateType.Locked:
            return 'Locked';
        case DeviceStateType.NodeDisconnected:
            return 'Node Disconnected';
    }
    return '';
};
