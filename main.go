package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	// Parse command line arguments
	var notebookName string
	flag.StringVar(&notebookName, "notebook", "Imported Notebook", "Notebook name")
	flag.StringVar(&notebookName, "n", "Imported Notebook", "Short form for notebook name")
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("Usage: md2nsx [options] <markdown_folder>")
		fmt.Println("Options:")
		fmt.Println("  -n, --notebook <name>  Set notebook name (default: \"Imported Notebook\")")
		fmt.Println("Examples:")
		fmt.Println("  md2nsx ./markdown-files")
		fmt.Println("  md2nsx -n \"My Notes\" ./markdown-files")
		fmt.Println("  md2nsx --notebook \"My Notes\" ./markdown-files")
		fmt.Println("")
		fmt.Println("Important: Flags must come BEFORE folder argument")
		fmt.Println("  [OK] Correct: md2nsx --notebook \"Name\" ./folder")
		fmt.Println("  [X]  Wrong:   md2nsx ./folder --notebook \"Name\"")
		os.Exit(1)
	}

	markdownFolder := args[0]

	// Validate input folder
	if _, err := os.Stat(markdownFolder); os.IsNotExist(err) {
		log.Fatalf("Error: Markdown folder '%s' does not exist", markdownFolder)
	}

	// Create converter instance
	converter := NewNSXConverter()

	// Perform batch conversion
	err := converter.BatchConvert(markdownFolder, notebookName)
	if err != nil {
		log.Fatalf("Error during conversion: %v", err)
	}

	fmt.Printf("Successfully converted markdown files in '%s' to NSX format\n", markdownFolder)
}
