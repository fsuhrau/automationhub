import { ToArray } from '../helper/enum_to_array';

export enum TestExecutionType {
    Concurrent,
    Simultaneously,
}

export const getExecutionTypes = (): Array<Object> => {
    return ToArray(TestExecutionType);
};

export const getTestExecutionName = (type: TestExecutionType): string => {
    return TestExecutionType[type];
};
