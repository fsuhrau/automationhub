
const StringIsNumber = (value: any): boolean => !isNaN(Number(value));

export type IdName = {
    id: string,
    name: string,
}

export const ToArray = (en: any): Array<IdName> => {
    return Object.keys(en).filter(StringIsNumber).map(key => ({ id: key, name: en[ key ] } as IdName));
};
