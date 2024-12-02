package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gobwas/glob"
	"github.com/spf13/cobra"
)

func main() {
	var outputFile string

	var rootCmd = &cobra.Command{
		Use:   "repo-to-text [repo_path]",
		Short: "Convert a local repository into a text file",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			repoPath := args[0]

			// Add debug prints
			absPath, err := filepath.Abs(repoPath)
			if err != nil {
				fmt.Printf("Error getting absolute path: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Processing path: %s\n", absPath)

			err = convertRepoToText(repoPath, outputFile)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Repository contents saved to %s\n", outputFile)
		},
	}

	rootCmd.Flags().StringVarP(&outputFile, "output", "o", "repo_contents.txt", "Output file name")
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func convertRepoToText(repoPath, outputFile string) error {
	fmt.Printf("Starting to process repository: %s\n", repoPath) // Debug print

	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		return fmt.Errorf("repository path %s does not exist", repoPath)
	}

	// Get file tree and contents
	fmt.Println("Getting file tree...") // Debug print
	fileTree, err := getFileTreeAndContents(repoPath)
	if err != nil {
		return err
	}
	fmt.Printf("Found %d entries\n", len(fileTree)) // Debug print

	// Write output to file
	fmt.Printf("Writing to output file: %s\n", outputFile) // Debug print
	f, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer f.Close()

	writer := bufio.NewWriter(f)
	defer writer.Flush()

	writer.WriteString("# File Tree and Contents\n\n")
	for _, entry := range fileTree {
		writer.WriteString(entry)
		writer.WriteString("\n")
	}

	return nil
}

func getFileTreeAndContents(repoPath string) ([]string, error) {
	// Load gitignore patterns
	ignorePatterns, err := loadGitignorePatterns(repoPath)
	if err != nil {
		// Continue even if .gitignore can't be loaded
		ignorePatterns = []glob.Glob{}
	}

	var fileTree []string
	err = filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Skip git history files
		if strings.HasPrefix(path, ".git") {
			return nil
		}

		// Check if file matches any gitignore pattern
		relPath, _ := filepath.Rel(repoPath, path)
		for _, pattern := range ignorePatterns {
			if pattern.Match(relPath) {
				return nil // Skip this file
			}
		}

		// Add separator and file name
		fileTree = append(fileTree, "=====")
		fileTree = append(fileTree, fmt.Sprintf("## filename: %s", relPath))
		fileTree = append(fileTree, "=====")

		// Read file contents
		content, err := readFileContents(path)
		if err != nil {
			fileTree = append(fileTree, fmt.Sprintf("[Error reading file: %v]", err))
			return nil
		}
		fileTree = append(fileTree, content)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return fileTree, nil
}

func readFileContents(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	// Detect binary files by checking for null bytes
	buf := make([]byte, 512)
	n, err := f.Read(buf)
	if err != nil {
		return "", err
	}
	if bytes.Contains(buf[:n], []byte{0}) {
		return "[Binary file content omitted]", nil
	}

	// Reset the reader and read the entire file
	f.Seek(0, 0)
	scanner := bufio.NewScanner(f)
	var content strings.Builder
	for scanner.Scan() {
		content.WriteString(scanner.Text())
		content.WriteString("\n")
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}

	return content.String(), nil
}

func loadGitignorePatterns(repoPath string) ([]glob.Glob, error) {
	gitignorePath := filepath.Join(repoPath, ".gitignore")
	file, err := os.Open(gitignorePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var patterns []glob.Glob
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Convert gitignore pattern to glob pattern
		pattern := strings.Replace(line, "/", string(filepath.Separator), -1)
		g, err := glob.Compile(pattern)
		if err != nil {
			continue // Skip invalid patterns
		}
		patterns = append(patterns, g)
	}

	return patterns, scanner.Err()
}
