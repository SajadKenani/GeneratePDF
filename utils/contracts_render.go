package utils

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"

	"github.com/go-pdf/fpdf"
)

type TextStamp struct {
	Text     string
	X, Y     float64 // mm from top-left
	FontSize float64
	Page     int // 1-indexed
}

func StampPDF(templatePDFPath string, stamps []TextStamp) ([]byte, error) {
	overlayBytes, err := buildOverlayPDF(templatePDFPath, stamps)
	if err != nil {
		return nil, fmt.Errorf("build overlay: %w", err)
	}

	tmpOverlay, err := os.CreateTemp("", "overlay-*.pdf")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmpOverlay.Name())
	if _, err := tmpOverlay.Write(overlayBytes); err != nil {
		return nil, err
	}
	tmpOverlay.Close()

	tmpOut, err := os.CreateTemp("", "stamped-*.pdf")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmpOut.Name())
	tmpOut.Close()

	cmd := exec.Command("pdftk",
		templatePDFPath,
		"multistamp", tmpOverlay.Name(),
		"output", tmpOut.Name(),
	)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("pdftk failed: %w — %s", err, stderr.String())
	}

	return os.ReadFile(tmpOut.Name())
}
func buildOverlayPDF(templatePDFPath string, stamps []TextStamp) ([]byte, error) {
	pageCount, width, height, err := getPDFInfo(templatePDFPath)
	if err != nil {
		pageCount, width, height = 1, 210.0, 297.0
	}

	pdf := fpdf.NewCustom(&fpdf.InitType{
		UnitStr: "mm",
		Size:    fpdf.SizeType{Wd: width, Ht: height},
	})
	pdf.SetMargins(0, 0, 0)
	pdf.SetAutoPageBreak(false, 0)

	byPage := make(map[int][]TextStamp)
	for _, s := range stamps {
		byPage[s.Page] = append(byPage[s.Page], s)
	}

	for p := 1; p <= pageCount; p++ {
		pdf.AddPage()
		for _, s := range byPage[p] {
			pdf.SetFont("Helvetica", "", s.FontSize)
			pdf.SetXY(s.X, s.Y)
			pdf.Cell(0, 0, s.Text)
		}
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func getPDFInfo(pdfPath string) (pages int, widthMM, heightMM float64, err error) {
	out, err := exec.Command("pdfinfo", pdfPath).Output()
	if err != nil {
		return 0, 0, 0, err
	}

	lines := bytes.Split(out, []byte("\n"))
	for _, line := range lines {
		var n int
		if bytes.HasPrefix(line, []byte("Pages:")) {
			fmt.Sscanf(string(line), "Pages: %d", &n)
			pages = n
		}
		var wPt, hPt float64
		if bytes.HasPrefix(line, []byte("Page size:")) {
			fmt.Sscanf(string(line), "Page size: %f x %f pts", &wPt, &hPt)
			widthMM = wPt / 2.8346
			heightMM = hPt / 2.8346
		}
	}

	if pages == 0 {
		return 0, 0, 0, fmt.Errorf("could not parse pdfinfo output")
	}
	return pages, widthMM, heightMM, nil
}