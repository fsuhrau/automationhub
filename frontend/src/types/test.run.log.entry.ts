export default interface ITesRunLogEntryData {
    ID?: number,
    TestRunID: number,
    Level: string,
    Log: string,
    CreatedAt: Date,
    UpdatedAt: Date,
    DeletedAt: Date,
}
