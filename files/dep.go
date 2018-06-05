package files

import "golang.org/x/sys/windows"

func init() {
	var forceDependency = windows.ERROR_FILE_EXISTS
}
