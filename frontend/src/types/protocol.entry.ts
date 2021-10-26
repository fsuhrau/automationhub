export default interface IProtocolEntryData {
    ID?: number | null,
    TestProtocolID: number,
    Source: string,
    Level: string,
    Message: string,
    Data: string,
    Runtime: number,
    CreatedAt: Date,
    UpdatedAt: Date,
    DeletedAt: Date,
}
