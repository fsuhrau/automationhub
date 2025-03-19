
export default interface IProtocolPerformanceEntryData {
    ID?: number | null,
    TestProtocolID: number,
    Checkpoint: string,
    FPS: number,
    MEM: number,
    CPU: number,
    VertexCount: number,
    Triangles: number,
    Other: string,
    Runtime: number,
    ExecutionTime: number,
    CreatedAt: Date,
    UpdatedAt: Date,
    DeletedAt: Date,
}
