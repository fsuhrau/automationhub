// A stretch of video of a device's screen during a run. "mp4" files play in a
// video element; "mjpeg" files are JPEG frames, listed in indexFile with one
// "ms offset length" line per frame (ms since startedAt).
export default interface ITestRunRecordingData {
    id: number,
    testRunId: number,
    deviceId: number,
    codec: 'mp4' | 'mjpeg',
    fileName: string,
    indexFile?: string,
    width: number,
    height: number,
    startedAt: string,
    endedAt: string | null,
    size: number,
}
