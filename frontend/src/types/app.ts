import IProject from "../project/project";
import { PlatformType } from "./platform.type.enum";

export interface IAppData {
    ID: number,
    Name: string,
    Identifier: string,
    projectID: number,
    Project: IProject,
    DefaultParameter: string,
    Platform: PlatformType,
    CreatedAt: Date,
    UpdatedAt: Date,
    DeletedAt: Date,
}

export interface IAppBinaryData {
    ID: number,
    Name: string,
    Platform: string,
    Version: string,
    AppPath: string,
    Identifier: string,
    LaunchActivity: string,
    Additional: string,
    Hash: string,
    Size: number,
    Tags: string,
    CreatedAt: Date,
    UpdatedAt: Date,
    DeletedAt: Date,
}

export const prettySize = (bytes: number): string => {
    const KB = 1024;
    const MB = KB * 1024;
    const GB = MB * 1024;

    if (bytes >= GB) {
        return `${ (bytes / GB).toFixed(2) }GB`;
    }

    if (bytes >= MB) {
        return `${ (bytes / MB).toFixed(2) }MB`;
    }

    if (bytes >= KB) {
        return `${ (bytes / KB).toFixed(2) }KB`;
    }

    return `${ bytes }B`;
};
