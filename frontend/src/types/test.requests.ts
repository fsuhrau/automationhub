import IAppFunctionData from './app.function';
import {TestExecutionType} from './test.execution.type.enum';
import {UnityTestCategory} from './unity.test.category.type.enum';

// the edit page's settings of a test, sent to create or update it
export interface UpdateTestData {
    name: string,
    executionType: TestExecutionType,
    allDevices: boolean,
    devices: number[],
    unityTestCategoryType: UnityTestCategory,
    categories: string,
    // the picked Unity tests; stored ones carry more fields, which are ignored
    testFunctions: IAppFunctionData[],
    recordSession?: boolean,
}

// what starting a run of a test sends
export interface RunTestData {
    appBinaryId: number | null,
    startUrl: string | null,
    params: string,
}

// where the tests live: the project and the app
export interface TestScope {
    projectId: string,
    appId: number,
}
