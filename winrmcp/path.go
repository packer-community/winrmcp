package winrmcp

import "strings"

func winPath(path string) string {
	if len(path) == 0 {
		return path
	}

	path = strings.Trim(path, "'\"")
	return strings.Replace(path, "/", "\\", -1)
}
