package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/bmaupin/go-epub"
	"github.com/jung-kurt/gofpdf"
)

func main() {
	// Define the output format flag
	format := flag.String("format", "pdf", "Output format: pdf, mobi, epub")
	flag.Parse()

	if len(flag.Args()) < 1 {
		log.Fatal("Please provide at least one CBZ file.")
	}

	for _, cbzFile := range flag.Args() {
		switch *format {
		case "pdf":
			convertCBZToPDF(cbzFile)
		case "epub":
			convertCBZToEPUB(cbzFile)
		case "mobi":
			convertCBZToEPUB(cbzFile) // First convert to EPUB
			convertEPUBToMOBI(cbzFile)
		default:
			log.Fatalf("Unsupported format: %s", *format)
		}
	}
}

func convertCBZToPDF(cbzFile string) {
	// Open the CBZ file as a ZIP archive
	zipReader, err := zip.OpenReader(cbzFile)
	if err != nil {
		log.Fatalf("Failed to open CBZ file %s: %v", cbzFile, err)
	}
	defer zipReader.Close()

	// Create a new PDF document
	pdf := gofpdf.New("P", "mm", "A4", "")

	// Iterate through the files in the archive
	for _, file := range zipReader.File {
		if err := addImageToPDF(pdf, file); err != nil {
			log.Printf("Failed to add image from %s: %v", file.Name, err)
		}
	}

	// Save the PDF with the same name as the CBZ file
	pdfFileName := cbzFile[:len(cbzFile)-len(filepath.Ext(cbzFile))] + ".pdf"
	if err := pdf.OutputFileAndClose(pdfFileName); err != nil {
		log.Fatalf("Failed to save PDF file %s: %v", pdfFileName, err)
	}

	fmt.Printf("Converted %s to %s\n", cbzFile, pdfFileName)
}

func convertCBZToEPUB(cbzFile string) {
	// Open the CBZ file as a ZIP archive
	zipReader, err := zip.OpenReader(cbzFile)
	if err != nil {
		log.Fatalf("Failed to open CBZ file %s: %v", cbzFile, err)
	}
	defer zipReader.Close()

	// Create a new EPUB document
	e := epub.NewEpub(cbzFile)

	// Store temporary file paths for cleanup
	tmpFiles := []string{}

	// Iterate through the files in the archive
	for _, file := range zipReader.File {
		tmpFile, err := addImageToEPUB(e, file)
		if err != nil {
			log.Printf("Failed to add image from %s: %v", file.Name, err)
		} else {
			tmpFiles = append(tmpFiles, tmpFile)
		}
	}

	// Save the EPUB with the same name as the CBZ file
	epubFileName := cbzFile[:len(cbzFile)-len(filepath.Ext(cbzFile))] + ".epub"
	if err := e.Write(epubFileName); err != nil {
		log.Fatalf("Failed to save EPUB file %s: %v", epubFileName, err)
	}

	fmt.Printf("Converted %s to %s\n", cbzFile, epubFileName)

	// Clean up temporary files
	for _, tmpFile := range tmpFiles {
		os.Remove(tmpFile)
	}
}

func addImageToPDF(pdf *gofpdf.Fpdf, file *zip.File) error {
	// Open the file inside the ZIP archive
	fileReader, err := file.Open()
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", file.Name, err)
	}
	defer fileReader.Close()

	// Create a new page in the PDF
	pdf.AddPage()

	// Register the image
	imgOptions := gofpdf.ImageOptions{ImageType: "JPEG", ReadDpi: true}
	pdf.RegisterImageOptionsReader(file.Name, imgOptions, fileReader)

	// Add the image to the page
	pdf.ImageOptions(file.Name, 10, 10, 190, 0, false, imgOptions, 0, "")

	return nil
}

func addImageToEPUB(e *epub.Epub, file *zip.File) (string, error) {
	// Open the file inside the ZIP archive
	fileReader, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open file %s: %w", file.Name, err)
	}
	defer fileReader.Close()

	// Create a temporary file to store the image
	tmpFile, err := ioutil.TempFile("", "image-*.jpg")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary file: %w", err)
	}

	// Copy the image to the temporary file
	if _, err := io.Copy(tmpFile, fileReader); err != nil {
		tmpFile.Close()
		return "", fmt.Errorf("failed to copy image to temporary file: %w", err)
	}
	tmpFile.Close()

	// Add the image to the EPUB
	imgPath, err := e.AddImage(tmpFile.Name(), file.Name)
	if err != nil {
		return "", fmt.Errorf("failed to add image %s to EPUB: %w", file.Name, err)
	}

	// Add a section with the image
	_, err = e.AddSection(fmt.Sprintf(`<img src="%s"/>`, imgPath), file.Name, "", "")
	if err != nil {
		return "", fmt.Errorf("failed to add section for image %s: %w", file.Name, err)
	}

	return tmpFile.Name(), nil
}

func convertEPUBToMOBI(cbzFile string) {
	// Convert EPUB to MOBI using an external tool like Calibre's ebook-convert
	epubFileName := cbzFile[:len(cbzFile)-len(filepath.Ext(cbzFile))] + ".epub"
	mobiFileName := cbzFile[:len(cbzFile)-len(filepath.Ext(cbzFile))] + ".mobi"

	cmd := exec.Command("ebook-convert", epubFileName, mobiFileName)
	if err := cmd.Run(); err != nil {
		log.Fatalf("Failed to convert EPUB to MOBI: %v", err)
	}

	fmt.Printf("Converted %s to %s\n", epubFileName, mobiFileName)
}
