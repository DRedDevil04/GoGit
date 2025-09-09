package main

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"log"
	"os"
)

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
	content := out.String()[headerEnd+11:]
	return content, nil
}
