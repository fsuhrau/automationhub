// A key the app under test reported for a test through the plugin, e.g.
// player_id = 1234. A protocol has each key once; the last value sent wins.
export default interface IProtocolKeyData {
    id?: number,
    testProtocolId: number,
    key: string,
    value: string,
    createdAt?: Date,
}

// payload of the test_protocol_<id>_keys event
export interface ProtocolKeyPayload {
    testRunId: number,
    testProtocolId: number,
    key: IProtocolKeyData,
}
