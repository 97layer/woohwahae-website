package main

import (
	"flag"
	"log"
)

func parseArgs(cmd *flag.FlagSet, args []string) {
	if err := cmd.Parse(args); err != nil {
		log.Fatal(err)
	}
}
