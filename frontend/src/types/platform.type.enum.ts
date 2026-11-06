import { IdName, ToArray } from '../helper/enum_to_array';

export enum PlatformType {
    iOS,
    Android,
    Mac,
    Windows,
    Linux,
    Web,
    Editor,
    IOSSimulator,
}

// platforms an app can target; the simulator is a device platform only
export const getPlatformTypes = (): Array<IdName> => {
    return ToArray(PlatformType).filter(p => Number(p.id) !== PlatformType.IOSSimulator);
};

export const getPlatformName = (t: PlatformType): string => {
    return PlatformType[t];
};
