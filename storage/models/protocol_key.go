package models

// ProtocolKey is a value the app under test reported for a test through the
// plugin, e.g. player_id = 1234. A protocol has each key once; a key sent
// again replaces its value.
type ProtocolKey struct {
	Model
	TestProtocolID uint   `json:"testProtocolId" gorm:"uniqueIndex:idx_protocol_key"`
	Key            string `json:"key" gorm:"uniqueIndex:idx_protocol_key"`
	Value          string `json:"value"`
}
