
const StringIsNumber = (value: any): boolean => !isNaN(Number(value));

export const ToArray = (en: any): Array<Object> => {
    return Object.keys(en).filter(StringIsNumber).map(key => ({ id: key, name: en[ key ] }));
};
