package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	var (
		protoDir    string
		outDir      string
		goOut       string
		grpcOut     string
		goVersion   string
		grpcVersion string
	)

	flag.StringVar(&protoDir, "proto-dir", ".", "Directory containing .proto files")
	flag.StringVar(&outDir, "out-dir", ".", "Output directory for generated code")
	flag.StringVar(&goOut, "go-out", ".", "Output directory for go_out (relative to out-dir by default)")
	flag.StringVar(&grpcOut, "grpc-out", ".", "Output directory for go-grpc_out (relative to out-dir by default)")
	flag.StringVar(&goVersion, "go-version", "latest", "Version of protoc-gen-go to use")
	flag.StringVar(&grpcVersion, "grpc-version", "latest", "Version of protoc-gen-go-grpc to use")
	flag.Parse()

	if err := run(protoDir, outDir, goOut, grpcOut, goVersion, grpcVersion); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func run(protoDir, outDir, goOut, grpcOut, goVersion, grpcVersion string) error {
	// 1. Resolve absolute paths
	absProtoDir, err := filepath.Abs(protoDir)
	if err != nil {
		return fmt.Errorf("failed to resolving proto dir: %w", err)
	}
	absOutDir, err := filepath.Abs(outDir)
	if err != nil {
		return fmt.Errorf("failed to resolving out dir: %w", err)
	}

	// Ensure output directory exists
	if err := os.MkdirAll(absOutDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// 2. Install/Check Plugins
	if err := installPlugins(goVersion, grpcVersion); err != nil {
		return fmt.Errorf("failed to install plugins: %w", err)
	}

	// 3. Find Proto Files
	protoFiles, err := findProtoFiles(absProtoDir)
	if err != nil {
		return fmt.Errorf("failed to find proto files: %w", err)
	}
	if len(protoFiles) == 0 {
		log.Println("No .proto files found in", absProtoDir)
		return nil
	}

	// 4. Generate Code
	if err := generateCode(protoFiles, absProtoDir, absOutDir, goOut, grpcOut); err != nil {
		return fmt.Errorf("failed to generate code: %w", err)
	}

	log.Println("Generation completed successfully!")
	return nil
}

func installPlugins(goVersion, grpcVersion string) error {
	log.Printf("Ensuring plugins are installed: protoc-gen-go@%s, protoc-gen-go-grpc@%s", goVersion, grpcVersion)

	// Install protoc-gen-go
	cmd := exec.Command("go", "install", "google.golang.org/protobuf/cmd/protoc-gen-go@"+goVersion)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install protoc-gen-go: %w", err)
	}

	// Install protoc-gen-go-grpc
	cmd = exec.Command("go", "install", "google.golang.org/grpc/cmd/protoc-gen-go-grpc@"+grpcVersion)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install protoc-gen-go-grpc: %w", err)
	}

	return nil
}

func findProtoFiles(dir string) ([]string, error) {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".proto") {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func generateCode(files []string, protoInclude, outDir, goOut, grpcOut string) error {
	// Construct protoc command
	// protoc --go_out=<out> --go_out_opt=paths=source_relative --go-grpc_out=<out> --go-grpc_opt=paths=source_relative -I <proto-dir> <files>

	args := []string{
		"-I", protoInclude,
		"--go_out=" + outDir,
		"--go_opt=paths=source_relative",
		"--go-grpc_out=" + outDir,
		"--go-grpc_opt=paths=source_relative",
	}

	// Append all proto files
	args = append(args, files...)

	cmd := exec.Command("protoc", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Add GOPATH/bin to PATH for this command
	goBin, err := getGoBin()
	if err == nil {
		newPath := goBin + string(os.PathListSeparator) + os.Getenv("PATH")
		cmd.Env = append(os.Environ(), "PATH="+newPath)
	} else {
		log.Printf("Warning: could not determine GOPATH/bin: %v", err)
	}

	log.Printf("Running protoc...")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("protoc execution failed: %w", err)
	}

	return nil
}

func getGoBin() (string, error) {
	// First try GOBIN env var
	if goBin := os.Getenv("GOBIN"); goBin != "" {
		return goBin, nil
	}

	// Then try GOPATH
	goPath := os.Getenv("GOPATH")
	if goPath == "" {
		// If GOPATH is empty, default to $HOME/go
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		goPath = filepath.Join(home, "go")
	}

	return filepath.Join(goPath, "bin"), nil
}
