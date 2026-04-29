package engine

import "encoding/json"

func MarshalSnapshot(snapshot WorldSnapshot) ([]byte, error) {
	return json.Marshal(snapshot)
}

func UnmarshalSnapshot(data []byte) (WorldSnapshot, error) {
	var snapshot WorldSnapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return WorldSnapshot{}, err
	}
	return snapshot, nil
}
