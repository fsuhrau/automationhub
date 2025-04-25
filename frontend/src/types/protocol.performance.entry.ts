
export default interface IProtocolPerformanceEntryData {
    id?: number | null,
    testProtocolId: number,
    checkpoint: string,
    fps: number,
    mem: number,
    cpu: number,
    vertexCount: number,
    triangles: number,
    other: string,
    runtime: number,
    executionTime: number,
    createdAt: Date,
    updatedAt: Date,
    deletedAt: Date,
}
