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
	"github.com/nwaples/rardecode"
)

func main() {
	// Define the output format flag
	format := flag.String("format", "pdf", "Output format: pdf, mobi, epub")
	flag.Parse()

	if len(flag.Args()) < 1 {
		log.Fatal("Please provide at least one CBZ or CBR file.")
	}

	for _, file := range flag.Args() {
		switch filepath.Ext(file) {
		case ".cbz":
			convertCBZ(file, *format)
		case ".cbr":
			convertCBR(file, *format)
		default:
			log.Fatalf("Unsupported file format: %s", filepath.Ext(file))
		}
	}
}

func convertCBZ(cbzFile, format string) {
	switch format {
	case "pdf":
		convertCBZToPDF(cbzFile)
	case "epub":
		convertCBZToEPUB(cbzFile)
	case "mobi":
		convertCBZToEPUB(cbzFile) // First convert to EPUB
		convertEPUBToMOBI(cbzFile)
	default:
		log.Fatalf("Unsupported format: %s", format)
	}
}

func convertCBR(cbrFile, format string) {
	switch format {
	case "pdf":
		convertCBRToPDF(cbrFile)
	case "epub":
		convertCBRToEPUB(cbrFile)
	case "mobi":
		convertCBRToEPUB(cbrFile) // First convert to EPUB
		convertEPUBToMOBI(cbrFile)
	default:
		log.Fatalf("Unsupported format: %s", format)
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
		fileReader, err := file.Open()
		if err != nil {
			log.Printf("Failed to open file %s: %v", file.Name, err)
			continue
		}
		defer fileReader.Close()

		if err := addImageToPDF(pdf, file.Name, fileReader); err != nil {
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
		fileReader, err := file.Open()
		if err != nil {
			log.Printf("Failed to open file %s: %v", file.Name, err)
			continue
		}
		defer fileReader.Close()

		tmpFile, err := addImageToEPUB(e, file.Name, fileReader)
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

func convertCBRToPDF(cbrFile string) {
	// Open the CBR file as a RAR archive
	r, err := os.Open(cbrFile)
	if err != nil {
		log.Fatalf("Failed to open CBR file %s: %v", cbrFile, err)
	}
	defer r.Close()

	rardecoder, err := rardecode.NewReader(r, "")
	if err != nil {
		log.Fatalf("Failed to create RAR reader for %s: %v", cbrFile, err)
	}

	// Create a new PDF document
	pdf := gofpdf.New("P", "mm", "A4", "")

	// Iterate through the files in the archive
	for {
		hdr, err := rardecoder.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Failed to read file from CBR %s: %v", cbrFile, err)
		}

		if err := addImageToPDF(pdf, hdr.Name, rardecoder); err != nil {
			log.Printf("Failed to add image from %s: %v", hdr.Name, err)
		}
	}

	// Save the PDF with the same name as the CBR file
	pdfFileName := cbrFile[:len(cbrFile)-len(filepath.Ext(cbrFile))] + ".pdf"
	if err := pdf.OutputFileAndClose(pdfFileName); err != nil {
		log.Fatalf("Failed to save PDF file %s: %v", pdfFileName, err)
	}

	fmt.Printf("Converted %s to %s\n", cbrFile, pdfFileName)
}

func convertCBRToEPUB(cbrFile string) {
	// Open the CBR file as a RAR archive
	r, err := os.Open(cbrFile)
	if err != nil {
		log.Fatalf("Failed to open CBR file %s: %v", cbrFile, err)
	}
	defer r.Close()

	rardecoder, err := rardecode.NewReader(r, "")
	if err != nil {
		log.Fatalf("Failed to create RAR reader for %s: %v", cbrFile, err)
	}

	// Create a new EPUB document
	e := epub.NewEpub(cbrFile)

	// Store temporary file paths for cleanup
	tmpFiles := []string{}

	// Iterate through the files in the archive
	for {
		hdr, err := rardecoder.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Failed to read file from CBR %s: %v", cbrFile, err)
		}

		tmpFile, err := addImageToEPUB(e, hdr.Name, rardecoder)
		if err != nil {
			log.Printf("Failed to add image from %s: %v", hdr.Name, err)
		} else {
			tmpFiles = append(tmpFiles, tmpFile)
		}
	}

	// Save the EPUB with the same name as the CBR file
	epubFileName := cbrFile[:len(cbrFile)-len(filepath.Ext(cbrFile))] + ".epub"
	if err := e.Write(epubFileName); err != nil {
		log.Fatalf("Failed to save EPUB file %s: %v", epubFileName, err)
	}

	fmt.Printf("Converted %s to %s\n", cbrFile, epubFileName)

	// Clean up temporary files
	for _, tmpFile := range tmpFiles {
		os.Remove(tmpFile)
	}
}

func addImageToPDF(pdf *gofpdf.Fpdf, fileName string, reader io.Reader) error {
	// Create a new page in the PDF
	pdf.AddPage()

	// Register the image
	imgOptions := gofpdf.ImageOptions{ImageType: "JPEG", ReadDpi: true}
	pdf.RegisterImageOptionsReader(fileName, imgOptions, reader)

	// Add the image to the page
	pdf.ImageOptions(fileName, 10, 10, 190, 0, false, imgOptions, 0, "")

	return nil
}

func addImageToEPUB(e *epub.Epub, fileName string, reader io.Reader) (string, error) {
	// Create a temporary file to store the image
	tmpFile, err := ioutil.TempFile("", "image-*.jpg")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary file: %w", err)
	}

	// Copy the image to the temporary file
	if _, err := io.Copy(tmpFile, reader); err != nil {
		tmpFile.Close()
		return "", fmt.Errorf("failed to copy image to temporary file: %w", err)
	}
	tmpFile.Close()

	// Add the image to the EPUB
	imgPath, err := e.AddImage(tmpFile.Name(), fileName)
	if err != nil {
		return "", fmt.Errorf("failed to add image %s to EPUB: %w", fileName, err)
	}

	// Add a section with the image
	_, err = e.AddSection(fmt.Sprintf(`<img src="%s"/>`, imgPath), fileName, "", "")
	if err != nil {
		return "", fmt.Errorf("failed to add section for image %s: %w", fileName, err)
	}

	return tmpFile.Name(), nil
}

func convertEPUBToMOBI(file string) {
	// Convert EPUB to MOBI using an external tool like Calibre's ebook-convert
	epubFileName := file[:len(file)-len(filepath.Ext(file))] + ".epub"
	mobiFileName := file[:len(file)-len(filepath.Ext(file))] + ".mobi"

	cmd := exec.Command("ebook-convert", epubFileName, mobiFileName)
	if err := cmd.Run(); err != nil {
		log.Fatalf("Failed to convert EPUB to MOBI: %v", err)
	}

	fmt.Printf("Converted %s to %s\n", epubFileName, mobiFileName)
}
