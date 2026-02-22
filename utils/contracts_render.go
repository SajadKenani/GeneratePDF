package utils

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"

	"github.com/go-pdf/fpdf"
)

type TextStamp struct {
	Text       string
	X, Y       float64
	FontSize   float64
	Page       int
	FontFamily string
	IsRTL      bool
}

// Arabic character forms mapping
var arabicShaping = map[rune][4]rune{
	// Letter: [isolated, final, initial, medial]
	'ا': {0xFE8D, 0xFE8E, 0xFE8D, 0xFE8E}, // Alef
	'ب': {0xFE8F, 0xFE90, 0xFE91, 0xFE92}, // Beh
	'ت': {0xFE95, 0xFE96, 0xFE97, 0xFE98}, // Teh
	'ث': {0xFE99, 0xFE9A, 0xFE9B, 0xFE9C}, // Theh
	'ج': {0xFE9D, 0xFE9E, 0xFE9F, 0xFEA0}, // Jeem
	'ح': {0xFEA1, 0xFEA2, 0xFEA3, 0xFEA4}, // Hah
	'خ': {0xFEA5, 0xFEA6, 0xFEA7, 0xFEA8}, // Khah
	'د': {0xFEA9, 0xFEAA, 0xFEA9, 0xFEAA}, // Dal
	'ذ': {0xFEAB, 0xFEAC, 0xFEAB, 0xFEAC}, // Thal
	'ر': {0xFEAD, 0xFEAE, 0xFEAD, 0xFEAE}, // Reh
	'ز': {0xFEAF, 0xFEB0, 0xFEAF, 0xFEB0}, // Zain
	'س': {0xFEB1, 0xFEB2, 0xFEB3, 0xFEB4}, // Seen
	'ش': {0xFEB5, 0xFEB6, 0xFEB7, 0xFEB8}, // Sheen
	'ص': {0xFEB9, 0xFEBA, 0xFEBB, 0xFEBC}, // Sad
	'ض': {0xFEBD, 0xFEBE, 0xFEBF, 0xFEC0}, // Dad
	'ط': {0xFEC1, 0xFEC2, 0xFEC3, 0xFEC4}, // Tah
	'ظ': {0xFEC5, 0xFEC6, 0xFEC7, 0xFEC8}, // Zah
	'ع': {0xFEC9, 0xFECA, 0xFECB, 0xFECC}, // Ain
	'غ': {0xFECD, 0xFECE, 0xFECF, 0xFED0}, // Ghain
	'ف': {0xFED1, 0xFED2, 0xFED3, 0xFED4}, // Feh
	'ق': {0xFED5, 0xFED6, 0xFED7, 0xFED8}, // Qaf
	'ك': {0xFED9, 0xFEDA, 0xFEDB, 0xFEDC}, // Kaf
	'ل': {0xFEDD, 0xFEDE, 0xFEDF, 0xFEE0}, // Lam
	'م': {0xFEE1, 0xFEE2, 0xFEE3, 0xFEE4}, // Meem
	'ن': {0xFEE5, 0xFEE6, 0xFEE7, 0xFEE8}, // Noon
	'ه': {0xFEE9, 0xFEEA, 0xFEEB, 0xFEEC}, // Heh
	'و': {0xFEED, 0xFEEE, 0xFEED, 0xFEEE}, // Waw
	'ي': {0xFEF1, 0xFEF2, 0xFEF3, 0xFEF4}, // Yeh
	'ى': {0xFEEF, 0xFEF0, 0xFBE8, 0xFBE9}, // Alef Maksura
	'ة': {0xFE93, 0xFE94, 0xFE93, 0xFE94}, // Teh Marbuta
	'ئ': {0xFE89, 0xFE8A, 0xFE8B, 0xFE8C}, // Yeh Hamza
}

// Check if a character connects to the next character
func isConnecting(r rune) bool {
	// Non-connecting characters
	nonConnecting := map[rune]bool{
		'ا': true, 'د': true, 'ذ': true, 'ر': true, 'ز': true, 'و': true,
	}
	return !nonConnecting[r]
}

// Shape Arabic text
func shapeArabic(text string) string {
	runes := []rune(text)
	result := make([]rune, len(runes))
	
	for i, r := range runes {
		if forms, ok := arabicShaping[r]; ok {
			var form int
			
			// Determine which form to use
			prevConnects := i > 0 && isConnecting(runes[i-1])
			nextConnects := i < len(runes)-1 && isConnecting(r)
			
			if prevConnects && nextConnects {
				form = 3 // medial
			} else if prevConnects {
				form = 1 // final
			} else if nextConnects {
				form = 2 // initial
			} else {
				form = 0 // isolated
			}
			
			result[i] = forms[form]
		} else {
			result[i] = r
		}
	}
	
	return string(result)
}
func processArabicText(text string) string {
	if !containsArabic(text) {
		return text
	}
	
	shaped := shapeArabic(text)
	runes := []rune(shaped)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	
	return string(runes)
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

	// Read and copy font to temp location
	fontData, err := os.ReadFile("/app/fonts/NotoSansArabic-Regular.ttf")
	if err != nil {
		return nil, fmt.Errorf("failed to read Arabic font: %w", err)
	}
	
	tmpFont := "NotoSansArabic-Regular.ttf"
	if err := os.WriteFile(tmpFont, fontData, 0644); err != nil {
		return nil, fmt.Errorf("failed to write temp font: %w", err)
	}
	defer os.Remove(tmpFont)
	
	pdf.AddUTF8Font("NotoArabic", "", tmpFont)
	
	if pdf.Error() != nil {
		return nil, fmt.Errorf("failed to load Arabic font: %w", pdf.Error())
	}

	byPage := make(map[int][]TextStamp)
	for _, s := range stamps {
		byPage[s.Page] = append(byPage[s.Page], s)
	}

	for p := 1; p <= pageCount; p++ {
		pdf.AddPage()
		for _, s := range byPage[p] {
			displayText := s.Text
			
			// Process Arabic text for proper display with shaping
			if containsArabic(s.Text) {
				pdf.SetFont("NotoArabic", "", s.FontSize)
				displayText = processArabicText(s.Text)
				
				// For RTL text, adjust X position to right-align
				textWidth := pdf.GetStringWidth(displayText)
				pdf.SetXY(s.X-textWidth, s.Y)
			} else {
				pdf.SetFont("Helvetica", "", s.FontSize)
				pdf.SetXY(s.X, s.Y)
			}
			
			if pdf.Error() != nil {
				return nil, fmt.Errorf("font error: %w", pdf.Error())
			}
			
			pdf.Cell(0, 0, displayText)
		}
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func containsArabic(s string) bool {
	for _, r := range s {
		if r >= 0x0600 && r <= 0x06FF {
			return true
		}
	}
	return false
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