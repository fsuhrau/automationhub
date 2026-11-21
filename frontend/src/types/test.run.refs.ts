// What identifies a test, one of its runs and one protocol of a run in the
// API urls of the test run service.
export interface TestRef {
    projectId: string,
    appId: number,
    testId: string | number,
}

export interface RunRef extends TestRef {
    runId: string | number,
}

export interface ProtocolRef extends RunRef {
    protocolId: string | number,
}
