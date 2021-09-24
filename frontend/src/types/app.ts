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
    AppFunctions: IAppFunctionData[],
    CreatedAt: Date,
    UpdatedAt: Date,
    DeletedAt: Date,
}
