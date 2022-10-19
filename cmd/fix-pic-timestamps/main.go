package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/chtisgit/fix-pic-timestamps/internal/config"
)

func main() {
	var opts config.Options

	flag.StringVar(&opts.TimeZone, "tz", "", "set timezone to use (if not set, uses system timezone)")
	flag.BoolVar(&opts.DryRun, "n", false, "dry run (enables verbose)")
	flag.BoolVar(&opts.Verbose, "v", false, "verbose")
	flag.BoolVar(&opts.Recursive, "r", false, "recursively process directories")
	flag.BoolVar(&opts.IgnoreErrs, "ignore-errors", false, "keep on running when errors occur")
	flag.BoolVar(&opts.Interactive, "i", false, "interactive mode")
	flag.IntVar(&opts.MaxAllowedOffset, "off", 2000, "file timestamps can be off by this many ms without being touched by this tool")
	flag.Parse()

	if opts.DryRun {
		opts.Verbose = true
	}

	if flag.NArg() == 0 {
		flag.Usage()
		return
	}

	now := time.Now()

	tz := now.Location()
	if opts.TimeZone != "" {
		var err error
		tz, err = time.LoadLocation(opts.TimeZone)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: cannot identify location %s\n", opts.TimeZone)
			return
		}
	} else if opts.Verbose {
		fmt.Fprintf(os.Stderr, "Using timezone %s\n", tz.String())
	}

	f := &fixer{
		opts:  &opts,
		atime: now,
		tz:    tz,
	}

	args := flag.Args()
	for _, path := range args {
		err := f.process(path)
		if err != nil {
			if err == errQuit || !f.opts.IgnoreErrs {
				fmt.Fprintf(os.Stderr, "Terminating.\n")
				return
			}
		}
	}
}
