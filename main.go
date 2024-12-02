package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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
			err := convertRepoToText(repoPath, outputFile)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Repository contents saved to %s\n", outputFile)
		},
	}

	rootCmd.Flags().StringVarP(&outputFile, "output", "o", "repo_contents.txt", "Output file name")
	rootCmd.Execute()
}

func convertRepoToText(repoPath, outputFile string) error {
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		return fmt.Errorf("repository path %s does not exist", repoPath)
	}

	// Get file tree and contents
	fileTree, err := getFileTreeAndContents(repoPath)
	if err != nil {
		return err
	}

	// Get Git history
	gitHistory, err := getGitHistory(repoPath)
	if err != nil {
		return err
	}

	// Write output to file
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

	writer.WriteString("\n# Git History\n\n")
	writer.WriteString(gitHistory)

	return nil
}

func getFileTreeAndContents(repoPath string) ([]string, error) {
	var fileTree []string
	err := filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Skip directories
		if info.IsDir() {
			return nil
		}
		relPath, _ := filepath.Rel(repoPath, path)
		fileTree = append(fileTree, fmt.Sprintf("## %s", relPath))

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

func getGitHistory(repoPath string) (string, error) {
	cmd := exec.Command("git", "-C", repoPath, "log", "--oneline")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("failed to get Git history: %w", err)
	}
	return out.String(), nil
}
