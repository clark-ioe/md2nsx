# Attachment Example Notes

This is an example note that demonstrates attachment processing functionality, including images and file links.

## Image Attachments

### Basic Image
![Sample Image](sample-image.png)

### Image Link
Click the image to jump to the link: [![GitHub Logo](sample-image.png)](https://github.com)

## File Attachments

### Document Files
- [Project Requirements Document](documentation.pdf) - PDF format requirements document
- [Technical Specification](specification.docx) - Word format specification
- [User Manual](manual.txt) - Plain text format user manual

### Data Files
- [Sales Data](sales-data.xlsx) - Excel format sales data
- [Customer Information](customer-list.csv) - CSV format customer list
- [Project Backup](project-backup.zip) - Compressed format project backup

### Code Files
- [Source Code](source-code.md) - Markdown format source code documentation
- [Configuration File](config.json) - JSON format configuration file

## Mixed Content

### Image and Text Combination
In the image below, we can see the overall system architecture:

![System Architecture](sample-image.png)

This architecture diagram shows:
- Frontend interface layer
- Business logic layer
- Data access layer
- Database layer

### File and Code Combination
Refer to the instructions in [API Documentation](api-docs.pdf), we can use the following code to call the interface:

```javascript
// Call example
fetch('/api/data', {
    method: 'POST',
    headers: {
        'Content-Type': 'application/json'
    },
    body: JSON.stringify({
        id: 123,
        name: 'test'
    })
});
```

## Attachment Management

### Supported Formats
| Type | Format | Description |
|------|--------|-------------|
| Images | PNG, JPG, GIF, SVG | Supports common image formats |
| Documents | PDF, DOC, DOCX, TXT | Supports office document formats |
| Data | XLS, XLSX, CSV | Supports spreadsheet data formats |
| Compression | ZIP, RAR | Supports compressed file formats |
| Code | MD, JSON, XML | Supports code file formats |

### Attachment Processing Features
- ✅ Automatic file type detection
- ✅ Thumbnail generation
- ✅ Embedding into NSX files
- ✅ Maintaining file integrity
- ✅ Support for relative paths

## Summary

Attachment functionality supports:
- Embedding of multiple file formats
- Automatic processing and display of images
- Creation and management of file links
- Display of mixed content 