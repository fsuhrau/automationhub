import { IdName, ToArray } from '../helper/enum_to_array';

export enum TestType {
    Unity,
    Cocos,
    Serenity,
    Scenario,
}

export const getTestTypes = (): Array<IdName> => {
    return ToArray(TestType);
};

export const getTestTypeName = (type: TestType): string => {
    return TestType[type];
};
