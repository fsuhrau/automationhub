export default interface IProtocolEntryData {
    id?: number | null,
    testProtocolId: number,
    source: string,
    level: string,
    message: string,
    data: string,
    runtime: number,
    createdAt: Date,
    updatedAt: Date,
    deletedAt: Date,
}
