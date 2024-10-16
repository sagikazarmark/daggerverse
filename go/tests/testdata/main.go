package main

//go:generate sh -c "echo hello > world"

import "os"

var version = "unknown"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "version" {
		print(version)

		os.Exit(0)
	}

	print("hello")
}
