import { TestType } from "./test.type.enum";

export default interface ICreateTestData {
    Name: string,
    TestType: TestType,
    UnityAllTests: boolean,
    UnitySelectedTests: string[],
    AllDevices: boolean,
    SelectedDevices: number[],
}