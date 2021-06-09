package main

import (
	"fmt"
	"os"

	"github.com/vigo/putio-cli/cli"
)

func main() {
	cmd := cli.NewApplication()
	if err := cmd.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
