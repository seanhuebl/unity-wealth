package transaction

func convertResponseFloatToInt(actualResponse map[string]interface{}) map[string]interface{} {
	if data, ok := actualResponse["data"].(map[string]interface{}); ok {
		if dc, ok := data["detailed_category"].(float64); ok {
			data["detailed_category"] = int(dc)
		}
	}
	return actualResponse
}
