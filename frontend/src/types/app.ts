import IProject from "../project/project";
import { PlatformType } from "./platform.type.enum";

export interface IAppData {
    id: number,
    name: string,
    identifier: string,
    projectId: number,
    project: IProject,
    defaultParameter: string,
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
