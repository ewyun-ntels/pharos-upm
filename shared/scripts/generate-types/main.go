package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	schemaDir    = "schema"
	goOutBase    = "types"
	srcLang      = "schema"
	targetLang   = "go"
	quicktypeBin = "quicktype"
)

func main() {
	// Get current working directory (should be shared/)
	sharedDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: %v", err)
	}

	// Clean existing types directory
	typesPath := filepath.Join(sharedDir, goOutBase)
	if _, err := os.Stat(typesPath); err == nil {
		fmt.Println("Cleaning existing Go types directory...")
		if err := os.RemoveAll(typesPath); err != nil {
			log.Fatalf("Failed to remove types directory: %v", err)
		}
	}

	// Create types directory
	if err := os.MkdirAll(typesPath, 0755); err != nil {
		log.Fatalf("Failed to create types directory: %v", err)
	}

	// Get domain directories
	schemaPath := filepath.Join(sharedDir, schemaDir)
	entries, err := os.ReadDir(schemaPath)
	if err != nil {
		log.Fatalf("Failed to read schema directory: %v", err)
	}

	// Process each domain
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		domain := entry.Name()
		fmt.Printf("Processing domain: %s\n", domain)

		domainDir := filepath.Join(schemaPath, domain)
		goDomainDir := filepath.Join(typesPath, domain)

		// Create domain output directory
		if err := os.MkdirAll(goDomainDir, 0755); err != nil {
			log.Printf("Failed to create domain directory %s: %v", domain, err)
			continue
		}

		// Get schema files
		schemaFiles, err := getSchemaFiles(domainDir)
		if err != nil {
			log.Printf("Failed to get schema files for %s: %v", domain, err)
			continue
		}

		if len(schemaFiles) == 0 {
			continue
		}

		// Generate Go types
		outputFile := filepath.Join(goDomainDir, "types.go")
		if err := generateGoTypes(schemaFiles, outputFile, domain); err != nil {
			log.Printf("Failed to generate types for %s: %v", domain, err)
			continue
		}

		fmt.Printf("Generated unified Go types package for domain: %s\n", domain)
	}

	fmt.Println("Go schema generation complete!")
	fmt.Println("")
	fmt.Println("To generate TypeScript types, run: pnpm generate:types")
}

func getSchemaFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".schema") || strings.Contains(name, ".schema") {
			files = append(files, filepath.Join(dir, name))
		}
	}
	return files, nil
}

func generateGoTypes(schemaFiles []string, outputFile, packageName string) error {
	args := []string{
		"--src-lang", srcLang,
	}
	args = append(args, schemaFiles...)
	args = append(args,
		"--lang", targetLang,
		"--out", outputFile,
		"--package", packageName,
		"--omit-empty",
	)

	cmd := exec.Command(quicktypeBin, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
