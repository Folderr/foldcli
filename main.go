// go:build go1.20
/*
Copyright © 2023 Folderr <contact@folderr.net>
*/
package main

import (
	"github.com/Folderr/foldcli/cmd"
	_ "github.com/Folderr/foldcli/cmd/init"
)

func main() {
	cmd.Execute()
}
