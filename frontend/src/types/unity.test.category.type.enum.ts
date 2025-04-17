import {IdName, ToArray} from '../helper/enum_to_array';

export enum UnityTestCategory {
    RunAllTests,
    RunAllOfCategory,
    RunSelectedTestsOnly,
}

export const getUnityTestCategoryTypes = (): Array<IdName> => {
    return ToArray(UnityTestCategory);
};

export const getUnityTestCategoryName = (type: UnityTestCategory): string => {
    switch (type) {
        case UnityTestCategory.RunAllTests:
            return 'All tests';
        case UnityTestCategory.RunAllOfCategory:
            return 'All of category';
        case UnityTestCategory.RunSelectedTestsOnly:
            return 'Selected tests only'
    }
    return ''
};
