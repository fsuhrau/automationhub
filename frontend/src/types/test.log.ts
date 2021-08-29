import ILogEntryData from "./log.entry";

export default interface ITestLogData {
    id?: number | null,
    TestRunID: number,
    StartedAt: number,
    EndedAt?: number,
    Entries: ILogEntryData[]
}