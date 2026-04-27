package main

import (
	"fmt"
	"os"

	"github.com/gkthiruvathukal/luc-cs/internal/tui"
)

func main() {
	if err := tui.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
