package test

import (
	"strings"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"github.com/kr/pretty"
)

const tmpPath = "./test/tmp"
const binaryName = "mani"
var binaryPath string
var update = flag.Bool("update", false, "update golden files")

type TemplateTest struct {
	Config  string // mani.yaml config to use
	Tmpl    string // Name
	Golden  string // Golden dir and filename <dir>/<filename>
	WantErr bool   // In-case error is wanted
	Args    string  // Command to run
}

type TestFile struct {
	t    *testing.T
	name string
	dir  string
}

func NewGoldenFile(t *testing.T, name string) *TestFile {
	return &TestFile{t: t, name: name, dir: "golden"}
}

func (tf *TestFile) path() string {
	tf.t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		tf.t.Fatal("problems recovering caller information")
	}

	return filepath.Join(filepath.Dir(filename), tf.dir, tf.name)
}

func (tf *TestFile) Write(content string) {
	tf.t.Helper()
	err := ioutil.WriteFile(tf.path(), []byte(content), 0644)
	if err != nil {
		tf.t.Fatalf("could not write %s: %v", tf.name, err)
	}
}

func (tf *TestFile) AsFile() *os.File {
	tf.t.Helper()
	file, err := os.Open(tf.path())
	if err != nil {
		tf.t.Fatalf("could not open %s: %v", tf.name, err)
	}
	return file
}

func (tf *TestFile) load() string {
	tf.t.Helper()

	content, err := ioutil.ReadFile(tf.path())
	if err != nil {
		tf.t.Fatalf("could not read file %s: %v", tf.name, err)
	}

	return string(content)
}

func copy(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

func clearTmp() {
	tmpDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("Could not get working directory")
		os.Exit(1)
	}

	var baseDir = filepath.Base(tmpDir)

	if baseDir == "tmp" {
		dir, _ := ioutil.ReadDir(tmpDir)
		for _, d := range dir {
			os.RemoveAll(path.Join([]string{d.Name()}...))
		}

	} else {
		fmt.Printf("Not inside tmp directory!")
		os.Exit(1)
	}
}

func diff(expected, actual interface{}) []string {
	return pretty.Diff(expected, actual)
}

// 1. Create mani binary
// 2. Create test/tmp directory
// 3. cd into test/tmp
func TestMain(m *testing.M) {
	err := os.Chdir("../..")

	if err != nil {
		fmt.Printf("could not change dir: %v", err)
		os.Exit(1)
	}

	make := exec.Command("make", "build-dev")

	err = make.Run()
	if err != nil {
		fmt.Printf("could not make binary for %s: %v", binaryName, err)
		os.Exit(1)
	}

	if _, err = os.Stat(tmpPath); os.IsNotExist(err) {
		err = os.Mkdir(tmpPath, 0755)

		if err != nil {
			fmt.Printf("could not create directory at %s: %v", tmpPath, err)
			os.Exit(1)
		}
	}

	abs, err := filepath.Abs(binaryName)
	if err != nil {
		fmt.Printf("could not get abs path for %s: %v", binaryName, err)
		os.Exit(1)
	}

	binaryPath = abs

	err = os.Chdir(tmpPath)
	if err != nil {
		fmt.Printf("could not change dir: %v", err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func Run(t *testing.T, tt TemplateTest) {
	t.Cleanup(func() {
		clearTmp()
	})

	var configPath = filepath.Join("../fixtures", tt.Config)
	copy(configPath, "./mani.yaml")

	cmd := exec.Command(binaryPath, strings.Split(tt.Args, " ")...)

	output, err := cmd.CombinedOutput()
	if (err != nil) != tt.WantErr {
		t.Fatalf("%s\nexpected (err != nil) to be %v, but got %v. err: %v", output, tt.WantErr, err != nil, err)
	}

	golden := NewGoldenFile(t, tt.Golden)
	actual := string(output)

	if *update {
		golden.Write(actual)
	}

	expected := golden.load()

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("diff: %v", diff(expected, actual))
	}
}
