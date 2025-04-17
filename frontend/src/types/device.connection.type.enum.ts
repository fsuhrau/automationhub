import { ToArray } from '../helper/enum_to_array';

export enum DeviceConnectionType {
    USB,
    Remote,
    HubNode,
}
export const getConnectionTypes = (): Array<Object> => {
    return ToArray(DeviceConnectionType);
};

export const getConnectionTypeName = (type: DeviceConnectionType): string => {
    return DeviceConnectionType[type];
};
