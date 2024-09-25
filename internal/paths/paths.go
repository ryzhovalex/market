package paths

import (
	"os"
	"path/filepath"
)

func MustGetCwd() string {
	r, be := os.Getwd()
	if be != nil {
		panic(be)
	}
	return r
}

func MustGetExeDir() string {
	exePath, be := os.Executable()
	if be != nil {
		panic(be)
	}
	return filepath.Dir(exePath)
}
