package main

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"fmt"
	"io"
	"log"
	"os"
)

type TreeEntry struct {
	Mode string
	Type string
	SHA  string
	Name string
}

// Usage: your_program.sh <command> <arg1> <arg2> ...
func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Fprintf(os.Stderr, "Logs from your program will appear here!\n")

	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: mygit <command> [<args>...]\n")
		os.Exit(1)
	}

	switch command := os.Args[1]; command {
	case "init":
		// Uncomment this block to pass the first stage!

		for _, dir := range []string{".git", ".git/objects", ".git/refs"} {
			if err := os.MkdirAll(dir, 0755); err != nil {
				fmt.Fprintf(os.Stderr, "Error creating directory: %s\n", err)
			}
		}

		headFileContents := []byte("ref: refs/heads/main\n")
		if err := os.WriteFile(".git/HEAD", headFileContents, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing file: %s\n", err)
		}

		fmt.Println("Initialized git directory")

	case "cat-file":
		if len(os.Args) < 4 {
			fmt.Fprintf(os.Stderr, "usage: git <command> [<args>...]\n")
			os.Exit(1)
		}
		if os.Args[2] != "-p" {
			fmt.Fprintf(os.Stderr, "Unknown flag %s\n", os.Args[2])
			os.Exit(1)
		}
		shaHash := os.Args[3]
		content, err := getObjectContent(shaHash)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error retrieving object: %s\n", err)
			os.Exit(1)
		}
		fmt.Print(content)
	case "hash-object":
		if len(os.Args) < 4 {
			fmt.Fprintf(os.Stderr, "usage: git <command> [<args>...]\n")
			os.Exit(1)
		}
		if os.Args[2] != "-w" {
			fmt.Fprintf(os.Stderr, "Unknown flag %s\n", os.Args[2])
			os.Exit(1)
		}
		shaHash, err := writeObjectToGit(os.Args[3])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing object: %s\n", err)
			os.Exit(1)
		}
		fmt.Println(shaHash)

	case "ls-tree":
		if len(os.Args) < 4 {
			fmt.Fprintf(os.Stderr, "usage: git <command> [<args>...]\n")
			os.Exit(1)
		}
		shaHash := os.Args[2]
		entries, err := readTreeObject(shaHash)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading tree object: %s\n", err)
			os.Exit(1)
		}
		for _, entry := range entries {
			fmt.Println(entry)
		}
	default:
		fmt.Fprintf(os.Stderr, "Unknown command %s\n", command)
		os.Exit(1)
	}
}

func getObjectContent(shaHash string) (string, error) {
	// Gets Content of Specified SHA Object
	if len(shaHash) != 40 {
		return "", fmt.Errorf("invalid len of the hash")
	}
	dir := fmt.Sprintf(".git/objects/%s", shaHash[:2])
	file := fmt.Sprintf("%s/%s", dir, shaHash[2:])
	data, err := os.ReadFile(file)
	if err != nil {
		return "", err
	}
	r, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	var out bytes.Buffer
	io.Copy(&out, r)
	headerEnd := bytes.IndexByte(out.Bytes(), 0)
	if headerEnd == -1 {
		log.Fatal("no null terminator found in header")
	}
	content := out.String()[headerEnd+1:]
	return content, nil
}

func writeObjectToGit(filePath string) (string, error) {
	// read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("error reading file: %w", err)
	}

	// create blob header + data
	blobHeader := fmt.Sprintf("blob %d\x00", len(data))
	blobData := append([]byte(blobHeader), data...)

	// compute SHA-1 of blob
	h := sha1.New()
	h.Write(blobData)
	shaHash := fmt.Sprintf("%x", h.Sum(nil))

	// compress blobData
	var compressedData bytes.Buffer
	w := zlib.NewWriter(&compressedData)
	if _, err := w.Write(blobData); err != nil {
		return "", fmt.Errorf("error compressing: %w", err)
	}
	w.Close()

	// create directory for object
	dir := fmt.Sprintf(".git/objects/%s", shaHash[:2])
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("error creating dir: %w", err)
	}

	// write compressed blob
	filePath = fmt.Sprintf("%s/%s", dir, shaHash[2:])
	if err := os.WriteFile(filePath, compressedData.Bytes(), 0644); err != nil {
		return "", fmt.Errorf("error writing object: %w", err)
	}

	return shaHash, nil
}

func readTreeObject(shaHash string) ([]string, error) {
	content, err := getObjectContent(shaHash)
	if err != nil {
		return nil, err
	}
	var entries []string
	for len(content) > 0 {
		entry, err := getTreeEntry(&content)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry.Name)
	}
	return entries, nil
}

func getTreeEntry(content *string) (TreeEntry, error) {
	var entry TreeEntry
	// Parse mode :  <mode> <name>\0<20_byte_sha>
	modeEnd := bytes.IndexByte([]byte(*content), ' ')
	if modeEnd == -1 {
		return entry, fmt.Errorf("invalid tree entry: no space after mode")
	}
	entry.Mode = (*content)[:modeEnd]
	*content = (*content)[modeEnd+1:]

	// Parse name
	nameEnd := bytes.IndexByte([]byte(*content), 0)
	if nameEnd == -1 {
		return entry, fmt.Errorf("invalid tree entry: no null terminator after name")
	}
	entry.Name = (*content)[:nameEnd]
	*content = (*content)[nameEnd+1:]

	// Parse SHA
	if len(*content) < 20 {
		return entry, fmt.Errorf("invalid tree entry: SHA is too short")
	}
	entry.SHA = (*content)[:20]
	if len(*content) > 20 {
		*content = (*content)[20:]
	} else {
		*content = ""
	}
	return entry, nil
}
