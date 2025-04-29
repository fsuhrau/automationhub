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
    additional: string,
    hash: string,
    size: number,
    tags: string,
    createdAt: Date,
    updatedAt: Date,
    deletedAt: Date,
}
