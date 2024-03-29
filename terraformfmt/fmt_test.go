// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package terraformfmt

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

/* TODO?
func TestFmt_TestFiles(t *testing.T) {
	const inSuffix = "_in.tftest.hcl"
	const outSuffix = "_out.tftest.hcl"
	const gotSuffix = "_got.tftest.hcl"
	entries, err := ioutil.ReadDir("testdata")
	if err != nil {
		t.Fatal(err)
	}

	for _, info := range entries {
		if info.IsDir() {
			continue
		}
		filename := info.Name()
		if !strings.HasSuffix(filename, inSuffix) {
			continue
		}
		testName := filename[:len(filename)-len(inSuffix)]
		t.Run(testName, func(t *testing.T) {
			inFile := filepath.Join("testdata", testName+inSuffix)
			wantFile := filepath.Join("testdata", testName+outSuffix)
			input, err := ioutil.ReadFile(inFile)
			if err != nil {
				t.Fatal(err)
			}
			inHCL, diags := hclwrite.ParseConfig(input, inFile, hcl.InitialPos)
			if diags.HasErrors() {
				t.Fatalf("parsing input %s: %s", inFile, diags)
			}

			want, err := os.ReadFile(wantFile)
			if err != nil {
				t.Fatal(err)
			}
			FormatBody(inHCL.Body(), nil)
			got := inHCL.Bytes()

			if diff := cmp.Diff(string(want), string(got)); diff != "" {
				t.Errorf("wrong result\n%s", diff)
			}
		})
	}
}
*/

func TestFmt(t *testing.T) {
	const inSuffix = "_in.tf"
	const outSuffix = "_out.tf"
	entries, err := os.ReadDir("testdata")
	if err != nil {
		t.Fatal(err)
	}

	for _, info := range entries {
		if info.IsDir() {
			continue
		}
		filename := info.Name()
		if !strings.HasSuffix(filename, inSuffix) {
			continue
		}
		testName := filename[:len(filename)-len(inSuffix)]
		t.Run(testName, func(t *testing.T) {
			inFile := filepath.Join("testdata", testName+inSuffix)
			wantFile := filepath.Join("testdata", testName+outSuffix)
			input, err := os.ReadFile(inFile)
			if err != nil {
				t.Fatal(err)
			}
			inHCL, diags := hclwrite.ParseConfig(input, inFile, hcl.InitialPos)
			if diags.HasErrors() {
				t.Fatalf("parsing input %s: %s", inFile, diags)
			}

			want, err := os.ReadFile(wantFile)
			if err != nil {
				t.Fatal(err)
			}
			FormatBody(inHCL.Body())
			got := inHCL.Bytes()

			if diff := cmp.Diff(string(want), string(got)); diff != "" {
				t.Errorf("wrong result\n%s", diff)
			}
		})
	}
}

/* TODO

func TestFmt_nonexist(t *testing.T) {
	tempDir := fmtFixtureWriteDir(t)

	ui := new(cli.MockUi)
	c := &FmtCommand{
		Meta: Meta{
			testingOverrides: metaOverridesForProvider(testProvider()),
			Ui:               ui,
		},
	}

	missingDir := filepath.Join(tempDir, "doesnotexist")
	args := []string{missingDir}
	if code := c.Run(args); code != 2 {
		t.Fatalf("wrong exit code. errors: \n%s", ui.ErrorWriter.String())
	}

	expected := "No file or directory at"
	if actual := ui.ErrorWriter.String(); !strings.Contains(actual, expected) {
		t.Fatalf("expected:\n%s\n\nto include: %q", actual, expected)
	}
}

func TestFmt_syntaxError(t *testing.T) {
	tempDir := testTempDir(t)

	invalidSrc := `
a = 1 +
`

	err := ioutil.WriteFile(filepath.Join(tempDir, "invalid.tf"), []byte(invalidSrc), 0644)
	if err != nil {
		t.Fatal(err)
	}

	ui := new(cli.MockUi)
	c := &FmtCommand{
		Meta: Meta{
			testingOverrides: metaOverridesForProvider(testProvider()),
			Ui:               ui,
		},
	}

	args := []string{tempDir}
	if code := c.Run(args); code != 2 {
		t.Fatalf("wrong exit code. errors: \n%s", ui.ErrorWriter.String())
	}

	expected := "Invalid expression"
	if actual := ui.ErrorWriter.String(); !strings.Contains(actual, expected) {
		t.Fatalf("expected:\n%s\n\nto include: %q", actual, expected)
	}
}

func TestFmt_snippetInError(t *testing.T) {
	tempDir := testTempDir(t)

	backendSrc := `terraform {backend "s3" {}}`

	err := ioutil.WriteFile(filepath.Join(tempDir, "backend.tf"), []byte(backendSrc), 0644)
	if err != nil {
		t.Fatal(err)
	}

	ui := new(cli.MockUi)
	c := &FmtCommand{
		Meta: Meta{
			testingOverrides: metaOverridesForProvider(testProvider()),
			Ui:               ui,
		},
	}

	args := []string{tempDir}
	if code := c.Run(args); code != 2 {
		t.Fatalf("wrong exit code. errors: \n%s", ui.ErrorWriter.String())
	}

	substrings := []string{
		"Argument definition required",
		"line 1, in terraform",
		`1: terraform {backend "s3" {}}`,
	}
	for _, substring := range substrings {
		if actual := ui.ErrorWriter.String(); !strings.Contains(actual, substring) {
			t.Errorf("expected:\n%s\n\nto include: %q", actual, substring)
		}
	}
}

func TestFmt_manyArgs(t *testing.T) {
	tempDir := fmtFixtureWriteDir(t)
	// Add a second file
	secondSrc := `locals { x = 1 }`

	err := ioutil.WriteFile(filepath.Join(tempDir, "second.tf"), []byte(secondSrc), 0644)
	if err != nil {
		t.Fatal(err)
	}

	ui := new(cli.MockUi)
	c := &FmtCommand{
		Meta: Meta{
			testingOverrides: metaOverridesForProvider(testProvider()),
			Ui:               ui,
		},
	}

	args := []string{
		filepath.Join(tempDir, "main.tf"),
		filepath.Join(tempDir, "second.tf"),
	}
	if code := c.Run(args); code != 0 {
		t.Fatalf("wrong exit code. errors: \n%s", ui.ErrorWriter.String())
	}

	got, err := filepath.Abs(strings.TrimSpace(ui.OutputWriter.String()))
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(tempDir, fmtFixture.filename)

	if got != want {
		t.Fatalf("wrong output\ngot:  %s\nwant: %s", got, want)
	}
}

func TestFmt_workingDirectory(t *testing.T) {
	tempDir := fmtFixtureWriteDir(t)

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Chdir(cwd)

	ui := new(cli.MockUi)
	c := &FmtCommand{
		Meta: Meta{
			testingOverrides: metaOverridesForProvider(testProvider()),
			Ui:               ui,
		},
	}

	args := []string{}
	if code := c.Run(args); code != 0 {
		t.Fatalf("wrong exit code. errors: \n%s", ui.ErrorWriter.String())
	}

	expected := fmt.Sprintf("%s\n", fmtFixture.filename)
	if actual := ui.OutputWriter.String(); actual != expected {
		t.Fatalf("got: %q\nexpected: %q", actual, expected)
	}
}

func TestFmt_directoryArg(t *testing.T) {
	tempDir := fmtFixtureWriteDir(t)

	ui := new(cli.MockUi)
	c := &FmtCommand{
		Meta: Meta{
			testingOverrides: metaOverridesForProvider(testProvider()),
			Ui:               ui,
		},
	}

	args := []string{tempDir}
	if code := c.Run(args); code != 0 {
		t.Fatalf("wrong exit code. errors: \n%s", ui.ErrorWriter.String())
	}

	got, err := filepath.Abs(strings.TrimSpace(ui.OutputWriter.String()))
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(tempDir, fmtFixture.filename)

	if got != want {
		t.Fatalf("wrong output\ngot:  %s\nwant: %s", got, want)
	}
}

func TestFmt_fileArg(t *testing.T) {
	tempDir := fmtFixtureWriteDir(t)

	ui := new(cli.MockUi)
	c := &FmtCommand{
		Meta: Meta{
			testingOverrides: metaOverridesForProvider(testProvider()),
			Ui:               ui,
		},
	}

	args := []string{filepath.Join(tempDir, fmtFixture.filename)}
	if code := c.Run(args); code != 0 {
		t.Fatalf("wrong exit code. errors: \n%s", ui.ErrorWriter.String())
	}

	got, err := filepath.Abs(strings.TrimSpace(ui.OutputWriter.String()))
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(tempDir, fmtFixture.filename)

	if got != want {
		t.Fatalf("wrong output\ngot:  %s\nwant: %s", got, want)
	}
}

func TestFmt_stdinArg(t *testing.T) {
	input := new(bytes.Buffer)
	input.Write(fmtFixture.input)

	ui := new(cli.MockUi)
	c := &FmtCommand{
		Meta: Meta{
			testingOverrides: metaOverridesForProvider(testProvider()),
			Ui:               ui,
		},
		input: input,
	}

	args := []string{"-"}
	if code := c.Run(args); code != 0 {
		t.Fatalf("wrong exit code. errors: \n%s", ui.ErrorWriter.String())
	}

	expected := fmtFixture.golden
	if actual := ui.OutputWriter.Bytes(); !bytes.Equal(actual, expected) {
		t.Fatalf("got: %q\nexpected: %q", actual, expected)
	}
}

func TestFmt_nonDefaultOptions(t *testing.T) {
	tempDir := fmtFixtureWriteDir(t)

	ui := new(cli.MockUi)
	c := &FmtCommand{
		Meta: Meta{
			testingOverrides: metaOverridesForProvider(testProvider()),
			Ui:               ui,
		},
	}

	args := []string{
		"-list=false",
		"-write=false",
		"-diff",
		tempDir,
	}
	if code := c.Run(args); code != 0 {
		t.Fatalf("wrong exit code. errors: \n%s", ui.ErrorWriter.String())
	}

	expected := fmt.Sprintf("-%s+%s", fmtFixture.input, fmtFixture.golden)
	if actual := ui.OutputWriter.String(); !strings.Contains(actual, expected) {
		t.Fatalf("expected:\n%s\n\nto include: %q", actual, expected)
	}
}

func TestFmt_check(t *testing.T) {
	tempDir := fmtFixtureWriteDir(t)

	ui := new(cli.MockUi)
	c := &FmtCommand{
		Meta: Meta{
			testingOverrides: metaOverridesForProvider(testProvider()),
			Ui:               ui,
		},
	}

	args := []string{
		"-check",
		tempDir,
	}
	if code := c.Run(args); code != 3 {
		t.Fatalf("wrong exit code. expected 3")
	}

	// Given that we give relative paths back to the user, normalize this temp
	// dir so that we're comparing against a relative-ized (normalized) path
	tempDir = c.normalizePath(tempDir)

	if actual := ui.OutputWriter.String(); !strings.Contains(actual, tempDir) {
		t.Fatalf("expected:\n%s\n\nto include: %q", actual, tempDir)
	}
}

func TestFmt_checkStdin(t *testing.T) {
	input := new(bytes.Buffer)
	input.Write(fmtFixture.input)

	ui := new(cli.MockUi)
	c := &FmtCommand{
		Meta: Meta{
			testingOverrides: metaOverridesForProvider(testProvider()),
			Ui:               ui,
		},
		input: input,
	}

	args := []string{
		"-check",
		"-",
	}
	if code := c.Run(args); code != 3 {
		t.Fatalf("wrong exit code. expected 3, got %d", code)
	}

	if ui.OutputWriter != nil {
		t.Fatalf("expected no output, got: %q", ui.OutputWriter.String())
	}
}

var fmtFixture = struct {
	filename      string
	input, golden []byte
}{
	"main.tf",
	[]byte(`  foo  =  "bar"
`),
	[]byte(`foo = "bar"
`),
}

func fmtFixtureWriteDir(t *testing.T) string {
	dir := testTempDir(t)

	err := ioutil.WriteFile(filepath.Join(dir, fmtFixture.filename), fmtFixture.input, 0644)
	if err != nil {
		t.Fatal(err)
	}

	return dir
}
*/
