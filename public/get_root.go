package public

import (
	"path/filepath"
	"runtime"
	"strings"
)

var basePath string

func init() {
	_, currentFile, _, _ := runtime.Caller(0)
	basePath = filepath.Dir(currentFile)
	rootPath := strings.Split(basePath, "/")
	basePath = strings.Join(rootPath[0:len(rootPath)-1], "/")
}
func Path(rel string) string {
	if filepath.IsAbs(rel) {
		return rel
	}
	return filepath.Join(basePath, rel)
}
