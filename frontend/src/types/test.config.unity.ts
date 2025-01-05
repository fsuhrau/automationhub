import IUnityTestFunctionData from './unity.test.function';
import { UnityTestCategory } from "./unity.test.category.type.enum";

export default interface ITestConfigUnityData {
    ID?: number | null,
    TestConfigID: number,
    UnityTestCategoryType: UnityTestCategory,
    UnityTestFunctions: Array<IUnityTestFunctionData>,
    Categories: string,
    CreatedAt: Date,
    UpdatedAt: Date,
    DeletedAt: Date,
}
