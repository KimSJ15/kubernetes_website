package gexec

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
)

var tmpDir string

/*
Build uses go build to compile the package at packagePath.  The resulting binary is saved off in a temporary directory.
A path pointing to this binary is returned.

Build uses the $GOPATH set in your environment.  It passes the variadic args on to `go build`.
*/
func Build(packagePath string, args ...string) (compiledPath string, err error) {
	return BuildIn(os.Getenv("GOPATH"), packagePath, args...)
}

/*
BuildIn is identical to Build but allows you to specify a custom $GOPATH (the first argument).
*/
func BuildIn(gopath string, packagePath string, args ...string) (compiledPath string, err error) {
	tmpDir, err := temporaryDirectory()
	if err != nil {
		return "", err
	}

	if len(gopath) == 0 {
		return "", errors.New("$GOPATH not provided when building " + packagePath)
	}

	executable := filepath.Join(tmpDir, path.Base(packagePath))
	if runtime.GOOS == "windows" {
		executable = executable + ".exe"
	}

	cmdArgs := append([]string{"build"}, args...)
	cmdArgs = append(cmdArgs, "-o", executable, packagePath)

	build := exec.Command("go", cmdArgs...)
	build.Env = append([]string{"GOPATH=" + gopath}, os.Environ()...)

	output, err := build.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("Failed to build %s:\n\nError:\n%s\n\nOutput:\n%s", packagePath, err, string(output))
	}

	return executable, nil
}

/*
You should call CleanupBuildArtifacts before your test ends to clean up any temporary artifacts generated by
gexec. In Ginkgo this is typically done in an AfterSuite callback.
*/
func CleanupBuildArtifacts() {
	if tmpDir != "" {
		os.RemoveAll(tmpDir)
	}
}

func temporaryDirectory() (string, error) {
	var err error
	if tmpDir == "" {
		tmpDir, err = ioutil.TempDir("", "gexec_artifacts")
		if err != nil {
			return "", err
		}
	}

	return ioutil.TempDir(tmpDir, "g")
}
