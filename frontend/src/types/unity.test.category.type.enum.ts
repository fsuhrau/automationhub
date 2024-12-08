import { IdName, ToArray } from '../helper/enum_to_array';

export enum UnityTestCategory {
    RunAllTests,
    RunAllOfCategory,
    RunSelectedTestsOnly,
}

export const getUnityTestCategoryTypes = (): Array<IdName> => {
    return ToArray(UnityTestCategory);
};

export const getUnityTestCategoryName = (type: UnityTestCategory): string => {
    return UnityTestCategory[type];
};
