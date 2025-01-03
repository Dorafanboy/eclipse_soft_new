package format

func FormatStatus(enabled bool) string {
	if enabled {
		return "✅ Enabled"
	}
	return "❌ Disabled"
}
