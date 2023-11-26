package util

import "encoding/json"

func StructToMap(structure interface{}) (transformedMap map[string]interface{}, err error) {
	var structJSON []byte
	structJSON, err = json.Marshal(structure)
	if err != nil {
		return
	}
	err = json.Unmarshal(structJSON, &transformedMap)
	if err != nil {
		return
	}
	return
}
