package protocol

import "strings"

// ExtractParams reads run parameters: key=value pairs separated by ";" as
// the run dialog sends them, or by line breaks as a run stores them and
// "Run again" sends them back. Only the first "=" separates key and value,
// so values such as URLs stay whole. Reading a stored run with ";" only made
// "MarketId=xu\nProxy=http://..." the market "xu\nProxy" and lost the proxy.
func ExtractParams(param string) map[string]string {
	env := make(map[string]string)
	params := strings.FieldsFunc(param, func(r rune) bool { return r == ';' || r == '\n' || r == '\r' })
	for _, p := range params {
		key, value, ok := strings.Cut(p, "=")
		key = strings.TrimSpace(key)
		if !ok || key == "" {
			continue
		}
		env[key] = strings.TrimSpace(value)
	}
	return env
}
