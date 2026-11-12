export default interface IProtocolEntryData {
    id?: number | null,
    testProtocolId: number,
    source: string,
    level: string,
    message: string,
    data: string,
    runtime: number,
    // when the entry was written, on the master's clock
    timestamp?: Date,
    createdAt: Date,
    updatedAt: Date,
    deletedAt: Date,
}
