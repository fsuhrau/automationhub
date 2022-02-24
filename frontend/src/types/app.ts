import IAppFunctionData from './app.function';

export default interface IAppData {
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
    AppFunctions: IAppFunctionData[],
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
