package main

import (
	"github.com/jcw/flow/flow"
	_ "github.com/jcw/flow/serial"
	_ "github.com/jcw/flow/workers"
)

func main() {
	flow.LoadFile("config.json").Run()
}