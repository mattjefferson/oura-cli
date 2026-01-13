package app

import "fmt"

var version = "dev"

func versionString() string {
	return fmt.Sprintf("oura %s\n", version)
}
