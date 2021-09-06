export default interface IProtocolEntryData {
    ID?: number | null,
    TestProtocolID: number,
    Source: string,
    Level: string,
    Info: string,
    CreatedAt: Date,
    UpdatedAt: Date,
    DeletedAt: Date,
}
