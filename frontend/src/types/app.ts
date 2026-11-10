import {BundleInfo} from "./bundle.info";
import IProject from "../project/project";
import {PlatformType} from "./platform.type.enum";

export type AppParameterType = 'string' | 'option'

export interface AppParameterString {
    type: AppParameterType
    defaultValue: string
}

export interface AppParameterOption {
    type: AppParameterType
    options: string[]
    defaultValue: string
}

export type Parameter = AppParameterString | AppParameterOption;

export interface AppParameter {
    id: number
    name: string
    type: Parameter
}

export interface IAppData {
    id: number,
    name: string,
    identifier: string,
    projectId: number,
    project: IProject,
    parameter: AppParameter[],
    platform: PlatformType,
    createdAt: Date,
    updatedAt: Date,
    deletedAt: Date,
}

export interface IAppBinaryData {
    id: number,
    name: string,
    platform: string,
    version: string,
    appPath: string,
    identifier: string,
    launchActivity: string,
    // as the master stores them
    android?: { LaunchActivity: string },
    executable?: { Executable: string },
    additional: string,
    hash: string,
    size: number,
    tags: string,
    // how many test runs used the bundle (in the bundle list)
    runCount?: number,
    // all read from the file; absent for bundles uploaded before the master
    // read it, and only a summary in the bundle list
    info?: BundleInfo,
    createdAt: Date,
    updatedAt: Date,
    deletedAt: Date,
}

// a test run that used a bundle
export interface BundleRun {
    id: number,
    testId: number,
    testName: string,
    createdAt: string,
    // running, passed, failed, unstable, interrupted, not_run or empty
    status: string,
    passed: number,
    failed: number,
    unstable: number,
    open: number,
}

export interface IAppBundleDetail extends IAppBinaryData {
    runCount: number,
    runs: BundleRun[],
}
