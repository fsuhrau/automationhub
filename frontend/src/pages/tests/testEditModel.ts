// The test edit form without React: its state from a test, what keeps it
// from saving (the same rules the master checks), unsaved changes, and the
// Unity test functions it picks.
import IAppFunctionData from '../../types/app.function';
import {TestExecutionType} from '../../types/test.execution.type.enum';
import {UnityTestCategory} from '../../types/unity.test.category.type.enum';

export interface TestForm {
    name: string,
    executionType: TestExecutionType,
    allDevices: boolean,
    devices: number[],
    mode: UnityTestCategory,
    categories: string[],
    functions: IAppFunctionData[],
}

export interface TestLike {
    name: string,
    testConfig: {
        executionType: TestExecutionType,
        allDevices: boolean,
        devices?: { deviceId: number }[] | null,
        unity?: { testCategoryType: UnityTestCategory, categories?: string, testFunctions?: IAppFunctionData[] | null } | null,
    },
}

export const splitCategories = (text: string | undefined | null): string[] =>
    Array.from(new Set((text ?? '').split(',').map(c => c.trim()).filter(c => c.length > 0)));

export const formOf = (test: TestLike): TestForm => ({
    name: test.name,
    executionType: test.testConfig.executionType,
    allDevices: test.testConfig.allDevices,
    devices: (test.testConfig.devices ?? []).map(d => d.deviceId),
    mode: test.testConfig.unity?.testCategoryType ?? UnityTestCategory.RunAllTests,
    categories: splitCategories(test.testConfig.unity?.categories),
    functions: (test.testConfig.unity?.testFunctions ?? []).map(f => ({assembly: f.assembly, class: f.class, method: f.method})),
});

// newForm is where a new test starts: all Unity tests, on all devices,
// spread over the free ones
export const newForm = (): TestForm => ({
    name: '',
    executionType: TestExecutionType.Concurrent,
    allDevices: true,
    devices: [],
    mode: UnityTestCategory.RunAllTests,
    categories: [],
    functions: [],
});

export const functionKey = (f: IAppFunctionData): string => `${f.assembly}|${f.class}|${f.method}`;

// addFunctions adds functions that aren't picked yet, keeping the order
export const addFunctions = (picked: IAppFunctionData[], add: IAppFunctionData[]): IAppFunctionData[] => {
    const keys = new Set(picked.map(functionKey));
    const out = [...picked];
    add.forEach(f => {
        if (!keys.has(functionKey(f))) {
            keys.add(functionKey(f));
            out.push(f);
        }
    });
    return out;
};

export interface FormErrors {
    name?: string,
    devices?: string,
    tests?: string,
    categories?: string,
}

export const validate = (f: TestForm, unity: boolean): FormErrors => {
    const errors: FormErrors = {};
    if (f.name.trim() === '') errors.name = 'The test needs a name.';
    if (!f.allDevices && f.devices.length === 0) errors.devices = 'Pick at least one device, or run on all devices.';
    if (unity && f.mode === UnityTestCategory.RunSelectedTestsOnly && f.functions.length === 0) errors.tests = 'Pick at least one test.';
    if (unity && f.mode === UnityTestCategory.RunAllOfCategory && f.categories.length === 0) errors.categories = 'Name at least one category.';
    return errors;
};

const comparable = (f: TestForm) => JSON.stringify([
    f.name.trim(), f.executionType, f.allDevices, f.allDevices ? [] : [...f.devices].sort((a, b) => a - b),
    f.mode, [...f.categories].sort(), f.functions.map(functionKey).sort(),
]);

export const isDirty = (a: TestForm, b: TestForm): boolean => comparable(a) !== comparable(b);

export const toRequest = (f: TestForm) => ({
    name: f.name.trim(),
    executionType: f.executionType,
    allDevices: f.allDevices,
    devices: f.allDevices ? [] : f.devices,
    unityTestCategoryType: f.mode,
    categories: f.categories.join(','),
    testFunctions: f.functions,
});

// shortClass and shortMethod make Unity names readable: the class without
// its namespace, the method without its return type
export const shortClass = (cls: string): string => cls.split('.').pop() ?? cls;
export const shortMethod = (method: string): string => method.trim().split(' ').pop() ?? method;

// groupByClass groups functions by their class, in the order classes appear
export const groupByClass = (functions: IAppFunctionData[]): [string, IAppFunctionData[]][] => {
    const groups = new Map<string, IAppFunctionData[]>();
    functions.forEach(f => groups.set(f.class, [...(groups.get(f.class) ?? []), f]));
    return Array.from(groups);
};

export const matchesFunction = (f: IAppFunctionData, text: string): boolean => {
    const needle = text.trim().toLowerCase();
    return needle === '' || [f.method, f.class, f.assembly].some(v => (v ?? '').toLowerCase().includes(needle));
};

// parsePastedTests reads the test list Unity copies (a JSON array of
// {assembly, class, method})
export const parsePastedTests = (text: string): { functions: IAppFunctionData[] } | { error: string } => {
    let data: unknown;
    try {
        data = JSON.parse(text);
    } catch {
        return {error: "That isn't JSON. Copy the test list from Unity (Tools › Automation Hub) and paste it here."};
    }
    if (!Array.isArray(data)) return {error: 'Expected a list of tests.'};
    const functions = data.filter((d): d is IAppFunctionData =>
        !!d && typeof d === 'object' && typeof (d as IAppFunctionData).method === 'string' && typeof (d as IAppFunctionData).class === 'string')
        .map(d => ({assembly: d.assembly ?? '', class: d.class, method: d.method}));
    if (functions.length === 0) return {error: 'The list has no tests with a class and a method.'};
    return {functions};
};
