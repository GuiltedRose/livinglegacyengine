package engine

import "encoding/json"

// MarshalSnapshot encodes a snapshot with encoding/json.
func MarshalSnapshot(snapshot WorldSnapshot) ([]byte, error) {
	return json.Marshal(snapshot)
}

// UnmarshalSnapshot decodes snapshot JSON without restoring it into a world.
func UnmarshalSnapshot(data []byte) (WorldSnapshot, error) {
	var snapshot WorldSnapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return WorldSnapshot{}, err
	}
	return snapshot, nil
}
