package main

import (
	"archive/zip"
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	chromahtml "github.com/alecthomas/chroma/formatters/html"
	"github.com/gabriel-vasile/mimetype"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer"
	rendererhtml "github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
)

// Pre-compiled regex patterns for better performance
var (
	imagePattern = regexp.MustCompile(`!\[(.*?)\]\((.*?)(?:\s+"(.*?)")?\)`)
	linkPattern  = regexp.MustCompile(`\[(.*?)\]\((.*?)\.(pdf|doc|docx|txt|zip|rar|md|csv|xls|xlsx)\)`)

	checkboxPattern = regexp.MustCompile(`<input[^>]*type="checkbox"[^>]*>`)
)

// NSXConverter handles the conversion from Markdown to NSX format
type NSXConverter struct {
	processedImages []ProcessedImage
	attachments     map[string]Attachment
}

// ProcessedImage represents a processed image file
type ProcessedImage struct {
	MD5Hash      string
	ImageDataB64 string
}

// Attachment represents a file attachment
type Attachment struct {
	MD5    string `json:"md5"`
	Name   string `json:"name"`
	Size   int64  `json:"size"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Type   string `json:"type"`
	CTime  int64  `json:"ctime"`
	Ref    string `json:"ref"`
}

// Note represents a note in NSX format
type Note struct {
	Category   string                `json:"category"`
	ParentID   string                `json:"parent_id"`
	Title      string                `json:"title"`
	Thumb      *string               `json:"thumb,omitempty"`
	MTime      int64                 `json:"mtime"`
	CTime      int64                 `json:"ctime"`
	Latitude   float64               `json:"latitude"`
	Longitude  float64               `json:"longitude"`
	Encrypt    bool                  `json:"encrypt"`
	Attachment map[string]Attachment `json:"attachment"`
	Brief      string                `json:"brief"`
	Content    string                `json:"content"`
	Tag        []string              `json:"tag"`
}

// Notebook represents a notebook in NSX format
type Notebook struct {
	Category string `json:"category"`
	ParentID string `json:"parent_id"`
	Title    string `json:"title"`
}

// NotebookConfig represents the notebook configuration
type NotebookConfig struct {
	Note     []string `json:"note"`
	Notebook []string `json:"notebook"`
}

// NewNSXConverter creates a new NSX converter instance
func NewNSXConverter() *NSXConverter {
	return &NSXConverter{
		processedImages: make([]ProcessedImage, 0),
		attachments:     make(map[string]Attachment),
	}
}

// BatchConvert converts all markdown files in a folder to NSX format
func (c *NSXConverter) BatchConvert(mdFolder, notebookName string) error {
	outputDir := "temp_nsx_output"

	// Generate a better output filename
	var outputNSXPath string
	if strings.HasSuffix(mdFolder, "/") || strings.HasSuffix(mdFolder, "\\") {
		// If folder ends with slash, use the parent folder name
		cleanFolder := strings.TrimSuffix(strings.TrimSuffix(mdFolder, "/"), "\\")
		outputNSXPath = cleanFolder + ".nsx"
	} else {
		// Otherwise use the folder name as is
		outputNSXPath = mdFolder + ".nsx"
	}

	// Clean and create temporary directory
	if err := os.RemoveAll(outputDir); err != nil {
		return fmt.Errorf("failed to clean output directory: %w", err)
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	defer func() {
		if err := os.RemoveAll(outputDir); err != nil {
			log.Printf("Warning: Could not clean up temporary directory: %v", err)
		}
	}()

	// Find all markdown files
	mdFiles, err := filepath.Glob(filepath.Join(mdFolder, "*.md"))
	if err != nil {
		return fmt.Errorf("failed to find markdown files: %w", err)
	}

	if len(mdFiles) == 0 {
		return fmt.Errorf("no markdown files found in %s", mdFolder)
	}

	fmt.Printf("Found %d markdown files to convert\n", len(mdFiles))

	// Generate notebook ID
	notebookID := "nb_" + c.generateMD5Hash(notebookName)
	fmt.Printf("Using notebook ID: %s\n", notebookID)

	// Convert each file
	for _, mdFile := range mdFiles {
		fmt.Printf("Converting %s...\n", filepath.Base(mdFile))

		// Read markdown content
		mdContent, err := c.readFileWithEncoding(mdFile)
		if err != nil {
			log.Printf("Error reading %s: %v", mdFile, err)
			continue
		}

		// Process images and attachments
		processedContent, err := c.processAttachments(mdFile, mdContent)
		if err != nil {
			log.Printf("Error processing attachments for %s: %v", mdFile, err)
			continue
		}

		// Create note object
		title := strings.TrimSuffix(filepath.Base(mdFile), ".md")
		if title == "" {
			title = "Untitled"
		}

		note, titleBase64, err := c.createNote(title, processedContent, notebookID)
		if err != nil {
			log.Printf("Error creating note for %s: %v", mdFile, err)
			continue
		}

		// Create note file
		noteFilename := "note_" + titleBase64
		noteFilePath := filepath.Join(outputDir, noteFilename)

		noteData, err := json.MarshalIndent(note, "", "  ")
		if err != nil {
			log.Printf("Error marshaling note %s: %v", mdFile, err)
			continue
		}

		if err := os.WriteFile(noteFilePath, noteData, 0644); err != nil {
			log.Printf("Error writing note file %s: %v", noteFilePath, err)
			continue
		}

		fmt.Printf("  Successfully converted: %s -> %s\n", filepath.Base(mdFile), noteFilename)
	}

	// Package into NSX file
	fmt.Printf("Packaging into %s\n", outputNSXPath)
	if err := c.packageNSX(outputDir, outputNSXPath, notebookName, notebookID); err != nil {
		return fmt.Errorf("failed to package NSX: %w", err)
	}

	fmt.Printf("Successfully converted %d files to %s\n", len(mdFiles), outputNSXPath)
	return nil
}

// readFileWithEncoding reads a file with UTF-8 encoding
func (c *NSXConverter) readFileWithEncoding(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	// Check if it's valid UTF-8
	if !isValidUTF8(data) {
		fmt.Printf("Warning: File %s may contain non-UTF-8 characters\n", filePath)
	}

	return string(data), nil
}

// isValidUTF8 checks if the data is valid UTF-8
func isValidUTF8(data []byte) bool {
	return utf8.Valid(data)
}

// processAttachments processes images and file attachments in markdown content
func (c *NSXConverter) processAttachments(mdFile, mdContent string) (string, error) {
	// Process image links - support both basic and title formats
	imageMatches := imagePattern.FindAllStringSubmatch(mdContent, -1)

	// Process file links
	linkMatches := linkPattern.FindAllStringSubmatch(mdContent, -1)

	// Process all matches
	for _, match := range imageMatches {
		if len(match) >= 3 {
			altText, link := match[1], match[2]
			// Use title as alt text if available, otherwise use alt text
			if len(match) >= 4 && match[3] != "" {
				altText = match[3]
			}
			if err := c.processAttachment(mdFile, "image", altText, link, &mdContent); err != nil {
				fmt.Printf("Warning: Failed to process image %s: %v", link, err)
			}
		}
	}

	for _, match := range linkMatches {
		if len(match) >= 4 {
			text, link, ext := match[1], match[2], match[3]
			fullLink := link + "." + ext
			if err := c.processAttachment(mdFile, "link", text, fullLink, &mdContent); err != nil {
				fmt.Printf("Warning: Failed to process link %s: %v", fullLink, err)
			}
		}
	}

	return mdContent, nil
}

// processAttachment processes a single attachment
func (c *NSXConverter) processAttachment(mdFile, matchType, altText, link string, mdContent *string) error {
	// Find the file
	filePath, err := c.findFile(mdFile, link)
	if err != nil {
		return fmt.Errorf("file not found: %s", link)
	}

	// Read file data
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	// Calculate MD5 hash
	md5Hash := c.generateMD5Hash(string(fileData))

	// Detect MIME type
	mimeType := mimetype.Detect(fileData).String()

	isImage := false
	switch matchType {
	case "image":
		isImage = true
	case "link":
		isImage = strings.HasPrefix(mimeType, "image/")
	}

	width, height := 0, 0
	if isImage {
		width, height = 400, 300
	}

	timestamp := time.Now().Unix()

	originalFilename := filepath.Base(filePath)
	filenameWithTimestamp := fmt.Sprintf("%s%d", originalFilename, timestamp)
	filenameB64 := base64.StdEncoding.EncodeToString([]byte(filenameWithTimestamp))
	fileKey := "file_" + filenameB64

	refContent := fmt.Sprintf("%d%s", timestamp, originalFilename)
	refB64 := base64.StdEncoding.EncodeToString([]byte(refContent))

	var htmlTag, originalMD string
	if isImage {
		htmlTag = fmt.Sprintf(`<img class="syno-notestation-image-object" src="webman/3rdparty/NoteStation/images/transparent.gif" border="0" width="%d" ref="%s" adjust="true"/>`, width, refB64)
		originalMD = fmt.Sprintf("![%s](%s)", altText, link)
	} else {
		displayText := altText
		if displayText == "" {
			switch matchType {
			case "link":
				displayText = "Attachment"
			default:
				displayText = "File"
			}
		}
		htmlTag = fmt.Sprintf(`<a href="%s" target="_blank">%s</a>`, refB64, displayText)
		originalMD = fmt.Sprintf("[%s](%s)", altText, link)
	}

	*mdContent = strings.ReplaceAll(*mdContent, originalMD, htmlTag)

	c.attachments[fileKey] = Attachment{
		MD5:    md5Hash,
		Name:   originalFilename,
		Size:   int64(len(fileData)),
		Width:  width,
		Height: height,
		Type:   mimeType,
		CTime:  timestamp,
		Ref:    refB64,
	}

	if isImage {
		imageDataB64 := base64.StdEncoding.EncodeToString(fileData)
		c.processedImages = append(c.processedImages, ProcessedImage{
			MD5Hash:      md5Hash,
			ImageDataB64: imageDataB64,
		})
	}

	fmt.Printf("  Processed %s: %s -> %s (MIME: %s)\n", matchType, filepath.Base(filePath), fileKey, mimeType)
	return nil
}

// findFile searches for a file in the current directory and parent directories
func (c *NSXConverter) findFile(mdFile, link string) (string, error) {
	// Check if file exists in current directory
	if _, err := os.Stat(link); err == nil {
		return link, nil
	}

	// Search in parent directory
	parentDir := filepath.Dir(mdFile)
	err := filepath.Walk(parentDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.Contains(filepath.Base(path), filepath.Base(link)) {
			link = path
			return filepath.SkipAll
		}
		return nil
	})

	if err != nil && err != filepath.SkipAll {
		return "", fmt.Errorf("file not found: %s", link)
	}

	return link, nil
}

type customCodeSpanRenderer struct{}

func (r *customCodeSpanRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindCodeSpan, r.renderCodeSpan)
}

func (r *customCodeSpanRenderer) renderCodeSpan(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		_, _ = w.WriteString(`<code 123 style="color: #e83e8c; background-color: #f8f9fa; padding: 2px 4px; border-radius: 3px;">`)
		for c := n.FirstChild(); c != nil; c = c.NextSibling() {
			segment := c.(*ast.Text).Segment
			value := segment.Value(source)
			if bytes.HasSuffix(value, []byte("\n")) {
				_, _ = w.Write(value[:len(value)-1])
				_, _ = w.Write([]byte(" "))
			} else {
				_, _ = w.Write(value)
			}
		}
		return ast.WalkSkipChildren, nil
	}
	_, _ = w.WriteString("</code>")
	return ast.WalkContinue, nil
}

type customCodeBlockPreWrapper struct{}

func (r *customCodeBlockPreWrapper) Start(code bool, styleAttr string) string {
	return `<code style="white-space: pre; font-family: Menlo, Monaco, Consolas, monospace; display: inline-block;">`
}

func (r *customCodeBlockPreWrapper) End(code bool) string {
	return `</code>`
}

type customBlockquoteRenderer struct{}

func (r *customBlockquoteRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindBlockquote, r.renderBlockquote)
}

func (r *customBlockquoteRenderer) renderBlockquote(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		_, _ = w.WriteString(`<blockquote style="margin: 1em 0; padding: 0.5em 1em; border-left: 4px solid #ccc; color: #666;">`)
	} else {
		_, _ = w.WriteString("</blockquote>")
	}
	return ast.WalkContinue, nil
}

var (
	styleCodeBlock = `
	background-color: #f6f8fa !important;border: 1px solid #d1d5da;padding: 16px;
	margin: 10px 0;border-radius: 12px;box-shadow: 0 1px 3px rgba(0,0,0,0.12), 0 1px 2px rgba(0,0,0,0.24);
	display:inline-block; overflow-x: auto;max-width: 100%;min-width: 60%;
	`
)

// markdownToHTML converts markdown content to HTML
func (c *NSXConverter) markdownToHTML(_, mdContent string) (string, error) {
	// Define extensions separately for readability
	extensions := []goldmark.Extender{
		extension.GFM,
		extension.Footnote,
		extension.DefinitionList,
		extension.Linkify,
		extension.Typographer,
		highlighting.NewHighlighting(
			highlighting.WithStyle("github"),
			highlighting.WithGuessLanguage(true),
			highlighting.WithWrapperRenderer(func(w util.BufWriter, context highlighting.CodeBlockContext, entering bool) {
				if entering {
					language, _ := context.Language()
					_, _ = w.WriteString(`<div style="` + styleCodeBlock + `"><pre class="language-` + string(language) + `">`)
				} else {
					_, _ = w.WriteString(`</pre></div>`)
				}
			}),
			highlighting.WithCodeBlockOptions(func(ctx highlighting.CodeBlockContext) []chromahtml.Option {
				return []chromahtml.Option{
					chromahtml.WithClasses(false),
					chromahtml.WithLineNumbers(true),
					chromahtml.WithAllClasses(false),
					chromahtml.TabWidth(4),
					chromahtml.WithPreWrapper(&customCodeBlockPreWrapper{}),
				}
			}),
		),
	}

	// Define renderer options separately
	rendererOptions := []renderer.Option{
		rendererhtml.WithXHTML(),
		rendererhtml.WithUnsafe(),
		rendererhtml.WithHardWraps(),
		renderer.WithNodeRenderers(
			util.Prioritized(&customCodeSpanRenderer{}, 100),
			util.Prioritized(&customBlockquoteRenderer{}, 100),
		),
	}

	md := goldmark.New(
		goldmark.WithExtensions(extensions...),
		goldmark.WithRendererOptions(rendererOptions...),
	)

	var buf bytes.Buffer
	if err := md.Convert([]byte(mdContent), &buf); err != nil {
		return "", err
	}
	htmlContent := buf.String()
	htmlContent = c.processTodoLists(htmlContent)

	return htmlContent, nil
}

var (
	htmlCheckboxChecked = `<input class="syno-notestation-editor-checkbox syno-notestation-editor-checkbox-checked note-station-checkbox-checked" ` +
		`src="webman/3rdparty/NoteStation/images/transparent.gif" type="image" data-mce-contenteditable="false" />`

	htmlCheckboxUnchecked = `<input class="syno-notestation-editor-checkbox note-station-checkbox" src="webman/3rdparty/NoteStation/images/transparent.gif" type="image" />`
)

// processTodoLists converts markdown todo lists to Note Station format
func (c *NSXConverter) processTodoLists(htmlContent string) string {
	matches := checkboxPattern.FindAllString(htmlContent, -1)
	for _, match := range matches {
		if strings.Contains(match, `checked=""`) {
			// checked
			htmlContent = strings.ReplaceAll(htmlContent, match, htmlCheckboxChecked)
		} else {
			// unchecked
			htmlContent = strings.ReplaceAll(htmlContent, match, htmlCheckboxUnchecked)
		}
	}
	return htmlContent
}

// createNote creates a note object
func (c *NSXConverter) createNote(title, markdownContent, parentID string) (*Note, string, error) {
	// Base64 encode title
	titleBase64 := base64.StdEncoding.EncodeToString([]byte(title))

	// Ensure content is not empty
	if strings.TrimSpace(markdownContent) == "" {
		markdownContent = "Empty note"
	}

	// Generate timestamp
	currentTime := time.Now().Unix()

	// Find thumbnail
	var thumb *string
	for fileKey, attachment := range c.attachments {
		if strings.HasPrefix(attachment.Type, "image/") {
			thumb = &fileKey
			break
		}
	}

	// Generate brief from original markdown content (before HTML conversion)
	brief := c.generateBriefFromMarkdown(markdownContent)

	// Convert markdown to HTML for the content
	htmlContent, err := c.markdownToHTML(title, markdownContent)
	if err != nil {
		return nil, "", fmt.Errorf("failed to convert markdown to HTML: %w", err)
	}

	note := &Note{
		Category:   "note",
		ParentID:   parentID,
		Title:      title,
		Thumb:      thumb,
		MTime:      currentTime,
		CTime:      currentTime,
		Latitude:   0,
		Longitude:  0,
		Encrypt:    false,
		Attachment: c.attachments,
		Brief:      brief,
		Content:    htmlContent,
		Tag:        []string{},
	}

	return note, titleBase64, nil
}

// generateBriefFromMarkdown generates a clean brief from markdown content
func (c *NSXConverter) generateBriefFromMarkdown(markdownContent string) string {
	plainText := c.cleanBrief(markdownContent)

	// Limit to 100 characters
	if len(plainText) > 100 {
		plainText = plainText[:100] + "..."
	}

	return plainText
}

// cleanBrief cleans up brief text
func (c *NSXConverter) cleanBrief(brief string) string {
	// Replace all whitespace characters with single spaces
	replacer := strings.NewReplacer(
		"\n", " ",
		"\r", " ",
		"\t", " ",
	)
	brief = replacer.Replace(brief)

	// Remove multiple consecutive spaces
	for strings.Contains(brief, "  ") {
		brief = strings.ReplaceAll(brief, "  ", " ")
	}

	return strings.TrimSpace(brief)
}

// packageNSX packages the converted files into NSX format
func (c *NSXConverter) packageNSX(outputDir, outputNSXPath, notebookName, notebookID string) error {
	zipFile, err := os.Create(outputNSXPath)
	if err != nil {
		return fmt.Errorf("failed to create NSX file: %w", err)
	}
	defer func() {
		if closeErr := zipFile.Close(); closeErr != nil {
			log.Printf("Warning: Failed to close zip file: %v", closeErr)
		}
	}()

	zipWriter := zip.NewWriter(zipFile)
	defer func() {
		if closeErr := zipWriter.Close(); closeErr != nil {
			log.Printf("Warning: Failed to close zip writer: %v", closeErr)
		}
	}()

	// Add note files
	noteFiles, err := filepath.Glob(filepath.Join(outputDir, "note_*"))
	if err != nil {
		return fmt.Errorf("failed to find note files: %w", err)
	}

	noteIDs := make([]string, 0)
	for _, noteFile := range noteFiles {
		noteID := filepath.Base(noteFile)
		noteIDs = append(noteIDs, noteID)

		// Read note data
		noteData, err := os.ReadFile(noteFile)
		if err != nil {
			log.Printf("Error reading note file %s: %v", noteFile, err)
			continue
		}

		// Add to zip
		writer, err := zipWriter.Create(noteID)
		if err != nil {
			log.Printf("Error creating zip entry %s: %v", noteID, err)
			continue
		}

		if _, err := writer.Write(noteData); err != nil {
			log.Printf("Error writing note to zip %s: %v", noteID, err)
			continue
		}
	}

	// Add image files
	for _, processedImage := range c.processedImages {
		fileKey := "file_" + processedImage.MD5Hash
		imageData, err := base64.StdEncoding.DecodeString(processedImage.ImageDataB64)
		if err != nil {
			log.Printf("Error decoding image data for %s: %v", fileKey, err)
			continue
		}

		writer, err := zipWriter.Create(fileKey)
		if err != nil {
			log.Printf("Error creating zip entry %s: %v", fileKey, err)
			continue
		}

		if _, err := writer.Write(imageData); err != nil {
			log.Printf("Error writing image to zip %s: %v", fileKey, err)
			continue
		}

		fmt.Printf("  Processed image: %s\n", fileKey)
	}

	// Add notebook
	notebook := Notebook{
		Category: "notebook",
		ParentID: "",
		Title:    notebookName,
	}

	notebookData, err := json.Marshal(notebook)
	if err != nil {
		return fmt.Errorf("failed to marshal notebook: %w", err)
	}

	writer, err := zipWriter.Create(notebookID)
	if err != nil {
		return fmt.Errorf("failed to create notebook entry: %w", err)
	}

	if _, err := writer.Write(notebookData); err != nil {
		return fmt.Errorf("failed to write notebook: %w", err)
	}

	// Add config
	config := NotebookConfig{
		Note:     noteIDs,
		Notebook: []string{notebookID},
	}

	configData, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	writer, err = zipWriter.Create("config.json")
	if err != nil {
		return fmt.Errorf("failed to create config entry: %w", err)
	}

	if _, err := writer.Write(configData); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// generateMD5Hash generates MD5 hash of a string
func (c *NSXConverter) generateMD5Hash(input string) string {
	hash := md5.Sum([]byte(input))
	return fmt.Sprintf("%x", hash)
}
