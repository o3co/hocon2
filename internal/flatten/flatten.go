package flatten

import "fmt"

// Flatten converts a nested map[string]any to a flat map[string]string
// with dot-separated keys.
func Flatten(m map[string]any) map[string]string {
	result := make(map[string]string)
	flattenRecursive(m, "", result)
	return result
}

func flattenRecursive(m map[string]any, prefix string, result map[string]string) {
	for k, v := range m {
		key := k
		if prefix != "" {
			key = prefix + "." + k
		}

		switch val := v.(type) {
		case map[string]any:
			if len(val) == 0 {
				continue
			}
			flattenRecursive(val, key, result)
		case []any:
			if len(val) == 0 {
				continue
			}
			for i, item := range val {
				indexKey := fmt.Sprintf("%s.%d", key, i)
				switch nested := item.(type) {
				case map[string]any:
					flattenRecursive(nested, indexKey, result)
				default:
					result[indexKey] = fmt.Sprintf("%v", item)
				}
			}
		case nil:
			result[key] = ""
		default:
			result[key] = fmt.Sprintf("%v", val)
		}
	}
}
