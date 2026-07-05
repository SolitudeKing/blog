package service

import "os"

func ensureParentDir(path string) error {
	return os.MkdirAll(parentDir(path), 0o755)
}

func createFile(path string) (*os.File, error) {
	return os.Create(path)
}

func openFile(path string) (*os.File, error) {
	return os.Open(path)
}

func removeFile(path string) error {
	return os.Remove(path)
}

func parentDir(path string) string {
	for index := len(path) - 1; index >= 0; index-- {
		if path[index] == '\\' || path[index] == '/' {
			return path[:index]
		}
	}
	return "."
}
