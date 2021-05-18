package integration

import (
	"strings"
	"flag"
	"fmt"
	// "io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"github.com/otiai10/copy"
	"github.com/kr/pretty"
)

const binaryName = "mani"
const testDir = "./test"
var tmpPath = filepath.Join(testDir, "tmp")
var goldenDir = filepath.Join(testDir, "integration", "golden")
var binaryPath string
var verbose = flag.Bool("verbose", false, "verbose")
var update = flag.Bool("update", false, "update golden files")
var dirty = flag.Bool("dirty", false, "Skip clean tmp directory after run")

type TemplateTest struct {
	TestName       string
	InputFiles     []string
	TestCmd        string
	Golden         string
	WantErr		   bool
}

type TestFile struct {
	t    *testing.T
	name string
	dir  string
}

func NewGoldenFile(t *testing.T, name string) *TestFile {
	return &TestFile{t: t, name: "stdout.golden", dir: filepath.Join("golden", name) }
}

func (tf *TestFile) Dir() string {
	tf.t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		tf.t.Fatal("problems recovering caller information")
	}

	return filepath.Join(filepath.Dir(filename), tf.dir)
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
	err := os.MkdirAll(filepath.Dir(tf.path()), os.ModePerm)
	if err != nil {
		tf.t.Fatalf("could not create directory %s: %v", tf.name, err)
	}

	err = ioutil.WriteFile(tf.path(), []byte(content), 0644)
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

// func copy(src, dst string) (int64, error) {
// 	sourceFileStat, err := os.Stat(src)

// 	if err != nil {
// 		return 0, err
// 	}

// 	if !sourceFileStat.Mode().IsRegular() {
// 		return 0, fmt.Errorf("%s is not a regular file", src)
// 	}

// 	source, err := os.Open(src)
// 	if err != nil {
// 		return 0, err
// 	}
// 	defer source.Close()

// 	destination, err := os.Create(dst)
// 	if err != nil {
// 		return 0, err
// 	}
// 	defer destination.Close()
// 	nBytes, err := io.Copy(destination, source)

// 	return nBytes, err
// }

func clearGolden(goldenDir string) {
	// Guard against accidently deleting outside directory
	if strings.Contains(goldenDir, "golden") {
		os.RemoveAll(goldenDir)
	}
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
	var tmpDir, _ = os.Getwd()
	var fixturesDir = filepath.Join(tmpDir, "../fixtures")
	var rootPath = filepath.Join(tmpDir, "../..")

	t.Cleanup(func() {
		if *dirty == false {
			clearTmp()
		}
	})

	// Copy fixture files
	for _, file := range tt.InputFiles {
		var configPath = filepath.Join(fixturesDir,  file)
		err := copy.Copy(configPath, filepath.Base(file))

		if err != nil {
			fmt.Printf("\x1b[31;1m%s\x1b[0m\n", fmt.Sprintf("error: %s", err))
			os.Exit(1)
		}
	}

	// Run test command
	cmd := exec.Command("sh", "-c", tt.TestCmd)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "MANI=" + binaryPath )

	output, err := cmd.CombinedOutput()

	// TEST: Check we get error if we want error
	if (err != nil) != tt.WantErr {
		t.Fatalf("%s\nexpected (err != nil) to be %v, but got %v. err: %v", output, tt.WantErr, err != nil, err)
	}

	if *verbose {
		fmt.Println(tt.TestCmd)
		fmt.Println(string(output))
	}

	// Save output from command as golden file
	golden := NewGoldenFile(t, tt.Golden)
	actual := string(output)

	if *update {
		clearGolden(golden.Dir())

		// Write stdout of test command to golden file
		golden.Write(actual)

		err := copy.Copy(tmpDir, golden.Dir())
		if err != nil {
			fmt.Printf("\x1b[31;1m%s\x1b[0m\n", fmt.Sprintf("error: %s", err))
			os.Exit(1)
		}
	} else {
		// Compare files
		expected := golden.load()
		if !reflect.DeepEqual(actual, expected) {
			t.Fatalf("diff: %v", diff(expected, actual))
		}

		err := filepath.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
			if path == tmpDir {
				return nil
			}

			if err != nil {
				return err
			}

			goldenPath := filepath.Join(rootPath, goldenDir, tt.Golden, filepath.Base(path))

			actual, err := ioutil.ReadFile(path)
			expected, err := ioutil.ReadFile(goldenPath)

			// TEST: Compare file content
			if !reflect.DeepEqual(actual, expected) {
				t.Fatalf("diff: %v", diff(expected, actual))
			}

			return nil
		})

		if err != nil {
			t.Fatalf("Error: %v", err)
		}
	}
}
