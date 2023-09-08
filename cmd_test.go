// Copyright 2022 bonzai-example Authors
// SPDX-License-Identifier: Apache-2.0

package example

import (
	"bytes"
	"log"
	"os"
	"testing"
)

// Unlike other Go projects, Bonzai commands don't really benefit from
// Go's example-based tests (which normally would be in package
// example_test). Instead, testing should be against the "pkg" library
// and, when necessary, the first-class Call functions directly. Final
// testing using the standalone executable or a composite command
// executable must always be done. Also never forget to do deployment
// testing by getting on a regular system and installing with "go
// install github.com/rwxrob/bonzai-example@latest" to ensure you have
// no errors with your versions, caching server, or dependencies.

func TestBarCmd(t *testing.T) {

	// capture the output
	buf := new(bytes.Buffer)
	log.SetFlags(0)
	log.SetOutput(buf)
	defer log.SetFlags(log.Flags())
	defer log.SetOutput(os.Stderr)

	BarCmd.Call(nil)

	t.Log(buf)
	if buf.String() != "would bar stuff\n" {
		t.Fail()
	}
}

func checkFileExists(filePath string) bool {
	_, error := os.Stat(filePath)
	//return !os.IsNotExist(err)
	return !errors.Is(error, os.ErrNotExist)
}

func TestCompileCmd(t *testing.T) {

	// capture the output
	buf := new(bytes.Buffer)
	log.SetFlags(0)
	log.SetOutput(buf)
	defer log.SetFlags(log.Flags())
	defer log.SetOutput(os.Stderr)

	CompileCmd.Call("test")

	t.Log(buf)
	if buf.String() != "Baz, suncreen song\n" {
		t.Fail()
	}

	var filePath string = "pdfs/test.pdf"

	isFileExist := checkFileExists(filePath)

	if isFileExist {
		fmt.Println("file exist")
	} else {
		t.Fail("file does not exist")
	}

}
