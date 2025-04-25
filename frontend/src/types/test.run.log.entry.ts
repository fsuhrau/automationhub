export default interface ITesRunLogEntryData {
    id?: number,
    testRunId: number,
    level: string,
    log: string,
    createdAt: Date,
    updatedAt: Date,
    deletedAt: Date,
}
