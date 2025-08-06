package testhelpers

import "github.com/google/uuid"

func ConvertResponseFloatToInt(actualResponse map[string]interface{}) map[string]interface{} {

	data, ok := actualResponse["data"].(map[string]interface{})
	if !ok {
		return actualResponse
	}
	if es, ok := data["effective_size"].(float64); ok {
		data["effective_size"] = int32(es)
	}
	if dc, ok := data["detailed_category"].(float64); ok {
		data["detailed_category"] = int(dc)
		actualResponse["data"] = data
		return actualResponse
	}
	if transactions, ok := data["transactions"].([]interface{}); ok {
		for _, item := range transactions {
			if tx, ok := item.(map[string]interface{}); ok {
				if dc, ok := tx["detailed_category"].(float64); ok {
					tx["detailed_category"] = int(dc)
				}
			}
		}
	}
	actualResponse["data"] = data
	return actualResponse
}

func StrPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func PrepareTxID(txID uuid.UUID, txIDRaw string) string {
	idVal := txID.String()
	if txIDRaw != "" {
		idVal = txIDRaw
	}
	return idVal
}
