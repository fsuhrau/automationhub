import IUnityTestFunctionData from './unity.test.function';
import { UnityTestCategory } from "./unity.test.category.type.enum";

export default interface ITestConfigUnityData {
    id?: number | null,
    testConfigId: number,
    testCategoryType: UnityTestCategory,
    testFunctions: Array<IUnityTestFunctionData>,
    categories: string,
    createdAt: Date,
    updatedAt: Date,
    deletedAt: Date,
}
