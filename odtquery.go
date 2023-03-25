// Copyright © 2023 Mark Summerfield. All rights reserved.
// License: GPL-3

package main

import (
	_ "embed"
	"errors"
	"fmt"
	"strings"

	"github.com/mark-summerfield/clip"
	"github.com/mark-summerfield/gong"
	"github.com/mark-summerfield/gset"
	"github.com/mark-summerfield/odt"
	"golang.org/x/exp/slices"
)

//go:embed Version.dat
var Version string

var expectedFiles = gset.New("content.xml", "META-INF/manifest.xml",
	"meta.xml", "mimetype", "styles.xml")

func main() {
	config := getConfig()
	if len(config.filenames) == 1 {
		if err := process(config.filenames[0], config, false); err != nil {
			fmt.Printf("error: %s\n", err)
		}
	} else {
		for _, filename := range config.filenames {
			fmt.Println("file:", gong.Bold(filename))
			if err := process(filename, config, true); err != nil {
				fmt.Printf("error: %s\n", err)
			}
		}
	}
}

func process(filename string, config *config, indent bool) error {
	doc, err := odt.Open(filename)
	if err != nil {
		return err
	}
	if config.verify {
		verify(doc, indent)
	}
	if config.list {
		list(doc, indent)
	}
	return nil
}

func verify(doc *odt.Odt, indent bool) {
	actualFiles := gset.New[string]()
	for name := range doc.Files {
		if expectedFiles.Contains(name) {
			actualFiles.Add(name)
		}
	}
	if indent {
		fmt.Print("  ")
	}
	if len(actualFiles) < len(expectedFiles) {
		missing := expectedFiles.Difference(actualFiles).ToSlice()
		slices.SortFunc(missing, func(a, b string) bool {
			return strings.ToLower(a) < strings.ToLower(b)
		})
		fmt.Print(gong.Italic("failed verification — missing files:"))
		for _, name := range missing {
			fmt.Print(" ", name)
		}
		fmt.Println("")
	} else {
		fmt.Println(gong.Italic("verified — contains all essential files"))
	}
}

func list(doc *odt.Odt, indent bool) {
	for name, text := range doc.Files {
		if indent {
			fmt.Print("  ")
		}
		fmt.Print(name)
		if len(text) == 0 {
			fmt.Println(" (empty)")
		} else {
			fmt.Printf(" (%s bytes)\n", gong.Commas(len(text)))
		}
	}
}

func getConfig() *config {
	parser := makeParser()
	listOpt := parser.Flag("list",
		"List each .odt file's contents. [default]")
	verifyOpt := parser.Flag("verify", "Verify each .odt file's contents.")
	if err := parser.Parse(); err != nil {
		parser.OnError(err)
	}
	config := &config{filenames: parser.Positionals,
		list: listOpt.Value(), verify: verifyOpt.Value()}
	if !(config.list || config.verify) {
		parser.OnError(errors.New("error: no action specified"))
	}
	return config
}

func makeParser() *clip.Parser {
	parser := clip.NewParserVersion(Version)
	parser.PositionalHelp = ".odt files to query"
	parser.PositionalCount = clip.OneOrMorePositionals
	parser.MustSetPositionalVarName("ODT_FILE")
	parser.LongDesc = "Queries .odt files."
	return &parser
}

type config struct {
	list      bool
	verify    bool
	filenames []string
}
