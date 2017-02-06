// Copyright (c) 2017 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mailru/easyjson/parser"
	// Reference the gen package to be friendly to vendoring tools,
	// as it is an indirect dependency.
	// (The temporary bootstrapping code uses it.)
	"github.com/mailru/easyjson/bootstrap"
	_ "github.com/mailru/easyjson/gen"
)

func generate(fname string) error {
	fInfo, err := os.Stat(fname)
	if err != nil {
		return err
	}

	p := parser.Parser{AllStructs: true}
	if err := p.Parse(fname, fInfo.IsDir()); err != nil {
		return fmt.Errorf("Error parsing %v: %v", fname, err)
	}

	var outName string
	if fInfo.IsDir() {
		outName = filepath.Join(fname, p.PkgName+"_easyjson.go")
	} else {
		s := strings.TrimSuffix(fname, ".go")
		if s == fname {
			return errors.New("Filename must end in '.go'")
		}
		outName = s + "_easyjson.go"
	}

	g := bootstrap.Generator{
		BuildTags:       "",
		PkgPath:         p.PkgPath,
		PkgName:         p.PkgName,
		Types:           p.StructNames,
		SnakeCase:       false,
		NoStdMarshalers: false,
		OmitEmpty:       false,
		LeaveTemps:      false,
		OutName:         outName,
		StubsOnly:       false,
		NoFormat:        false,
	}

	if err := g.Run(); err != nil {
		return fmt.Errorf("Bootstrap failed: %v", err)
	}
	return nil
}

func main() {
	flag.Parse()

	files := flag.Args()
	if len(files) != 1 {
		flag.Usage()
		os.Exit(1)
		return
	}

	file := files[0]

	if err := generate(file); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
		return
	}
}
