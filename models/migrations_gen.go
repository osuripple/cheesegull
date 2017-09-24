// +build ignore

package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

const fileHeader = `// THIS FILE HAS BEEN AUTOMATICALLY GENERATED
// To re-generate it, run "go generate" in the models folder.

package models

var migrations = [...]string{
`

func main() {
	// ReadDir gets all the files in the directory and then sorts them
	// alphabetically - thus we can be sure 0000 will come first and 0001 will
	// come afterwards.
	files, err := ioutil.ReadDir("migrations")
	check(err)

	out, err := os.Create("migrations.go")
	check(err)

	_, err = out.WriteString(fileHeader)
	check(err)

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".sql") || file.IsDir() {
			continue
		}
		f, err := os.Open("migrations/" + file.Name())
		check(err)

		out.WriteString("\t`")
		_, err = io.Copy(out, f)
		check(err)
		out.WriteString("`,\n")

		f.Close()
	}

	_, err = out.WriteString("}\n")
	check(err)

	check(out.Close())
}

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stdout, err)
		os.Exit(1)
	}
}
