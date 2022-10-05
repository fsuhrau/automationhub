import { IdName, ToArray } from '../helper/enum_to_array';

export enum PlatformType {
    iOS,
    Android,
    Mac,
    Windows,
    Linux,
    Web,
    Editor,
}

export const getPlatformTypes = (): Array<IdName> => {
    return ToArray(PlatformType);
};

export const getPlatformName = (t: PlatformType): string => {
    return PlatformType[t];
};
