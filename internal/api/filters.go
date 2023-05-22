package api

func filterkeypair(key, value string, filterterms map[string]string) map[string]string {
	switch key {
	case "dept":
		filterterms[key] = value
	case "name":
		filterterms[key] = value
	default:
	}
	return filterterms
}
