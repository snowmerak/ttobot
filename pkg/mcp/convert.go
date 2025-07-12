package mcp

import (
	"encoding/json"
	"fmt"
)

func ConvertViaJSON(from any, to any) error {
	data, err := json.Marshal(from)
	if err != nil {
		return fmt.Errorf("failed to marshal from type %T: %w", from, err)
	}
	err = json.Unmarshal(data, to)
	if err != nil {
		return fmt.Errorf("failed to unmarshal to type %T: %w", to, err)
	}
	return nil
}
