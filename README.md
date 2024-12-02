# repo-to-text

A CLI tool that converts a local repository into a single text file, respecting .gitignore patterns.

## Features

- Converts repository contents into a single text file
- Respects .gitignore patterns
- Skips binary files
- Formats output with clear file separators

## Installation

### Prerequisites

- Go 1.20 or higher

### Building from source

```bash
# Clone the repository
git clone https://github.com/yourusername/repo-to-text.git
cd repo-to-text
# Download dependencies
go mod download
# Build the application
go build
# Or install it globally
go install
```

### Cross-compilation

Build for different platforms:

```bash
# For Windows
GOOS=windows GOARCH=amd64 go build -o repo-to-text.exe
# For macOS
GOOS=darwin GOARCH=amd64 go build -o repo-to-text-mac
# For Linux
GOOS=linux GOARCH=amd64 go build -o repo-to-text-linux
```

## Usage

Basic usage:

```bash
repo-to-text [repository_path] -o [output_file]
```

Examples:

```bash
# Convert current directory
repo-to-text . -o output.txt
# Convert specific repository
repo-to-text ~/my-project -o result.txt
# Use default output name (repo_contents.txt)
repo-to-text .
```

### Options

- `-o, --output`: Specify output file name (default: "repo_contents.txt")
- `-h, --help`: Display help information

## Output Format

The output file will contain:

# File Tree and Contents

=====

# path/to/file1.txt

=====

# [file contents]

=====

# path/to/file2.go

=====
[file contents]
=====
