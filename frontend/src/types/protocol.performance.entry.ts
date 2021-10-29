
export default interface IProtocolPerformanceEntryData {
    ID?: number | null,
    TestProtocolID: number,
    Checkpoint: string,
    FPS: number,
    MEM: number,
    CPU: number,
    Other: string,
    Runtime: number,
    CreatedAt: Date,
    UpdatedAt: Date,
    DeletedAt: Date,
}