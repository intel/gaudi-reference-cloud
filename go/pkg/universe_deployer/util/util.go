// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package util

import (
	"bufio"
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"io/fs"
	"math/big"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	flag "github.com/spf13/pflag"
)

const ConfigCommitLabel = "cloud.intel.com/config-commit"

// Copy file, preserving mode bits.
func CopyFile(source string, dest string) error {
	sourceFile, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("opening source: %w", err)
	}
	defer sourceFile.Close()

	sourceInfo, err := sourceFile.Stat()
	if err != nil {
		return err
	}

	parent := path.Dir(dest)
	if err := os.MkdirAll(parent, 0750); err != nil {
		return fmt.Errorf("creating destination directory: %w", err)
	}

	destFile, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE|os.O_TRUNC, sourceInfo.Mode())
	if err != nil {
		return fmt.Errorf("creating destination file: %w", err)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	if err := sourceFile.Close(); err != nil {
		return fmt.Errorf("closing source: %w", err)
	}
	if err := destFile.Close(); err != nil {
		return fmt.Errorf("closing destination: %w", err)
	}
	return nil
}

// CopyDir recursively copies the entire directory from source to dest.
func CopyDir(source string, dest string) error {
	sourceDir, err := os.Open(source)
	if err != nil {
		return err
	}
	defer sourceDir.Close()

	// Read the contents of the source directory.
	entries, err := sourceDir.Readdir(-1)
	if err != nil {
		return err
	}

	// Create the destination directory.
	if err := os.MkdirAll(dest, 0750); err != nil {
		return err
	}

	// Iterate through each entry in the source directory.
	for _, entry := range entries {
		sourcePath := filepath.Join(source, entry.Name())
		destPath := filepath.Join(dest, entry.Name())

		if entry.IsDir() {
			// If it's a directory, recursively copy it.
			err := CopyDir(sourcePath, destPath)
			if err != nil {
				return err
			}
		} else {
			// If it's a file, copy it.
			err := CopyFile(sourcePath, destPath)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Run chmod on all files in a directory tree.
func ChmodAll(path string, mode fs.FileMode) error {
	err := filepath.WalkDir(path,
		func(s string, d fs.DirEntry, err error) error {
			if err := os.Chmod(s, mode); err != nil {
				return err
			}
			return nil
		})
	return err
}

// Run chmod on all files in a directory tree, then delete the directory tree.
func ChmodAndRemoveAll(ctx context.Context, path string, mode fs.FileMode) error {
	log := log.FromContext(ctx).WithName("ChmodAndRemoveAll").WithValues("path", path)
	log.V(2).Info("BEGIN")
	defer log.V(2).Info("END")
	errChmodWalk := filepath.WalkDir(path,
		func(s string, d fs.DirEntry, err error) error {
			if err := os.Chmod(s, mode); err != nil && !os.IsNotExist(err) {
				return err
			}
			return nil
		})
	if errChmodWalk != nil {
		// Continue if there are errors with chmod.
		log.V(2).Info(fmt.Sprintf("walk error: %v", errChmodWalk))
	}
	if err := os.RemoveAll(path); err != nil {
		// Log the error because this function is often called with defer which ignores the returned error.
		log.V(2).Info(fmt.Sprintf("delete directory error: %v", err))
		return err
	}
	return nil
}

// DeleteContents deletes all files and subdirectories in the specified directory.
func DeleteDirectoryContents(dir string) error {
	// Read the contents of the directory
	contents, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	// Iterate over each item in the directory
	for _, entry := range contents {
		// Create the full path
		path := filepath.Join(dir, entry.Name())
		// Check if it's a directory
		if entry.IsDir() {
			// Remove the directory
			if err := os.RemoveAll(path); err != nil {
				return err
			}
		} else {
			// Remove the file
			if err := os.Remove(path); err != nil {
				return err
			}
		}
	}
	return nil
}

// FileExists checks if a file or directory exists at the specified path.
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// Helper function to show files in a directory.
// For example:
//
//	util.Find(ctx, "-L", ".")
func Find(ctx context.Context, arg ...string) {
	cmd := exec.CommandContext(ctx, "/bin/find", arg...)
	cmd.Env = os.Environ()
	_ = RunCmd(ctx, cmd)
}

// Run "bazel info" and return the result as a map.
func BazelInfo(ctx context.Context, workspaceDir string, homeDir string, bazelBinary string) (map[string]string, error) {
	cmd := exec.CommandContext(ctx, bazelBinary, "info")
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "HOME="+homeDir)
	cmd.Env = append(cmd.Env, "PATH=/bin")
	cmd.Dir = workspaceDir
	output, err := RunCmdOutput(ctx, cmd)
	if err != nil {
		return nil, err
	}
	info := map[string]string{}
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		pair := strings.SplitN(scanner.Text(), ": ", 2)
		if len(pair) == 2 {
			info[pair[0]] = pair[1]
		}
	}
	return info, scanner.Err()
}

// Ensure that the command line includes a non-empty string for the flag.
func EnsureRequiredStringFlag(name string) {
	value, err := flag.CommandLine.GetString(name)
	if err != nil {
		panic(err)
	}
	if value == "" {
		fmt.Printf("required flag \"--%s\" is empty\n", name)
		flag.Usage()
		os.Exit(2)
	}
}

func GetOwnedDirectoryMarkerFileName() string {
	return "OWNED_BY_UNIVERSE_DEPLOYER.md"
}

func GetOwnedDirectoryReadmeContent() string {
	return fmt.Sprintf(`# Owned by Universe Deployer

This directory and any other directory containing a file named %s are owned by Universe Deployer.
	
**Universe Deployer may automatically replace the entire contents of owned directories at any time.**
	
Learn more at https://github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/tree/main/deployment/universe_deployer.
`, GetOwnedDirectoryMarkerFileName())
}

func WorkspaceDir() string {
	return os.Getenv("BUILD_WORKSPACE_DIRECTORY")
}

// If the provided filename is relative, and the application is running in Bazel,
// assume the filename is relative to the Bazel workspace and return the absolute path.
func AbsFromWorkspace(filename string) string {
	workspaceDir := WorkspaceDir()
	if filename != "" && workspaceDir != "" && !filepath.IsAbs(filename) {
		return filepath.Join(workspaceDir, filename)
	}
	return filename
}

func AbsFromWorkspaceList(filenames []string) []string {
	var result []string
	for _, filename := range filenames {
		result = append(result, AbsFromWorkspace(filename))
	}
	return result
}

func CopyBazelRunfiles(ctx context.Context, oldRunfilesDir string, newRunfilesDir string) error {
	if err := CopyDir(oldRunfilesDir, newRunfilesDir); err != nil {
		return err
	}
	return nil
}

func GenerateRandomAlphaNumericString(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		randomInt, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		result[i] = charset[randomInt.Int64()]
	}

	return string(result), nil
}
