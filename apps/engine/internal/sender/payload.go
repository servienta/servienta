package sender

func Str(m map[string]any, k string) string {
	if m == nil {
		return ""
	}
	s, _ := m[k].(string)
	return s
}

func StrOr(m map[string]any, k, def string) string {
	if s := Str(m, k); s != "" {
		return s
	}
	return def
}

func Int(m map[string]any, k string) int {
	if m == nil {
		return 0
	}
	if f, ok := m[k].(float64); ok {
		return int(f)
	}
	return 0
}
