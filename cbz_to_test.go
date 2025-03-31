package main

import (
	"archive/zip"
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

// createDummyCBZ creates a dummy CBZ file with sample images for testing
func createDummyCBZ(t *testing.T, cbzFileName string) {
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	// Create a sample image
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, color.RGBA{uint8(x), uint8(y), 0, 255})
		}
	}

	// Add the image to the ZIP archive
	imgFile, err := zipWriter.Create("sample.jpg")
	if err != nil {
		t.Fatalf("Failed to create image in ZIP: %v", err)
	}

	if err := jpeg.Encode(imgFile, img, nil); err != nil {
		t.Fatalf("Failed to encode image: %v", err)
	}

	// Close the ZIP writer
	if err := zipWriter.Close(); err != nil {
		t.Fatalf("Failed to close ZIP writer: %v", err)
	}

	// Write the buffer to a file
	if err := ioutil.WriteFile(cbzFileName, buf.Bytes(), 0644); err != nil {
		t.Fatalf("Failed to write CBZ file: %v", err)
	}
}

// TestConversion tests the conversion of a dummy CBZ file to various formats
func TestConversion(t *testing.T) {
	cbzFileName := "test.cbz"
	createDummyCBZ(t, cbzFileName)
	defer os.Remove(cbzFileName)

	// Test PDF conversion
	pdfFileName := cbzFileName[:len(cbzFileName)-len(filepath.Ext(cbzFileName))] + ".pdf"
	convertCBZToPDF(cbzFileName)
	if _, err := os.Stat(pdfFileName); os.IsNotExist(err) {
		t.Errorf("PDF file was not created: %s", pdfFileName)
	} else {
		os.Remove(pdfFileName)
	}

	// Test EPUB conversion
	epubFileName := cbzFileName[:len(cbzFileName)-len(filepath.Ext(cbzFileName))] + ".epub"
	convertCBZToEPUB(cbzFileName)
	if _, err := os.Stat(epubFileName); os.IsNotExist(err) {
		t.Errorf("EPUB file was not created: %s", epubFileName)
	} else {
		os.Remove(epubFileName)
	}

}
