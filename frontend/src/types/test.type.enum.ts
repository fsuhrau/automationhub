import { ToArray } from "../helper/enum_to_array";

export enum TestType {
    Unity,
    Cocos,
    Serenity,
    Scenario,
}

export const getTestTypes = (): Array<Object> => {
    return ToArray(TestType);
}

export const getTestTypeName = (type: TestType): string => {
    return TestType[type];
};
