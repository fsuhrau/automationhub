const StringIsNumber = (value: string): boolean => !isNaN(Number(value));

export type IdName = {
    id: string,
    name: string,
}

// the numeric members of a TypeScript enum as id/name pairs
export const ToArray = (en: Record<string, string | number>): Array<IdName> => {
    return Object.keys(en).filter(StringIsNumber).map(key => ({ id: key, name: String(en[ key ]) } as IdName));
};
