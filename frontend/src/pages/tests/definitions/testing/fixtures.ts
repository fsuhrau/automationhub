import IAppFunctionData from '../../../../types/app.function';
import {AppParameter, IAppBinaryData, IAppData} from '../../../../types/app';
import IDeviceData from '../../../../types/device';
import ITestData, {IRunSummary} from '../../../../types/test';
import {ITestRunListEntry} from '../../../../services/test.run.service';
import {PlatformType} from '../../../../types/platform.type.enum';
import {TestExecutionType} from '../../../../types/test.execution.type.enum';
import {TestType} from '../../../../types/test.type.enum';
import {UnityTestCategory} from '../../../../types/unity.test.category.type.enum';
import {DeviceStateType} from '../../../../types/deviceStateType';
import IProject from '../../../../project/project';

const DATE = new Date('2026-01-02T03:04:05Z');

export const aProject = (): IProject => ({id: 1, identifier: 'proj', name: 'Proj', companyId: 1, accessTokens: [], apps: []});

export const anApp = (over: Partial<IAppData> = {}): IAppData => ({
    id: 1, name: 'Hero', identifier: 'hero', projectId: 1, project: aProject(), parameter: [],
    platform: PlatformType.Editor, createdAt: DATE, updatedAt: DATE, deletedAt: DATE, ...over,
});

export const aParameter = (over: Partial<AppParameter> = {}): AppParameter =>
    ({id: 1, name: 'level', type: {type: 'string', defaultValue: ''}, ...over});

export const aTest = (over: Partial<ITestData> = {}, config: Partial<ITestData['testConfig']> = {}): ITestData => ({
    id: 7, appId: 1, name: 'Smoke', createdAt: DATE, updatedAt: DATE, deletedAt: DATE,
    testConfig: {
        testId: 7, executionType: TestExecutionType.Concurrent, type: TestType.Unity, allDevices: true, devices: [],
        unity: {
            testConfigId: 1, testCategoryType: UnityTestCategory.RunAllTests, testFunctions: [], categories: '',
            createdAt: DATE, updatedAt: DATE, deletedAt: DATE,
        },
        createdAt: DATE, updatedAt: DATE, deletedAt: DATE, ...config,
    },
    ...over,
});

export const aSummary = (over: Partial<IRunSummary> = {}): IRunSummary =>
    ({id: 11, createdAt: DATE, state: 'finished', status: 'passed', passed: 3, failed: 0, unstable: 0, open: 0, ...over});

export const aRunEntry = (over: Partial<ITestRunListEntry> = {}): ITestRunListEntry =>
    ({...aSummary(), parameter: '', bundle: null, endedAt: new Date('2026-01-02T03:05:05Z'), ...over});

export const aDevice = (over: Partial<IDeviceData> = {}): IDeviceData => ({
    id: 5, nodeId: 1, companyId: 1, deviceIdentifier: 'dev', deviceType: 0, platformType: PlatformType.Android, name: 'Pixel', alias: '',
    os: 'Android', osVersion: '14', status: DeviceStateType.Booted, isLocked: false, manager: '', isAcknowledged: true,
    connectionType: 0, connectionParameter: {} as IDeviceData['connectionParameter'], customParameter: [], deviceParameter: [], ...over,
});

export const aBundle = (over: Partial<IAppBinaryData> = {}): IAppBinaryData => ({
    id: 3, name: 'hero.apk', platform: 'android', version: '1.2.0', appPath: '', identifier: 'hero', launchActivity: '', additional: '',
    hash: 'h', size: 1, tags: '', createdAt: DATE, ...over,
} as IAppBinaryData);

export const aFunction = (method: string, cls = 'Ns.Fight'): IAppFunctionData => ({assembly: 'Game', class: cls, method});
