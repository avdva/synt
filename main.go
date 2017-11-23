// Copyright 2017 Aleksandr Demakin. All rights reserved.

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/avdva/synt/linter"

	log "github.com/sirupsen/logrus"
)

func main() {
	var reports []linter.Report
	flag.Parse()
	toParse := flag.Args()
	if len(toParse) == 0 {
		toParse = []string{"."}
	}
	for _, name := range toParse {
		if fi, err := os.Stat(name); err == nil && fi.IsDir() {
			if rep, err := linter.DoDir(name); err == nil {
				reports = append(reports, rep...)
			} else {
				log.Warnf("'%s' parse error: %v", name, err)
			}
		} else {
			log.Warnf("not a directory: %s", name)
		}
	}
	for _, report := range reports {
		fmt.Printf("%s: %s\n", report.Err, report.Location)
	}
	if len(reports) > 0 {
		os.Exit(1)
	}
}
