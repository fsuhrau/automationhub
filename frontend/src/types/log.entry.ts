export default interface ILogEntryData {
    ID?: number | null,
    TestLogID: number,
    Timestamp: number,
    Source: string,
    Info: string,
}
