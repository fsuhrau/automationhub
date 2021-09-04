import ILogEntryData from './log.entry';

export default interface ITestLogData {
    ID?: number | null,
    TestRunID: number,
    StartedAt: number,
    EndedAt?: number,
    Entries: ILogEntryData[]
}
