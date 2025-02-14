package languages

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"

	"github.com/nikogura/gomason/pkg/gomason"
	"github.com/pkg/errors"
)

func TestCreateGoPath(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "gomason")
	if err != nil {
		log.Printf("Error creating temp dir\n")
		t.Fail()
	}
	defer os.RemoveAll(tmpDir)

	lang, _ := GetByName("golang")
	_, err = lang.CreateWorkDir(tmpDir)
	if err != nil {
		log.Printf("Error creating gopath in %q: %s", tmpDir, err)
		t.Fail()
	}

	dirs := []string{"go", "go/src", "go/pkg", "go/bin"}

	for _, dir := range dirs {
		if _, err := os.Stat(fmt.Sprintf("%s/%s", tmpDir, dir)); os.IsNotExist(err) {
			t.Fail()
		}
	}
}

func TestCheckoutDefault(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "golang")
	if err != nil {
		log.Fatal("Failed to create temp dir for testing")
	}
	defer os.RemoveAll(tmpDir)

	lang, _ := GetByName("golang")
	gopath, err := lang.CreateWorkDir(tmpDir)
	if err != nil {
		log.Printf("Error creating GOPATH in %s: %s", tmpDir, err)
		t.Fail()
	}

	log.Printf("Checking out Master Branch")
	err = lang.Checkout(gopath, gomason.TestMetadataObj(), "master", true)
	if err != nil {
		log.Printf("Failed to checkout module: %s", err)
		t.Fail()
	}

	if _, err := os.Stat(fmt.Sprintf("%s/src/%s/metadata.json", gopath, gomason.TestModuleName())); os.IsNotExist(err) {
		log.Printf("Failed to checkout module")
		t.Fail()
	}
}

func TestCheckoutBranch(t *testing.T) {
	log.Printf("Checking out Test Branch")

	// making a separate temp dir here cos it steps on the other tests
	dir, err := ioutil.TempDir("", "gomason")
	if err != nil {
		log.Fatal("Error creating temp dir\n")
	}
	defer os.RemoveAll(dir)

	lang, _ := GetByName("golang")
	gopath, err := lang.CreateWorkDir(dir)
	if err != nil {
		log.Printf("Error creating GOPATH in %s: %s", dir, err)
		t.Fail()
	}

	err = lang.Checkout(gopath, gomason.TestMetadataObj(), "testbranch", true)
	if err != nil {
		log.Printf("Failed to checkout module: %s", err)
		t.Fail()
	}

	if _, err := os.Stat(fmt.Sprintf("%s/src/%s/test_file", gopath, gomason.TestModuleName())); os.IsNotExist(err) {
		log.Printf("Failed to checkout branch")
		t.Fail()
	}
}

func TestPrep(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "golang")
	if err != nil {
		log.Fatal("Failed to create temp dir for testing")
	}
	defer os.RemoveAll(tmpDir)

	log.Printf("Checking out Master Branch")
	lang, _ := GetByName("golang")
	gopath, err := lang.CreateWorkDir(tmpDir)
	if err != nil {
		log.Printf("Error creating GOPATH in %s: %s", tmpDir, err)
		t.Fail()
	}

	err = lang.Checkout(gopath, gomason.TestMetadataObj(), "master", true)
	if err != nil {
		log.Printf("Failed to checkout module: %s", err)
		t.Fail()
	}

	if _, err := os.Stat(fmt.Sprintf("%s/src/%s/metadata.json", gopath, gomason.TestModuleName())); os.IsNotExist(err) {
		log.Printf("Failed to checkout module")
		t.Fail()
	}

	err = lang.Prep(gopath, gomason.TestMetadataObj(), true)
	if err != nil {
		log.Printf("error running prep steps: %s", err)
		t.Fail()
	}
}

func TestBuildGoxInstall(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "golang")
	if err != nil {
		log.Fatal("Failed to create temp dir for testing")
	}
	defer os.RemoveAll(tmpDir)

	lang, _ := GetByName("golang")

	log.Printf("Installing Go\n")
	gopath, err := lang.CreateWorkDir(tmpDir)
	if err != nil {
		log.Printf("Error creating GOPATH in %s: %s", tmpDir, err)
		t.Fail()
	}

	err = GoxInstall(gopath, true)
	if err != nil {
		log.Printf("Error installing Gox: %s\n", err)
		t.Fail()
	}

	if _, err := os.Stat(fmt.Sprintf("%s/go/bin/gox", tmpDir)); os.IsNotExist(err) {
		log.Printf("Gox failed to install.")
		t.Fail()
	}
}

func TestBuild(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "golang")
	if err != nil {
		log.Fatal("Failed to create temp dir for testing")
	}
	defer os.RemoveAll(tmpDir)

	lang, _ := GetByName("golang")

	log.Printf("Running Build\n")
	gopath, err := lang.CreateWorkDir(tmpDir)
	if err != nil {
		log.Printf("Error creating GOPATH in %s: %s\n", tmpDir, err)
		t.Fail()
	}

	gomodule := gomason.TestMetadataObj().Package
	branch := "master"

	log.Printf("Checking out Master Branch")

	err = lang.Checkout(gopath, gomason.TestMetadataObj(), "master", true)
	if err != nil {
		log.Printf("Failed to checkout module: %s", err)
		t.Fail()
	}

	if _, err := os.Stat(fmt.Sprintf("%s/src/%s/metadata.json", gopath, gomason.TestModuleName())); os.IsNotExist(err) {
		log.Printf("Failed to checkout module")
		t.Fail()
	}

	err = lang.Prep(gopath, gomason.TestMetadataObj(), true)
	if err != nil {
		log.Printf("error running prep steps: %s", err)
		t.Fail()
	}

	err = lang.Build(gopath, gomason.TestMetadataObj(), branch, true)
	if err != nil {
		log.Printf("Error building: %s", err)
		t.Fail()
	}

	parts := strings.Split(gomodule, "/")

	binaryPrefix := parts[len(parts)-1]

	osname := runtime.GOOS
	archname := runtime.GOARCH

	workdir := fmt.Sprintf("%s/src/%s", gopath, gomodule)
	binary := fmt.Sprintf("%s/%s_%s_%s", workdir, binaryPrefix, osname, archname)

	log.Printf("Looking for binary: %s", binary)

	if _, err := os.Stat(binary); os.IsNotExist(err) {
		log.Printf("Gox failed to build binary: %s.\n", binary)
		t.Fail()
	} else {
		log.Printf("Binary found.")
	}
}

func TestTest(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "golang")
	if err != nil {
		log.Fatal("Failed to create temp dir for testing")
	}
	defer os.RemoveAll(tmpDir)

	lang, _ := GetByName("golang")

	log.Printf("Checking out Master Branch")
	gopath, err := lang.CreateWorkDir(tmpDir)
	if err != nil {
		log.Printf("Error creating GOPATH in %s: %s", tmpDir, err)
		t.Fail()
	}

	err = lang.Checkout(gopath, gomason.TestMetadataObj(), "master", true)
	if err != nil {
		log.Printf("Failed to checkout module: %s", err)
		t.Fail()
	}

	if _, err := os.Stat(fmt.Sprintf("%s/src/%s/metadata.json", gopath, gomason.TestModuleName())); os.IsNotExist(err) {
		log.Printf("Failed to checkout module")
		t.Fail()
	}

	err = lang.Prep(gopath, gomason.TestMetadataObj(), true)
	if err != nil {
		log.Printf("error running prep steps: %s", err)
		t.Fail()
	}

	err = lang.Test(gopath, gomason.TestMetadataObj().Package, true)
	if err != nil {
		log.Printf("error running go test: %s", err)
		t.Fail()
	}
}

func TestSignVerifyBinary(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "golang")
	if err != nil {
		log.Fatal("Failed to create temp dir for testing")
	}
	defer os.RemoveAll(tmpDir)

	shellCmd, err := exec.LookPath("gpg")
	if err != nil {
		log.Printf("Failed to check if gpg is installed:%s", err)
		t.Fail()
	}

	lang, _ := GetByName("golang")

	// create workspace
	gopath, err := lang.CreateWorkDir(tmpDir)
	if err != nil {
		log.Printf("Error creating GOPATH in %s: %s\n", tmpDir, err)
		t.Fail()
	}

	meta := gomason.TestMetadataObj()

	meta.Repository = fmt.Sprintf("http://localhost:%d/repo/tool", servicePort)

	darwin := gomason.PublishTarget{
		Source:      "gomason_darwin_amd64",
		Destination: "{{.Repository}}/gomason/{{.Version}}/darwin/amd64/gomason",
		Signature:   true,
		Checksums:   true,
	}

	linux := gomason.PublishTarget{
		Source:      "gomason_linux_amd64",
		Destination: "{{.Repository}}/gomason/{{.Version}}/linux/amd64/gomason",
		Signature:   true,
		Checksums:   true,
	}

	targets := []gomason.PublishTarget{darwin, linux}

	targetsMap := make(map[string]gomason.PublishTarget)

	for _, target := range targets {
		targetsMap[target.Source] = target
	}

	pubInfo := gomason.PublishInfo{
		Targets:    targets,
		TargetsMap: targetsMap,
	}

	meta.PublishInfo = pubInfo

	branch := "master"

	// build artifacts
	log.Printf("Running Build\n")
	err = lang.Build(gopath, meta, branch, true)
	if err != nil {
		log.Printf("Error building: %s", err)
		t.Fail()
	}

	// set up test keys
	keyring := fmt.Sprintf("%s/keyring.gpg", tmpDir)
	trustdb := fmt.Sprintf("%s/trustdb.gpg", tmpDir)

	meta.Options = make(map[string]interface{})
	meta.Options["keyring"] = keyring
	meta.Options["trustdb"] = trustdb

	// write gpg batch file
	defaultKeyText := `%echo Generating a default key
%no-protection
%transient-key
Key-Type: default
Subkey-Type: default
Name-Real: Gomason Tester
Name-Comment: with no passphrase
Name-Email: gomason-tester@foo.com
Expire-Date: 0
%commit
%echo done
`
	keyFile := fmt.Sprintf("%s/testkey", tmpDir)
	err = ioutil.WriteFile(keyFile, []byte(defaultKeyText), 0644)
	if err != nil {
		log.Printf("Error writing test key generation file: %s", err)
		t.Fail()
	}

	log.Printf("Keyring file: %s", keyring)
	log.Printf("Trustdb file: %s", trustdb)
	log.Printf("Test key generation file: %s", keyFile)

	// generate a test key
	cmd := exec.Command(shellCmd, "--trustdb", trustdb, "--no-default-keyring", "--keyring", keyring, "--batch", "--generate-key", keyFile)
	err = cmd.Run()
	if err != nil {
		log.Printf("****** Error creating test key: %s *****", err)
		t.Fail()
	}

	log.Printf("Done creating keyring and test keys")

	// sign binaries
	parts := strings.Split(meta.Package, "/")
	binaryPrefix := parts[len(parts)-1]

	for _, target := range meta.BuildInfo.Targets {
		archparts := strings.Split(target.Name, "/")

		osname := archparts[0]   // linux or darwin generally
		archname := archparts[1] // amd64 generally

		workdir := fmt.Sprintf("%s/src/%s", gopath, meta.Package)
		binary := fmt.Sprintf("%s/%s_%s_%s", workdir, binaryPrefix, osname, archname)

		if _, err := os.Stat(binary); os.IsNotExist(err) {
			fmt.Printf("Gox failed to build binary: %s\n", binary)
			log.Printf("Failed to find binary %s", binary)
			t.Fail()
		}

		err = gomason.SignBinary(meta, binary, true)
		if err != nil {
			err = errors.Wrap(err, "failed to sign binary")
			log.Printf("Failed to sign binary %s: %s", binary, err)
			t.Fail()
		}

		// verify binaries
		ok, err := gomason.VerifyBinary(binary, meta, true)
		if err != nil {
			log.Printf("Error verifying signature: %s", err)
			//t.Fail()
		}

		if !ok {
			log.Printf("Failed to verify signature on %s", binary)
			t.Fail()
		}

	}

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current working directory: %s", err)
	}

	fmt.Printf("Publishing\n")

	err = gomason.HandleArtifacts(meta, gopath, cwd, false, true, true, true)
	if err != nil {
		log.Fatalf("post-build processing failed: %s", err)
	}

	err = gomason.HandleExtras(meta, gopath, cwd, false, true, true)
	if err != nil {
		log.Fatalf("Extra artifact processing failed: %s", err)
	}
}
