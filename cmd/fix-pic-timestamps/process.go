package main

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/chtisgit/fix-pic-timestamps/internal/config"
)

const readDirN = 32

var errQuit = errors.New("quit")

type fixer struct {
	opts  *config.Options
	atime time.Time
	tz    *time.Location
}

func (f *fixer) process(path string) error {
	statInfo, err := os.Stat(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: cannot stat %s\n", path)
		return err
	}

	if !statInfo.IsDir() {
		return f.processFile(path, statInfo)
	}

	dirPaths := []string{path}

	for idx := 0; idx != len(dirPaths); idx++ {

		path := dirPaths[idx]
		if f.opts.Verbose {
			fmt.Fprintf(os.Stderr, "Descending into %s\n", path)
		}

		more, err := f.processDirectory(path)
		if err == errQuit || (err != nil && !f.opts.IgnoreErrs) {
			return err
		}

		dirPaths = append(dirPaths, more...)
	}

	return nil
}

func (f *fixer) processDirectory(path string) ([]string, error) {
	dir, err := os.Open(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: cannot open %s\n", path)
		return nil, err
	}
	defer dir.Close()

	dirPaths := []string{}
	for {
		entries, err := dir.ReadDir(readDirN)
		if err != nil {
			if err == io.EOF {
				break
			}

			fmt.Fprintf(os.Stderr, "error: cannot read directory %s\n", path)
			return nil, err
		}

		for i := range entries {
			info, err := entries[i].Info()
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: cannot stat %s\n", filepath.Join(path, info.Name()))
				continue
			}

			filePath := filepath.Join(path, info.Name())
			if info.IsDir() {
				if f.opts.Recursive {
					dirPaths = append(dirPaths, filePath)
				}

				continue
			}

			err = f.processFile(filePath, info)
			if err != nil {
				if err == errQuit || !f.opts.IgnoreErrs {
					return nil, err
				}
			}
		}
	}

	return dirPaths, nil
}

var timestampRegex = regexp.MustCompile("20[0-9]{2}(0[0-9]|10|11|12)[0-3][0-9]_[012][0-9][0-5][0-9][0-5][0-9]")

const timestampLayout = "20060102_150405"

func (f *fixer) processFile(filePath string, info fs.FileInfo) error {
	part := timestampRegex.FindString(info.Name())
	if part == "" {
		fmt.Fprintf(os.Stderr, "error: cannot find timestamp in %s\n", info.Name())
		return errors.New("parse error")
	}

	t, err := time.ParseInLocation(timestampLayout, part, f.tz)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: cannot parse timestamp in %s\n", info.Name())
		return err
	}

	return f.adjustTime(filePath, info, t)
}

func (f *fixer) adjustTime(filePath string, info fs.FileInfo, mtime time.Time) error {
	diff := int(mtime.Sub(info.ModTime()) / time.Millisecond)
	if diff < 0 {
		diff = -diff
	}
	if diff <= f.opts.MaxAllowedOffset {
		if f.opts.Verbose {
			fmt.Fprintf(os.Stderr, "%s: timestamp is within tolerances\n", filePath)
		}

		return nil
	}

	const timestampLayout = "2006-01-02 15:04:05"

	if f.opts.Verbose || f.opts.Interactive {
		fmt.Fprintf(os.Stderr, "%s: timestamp change %s -> %s\n", filePath, info.ModTime().Format(timestampLayout), mtime.Format(timestampLayout))
	}

	if f.opts.DryRun {
		return nil
	}

	if f.opts.Interactive {
		switch f.ask() {
		case InteractiveYes:
			break
		case InteractiveNo:
			return nil
		default:
			return errQuit
		}
	}

	err := os.Chtimes(filePath, f.atime, mtime)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err.Error())
	}

	return err
}

const (
	InteractiveYes int = iota + 1
	InteractiveNo
	InteractiveQuit
)

func (f *fixer) ask() int {
	for {
		fmt.Fprintf(os.Stderr, "Accept the change? [Y]es/[N]o/[Q]uit: ")

		var s string
		fmt.Scanf("%s", &s)
		if s == "" {
			continue
		}

		switch s[0] {
		case 'Y', 'y':
			return InteractiveYes
		case 'N', 'n':
			return InteractiveNo
		case 'Q', 'q':
			return InteractiveQuit
		default:
			continue
		}
	}

}
