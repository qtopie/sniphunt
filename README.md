# Sniphunt

Sniphunt is a high-performance code search library and CLI tool for Go. It supports both **Regex-based search** (ripgrep-like speed) and **Similarity-based code snippet search**.

It is designed to be fast, memory-efficient, and easy to integrate as a library.

## Features

- 🚀 **High Performance**: Uses `godirwalk` for fast directory traversal and a worker-pool model for concurrent processing.
- 🔍 **Regex Search**: Fast line-based regex search supporting PCRE-style syntax via `regexp2`.
- 🧬 **Similarity Search**: Find similar code snippets using Levenshtein distance.
- 📦 **Library-First**: Clean API designed to be imported into other Go projects.
- 🛠 **CLI Tool**: Powerful command-line interface with support for ignoring patterns, file extensions, and worker count control.
- 🔋 **Optimized**: Utilizes `sync.Pool` for buffer reuse and literal pre-filtering to minimize GC pressure and CPU usage.

## Installation

### CLI Tool
Install the `sniphunt` command:
```bash
go install github.com/qtopie/sniphunt/cmd/sniphunt@latest
```

### Library
Add Sniphunt to your project:
```bash
go get github.com/qtopie/sniphunt
```

## Usage

### Command Line

**Regex Search**
Search for a pattern in all `.go` files:
```bash
sniphunt -pattern "func.*main" -dir . -ext .go
```

**Similarity Search**
Find files most similar to a given snippet:
```bash
sniphunt -input snippet.java -dir ./src -ext .java
```

**Advanced Options**
```bash
sniphunt -pattern "TODO" -ignore ".git,vendor" -workers 8 -dir .
```

### As a Library

```go
import (
    "context"
    "fmt"
    "github.com/qtopie/sniphunt/pkg/search"
)

func main() {
    s := search.NewSearcher()
    s.Extensions = []string{".go"}
    
    ctx := context.Background()
    matchChan, errChan := s.Search(ctx, ".", "func.*main")
    
    for match := range matchChan {
        fmt.Printf("Found: %s:%d\n", match.Path, match.LineNum)
    }
}
```

## Project Structure

- `cmd/sniphunt`: CLI entry point.
- `pkg/search`: Core search logic (Regex & Walking).
- `pkg/similarity`: Similarity search algorithms.

## License
MIT
