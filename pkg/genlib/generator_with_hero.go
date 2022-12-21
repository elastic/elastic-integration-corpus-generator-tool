// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License 2.0;
// you may not use this file except in compliance with the Elastic License 2.0.

package genlib

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

var heroTemplate = []byte(`
<%!
import "hero/genlib"

var globalFieldMap map[string]genlib.EmitF
var globalState *genlib.GenState

func generate(fieldName string) interface{} {
    return genlib.GenerateFromHero(fieldName, globalFieldMap, globalState)
}

func globals(fieldMap map[string]genlib.EmitF, state *genlib.GenState) {
	if globalFieldMap == nil {
		globalFieldMap = fieldMap
	}
    
	if globalState == nil {
		globalState = state
	}
}
%>
<%: func Template(fieldMap map[string]genlib.EmitF, state *genlib.GenState, w io.Writer) (int, error) %>
<% globals(fieldMap, state) %>
`)

var goMod = []byte(`
module hero

require (
	github.com/Pallinder/go-randomdata v1.2.0 // indirect
	github.com/elastic/go-ucfg v0.8.5 // indirect
	github.com/shiyanhui/hero v0.0.2 // indirect
	golang.org/x/mod v0.4.1 // indirect
	golang.org/x/sync v0.0.0-20201207232520-09787c993a3a // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)
`)

var mainGo = []byte(`
package main

import (
	"context"
	"fmt"
	"hero/genlib"
	"hero/genlib/config"
	"hero/genlib/fields"
	"hero/template"
	"os"
)

func main() {
	if len(os.Args) != 3 && len(os.Args) != 2 {
		os.Exit(1)
	}

	fieldsPath := os.Args[1]

	configPath := ""
	if len(os.Args) == 3 {
		configPath = os.Args[2]
	}
	
	var err error
	cfg := config.Config{}
	if configPath != "" {
		cfg, err = config.LoadConfig(configPath)
		if err != nil {
			fmt.Println(err)
			os.Exit(2)
		}
	}

	fieldsFromYaml, err := fields.LoadFieldsWithTemplate(context.TODO(), fieldsPath)
	if err != nil {
		fmt.Println(err)
		os.Exit(3)
	}

	// Preprocess the fields, generating appropriate emit functions
	fieldMap := make(map[string]genlib.EmitF)
	for _, field := range fieldsFromYaml {
		if err := genlib.BindField(cfg, field, fieldMap); err != nil {
			fmt.Println(err)
			os.Exit(4)
		}
	}

	state := genlib.NewGenState()

	for {
		_, err := template.Template(fieldMap, state, os.Stdout)
		if err != nil {
			fmt.Println(err)
			os.Exit(6)
		}

		state.Inc()

		_, err = os.Stdout.WriteString("\n")		
		if err != nil {
			fmt.Println(err)
			os.Exit(7)
		}
	}
}
`)

// GeneratorWithHero
type GeneratorWithHero struct {
	closed      bool
	closedLock  *sync.Mutex
	heroCommand *exec.Cmd
	heroStdout  chan string
}

func copyDir(source, destination string) error {
	return filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		relPath := strings.Replace(path, source, "", 1)
		if info.IsDir() {
			return os.Mkdir(filepath.Join(destination, relPath), 0770)
		} else {
			if data, err := ioutil.ReadFile(filepath.Join(source, relPath)); err != nil {
				return err
			} else {
				return ioutil.WriteFile(filepath.Join(destination, relPath), data, 0660)
			}
		}
	})
}

func NewGeneratorWithHero(tpl []byte, configPath, fieldsYamlPath string) (*GeneratorWithHero, error) {
	if len(tpl) == 0 {
		tpl = []byte("")
	}

	tmpDir, err := os.MkdirTemp("", "hero-*")
	if err != nil {
		return nil, err
	}

	defer func() {
		// _ = os.RemoveAll(tmpDir)
	}()

	templateDir := filepath.Join(tmpDir, "template")
	err = os.Mkdir(templateDir, 0770)
	if err != nil {
		return nil, err
	}

	randomFilename := path.Base(tmpDir) + ".html"
	fullPath := filepath.Join(templateDir, randomFilename)
	err = ioutil.WriteFile(fullPath, []byte(fmt.Sprintf(`%s%s`, heroTemplate, tpl)), 0660)
	if err != nil {
		return nil, err
	}

	_, filename, _, _ := runtime.Caller(0)
	localDir := path.Dir(filename)
	generatorInterfaceGo, err := ioutil.ReadFile(filepath.Join(localDir, "/generator_interface.go"))
	if err != nil {
		return nil, err
	}

	genlibDir := filepath.Join(tmpDir, "genlib")
	err = os.Mkdir(genlibDir, 0770)
	if err != nil {
		return nil, err
	}

	generatorInterfaceGo = bytes.Replace(generatorInterfaceGo, []byte("github.com/elastic/elastic-integration-corpus-generator-tool/pkg/"), []byte("hero/"), -1)
	err = ioutil.WriteFile(filepath.Join(genlibDir, "generator_interface.go"), generatorInterfaceGo, 0660)
	if err != nil {
		return nil, err
	}

	err = copyDir(filepath.Join(localDir, "config"), filepath.Join(genlibDir, "config"))
	if err != nil {
		return nil, err
	}

	err = copyDir(filepath.Join(localDir, "fields"), filepath.Join(genlibDir, "fields"))
	if err != nil {
		return nil, err
	}

	fullPathMainGo := filepath.Join(tmpDir, "main.go")
	err = ioutil.WriteFile(fullPathMainGo, []byte("package main"), 0660)
	if err != nil {
		return nil, err
	}

	fullPathGoMod := filepath.Join(tmpDir, "go.mod")
	err = ioutil.WriteFile(fullPathGoMod, goMod, 0660)
	if err != nil {
		return nil, err
	}

	goExecutable, err := exec.LookPath("go")
	if err != nil {
		return nil, err
	}

	goGet := &exec.Cmd{
		Dir:    tmpDir,
		Path:   goExecutable,
		Args:   []string{goExecutable, "get", "."},
		Stdout: os.Stdout,
		Stderr: os.Stdout,
	}

	err = goGet.Run()
	if err != nil {
		return nil, err
	}

	heroExecutable, err := exec.LookPath("hero")
	if err != nil {
		return nil, err
	}

	heroCompile := &exec.Cmd{
		Dir:    tmpDir,
		Path:   heroExecutable,
		Args:   []string{heroExecutable, "-source", fmt.Sprintf("%s/template", tmpDir), "-extensions", ".html", "-pkgname", "template"},
		Stdout: os.Stdout,
		Stderr: os.Stdout,
	}

	// Ignore error: hero runs go vet and it fails
	_ = heroCompile.Run()

	err = ioutil.WriteFile(fullPathMainGo, mainGo, 0660)
	if err != nil {
		return nil, err
	}

	// RUN TWICE: go get hero/*
	goGet = &exec.Cmd{
		Dir:    tmpDir,
		Path:   goExecutable,
		Args:   []string{goExecutable, "get", "."},
		Stdout: os.Stdout,
		Stderr: os.Stdout,
	}

	err = goGet.Run()
	if err != nil {
		return nil, err
	}

	heroCommandName := path.Base(tmpDir)
	goBuild := &exec.Cmd{
		Dir:    tmpDir,
		Path:   goExecutable,
		Args:   []string{goExecutable, "build", "-o", heroCommandName, "main.go"},
		Stdout: os.Stdout,
		Stderr: os.Stdout,
	}

	err = goBuild.Run()
	if err != nil {
		return nil, err
	}

	heroCommand := &exec.Cmd{
		Dir:    tmpDir,
		Path:   heroCommandName,
		Args:   []string{heroCommandName, fieldsYamlPath, configPath},
		Stderr: os.Stdout,
	}

	stdoutPipe, err := heroCommand.StdoutPipe()
	if err != nil {
		return nil, err
	}

	stdoutScanner := bufio.NewScanner(stdoutPipe)

	err = heroCommand.Start()
	if err != nil {
		return nil, err
	}

	gen := &GeneratorWithHero{heroCommand: heroCommand, heroStdout: make(chan string), closed: false, closedLock: new(sync.Mutex)}

	go func() {
		for stdoutScanner.Scan() {
			gen.closedLock.Lock()
			if gen.closed {
				close(gen.heroStdout)
				gen.closedLock.Unlock()
				return
			}
			gen.closedLock.Unlock()
			gen.heroStdout <- stdoutScanner.Text()
		}
	}()

	return gen, nil
}

func (gen *GeneratorWithHero) Close() error {
	gen.closedLock.Lock()
	defer gen.closedLock.Unlock()
	gen.closed = true
	return gen.heroCommand.Process.Kill()
}

func (gen *GeneratorWithHero) Emit(state *GenState, buf *bytes.Buffer) error {
	if err := gen.emit(buf); err != nil {
		return err
	}

	state.counter += 1

	return nil
}

func (gen *GeneratorWithHero) emit(buf *bytes.Buffer) error {
	ctx, cancel := context.WithTimeout(context.TODO(), 1*time.Second)
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			_, err := fmt.Fprint(buf, <-gen.heroStdout)
			return err
		}
	}
}
