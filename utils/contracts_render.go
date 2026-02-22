package utils

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"unicode"

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

// Arabic character forms: [isolated, final, initial, medial]
var arabicForms = map[rune][4]rune{
	'ا': {0xFE8D, 0xFE8E, 0xFE8D, 0xFE8E},
	'أ': {0xFE83, 0xFE84, 0xFE83, 0xFE84},
	'إ': {0xFE87, 0xFE88, 0xFE87, 0xFE88},
	'آ': {0xFE81, 0xFE82, 0xFE81, 0xFE82},
	'ؤ': {0xFE85, 0xFE86, 0xFE85, 0xFE86},
	'ء': {0xFE80, 0xFE80, 0xFE80, 0xFE80},
	'ب': {0xFE8F, 0xFE90, 0xFE91, 0xFE92},
	'ت': {0xFE95, 0xFE96, 0xFE97, 0xFE98},
	'ث': {0xFE99, 0xFE9A, 0xFE9B, 0xFE9C},
	'ج': {0xFE9D, 0xFE9E, 0xFE9F, 0xFEA0},
	'ح': {0xFEA1, 0xFEA2, 0xFEA3, 0xFEA4},
	'خ': {0xFEA5, 0xFEA6, 0xFEA7, 0xFEA8},
	'د': {0xFEA9, 0xFEAA, 0xFEA9, 0xFEAA},
	'ذ': {0xFEAB, 0xFEAC, 0xFEAB, 0xFEAC},
	'ر': {0xFEAD, 0xFEAE, 0xFEAD, 0xFEAE},
	'ز': {0xFEAF, 0xFEB0, 0xFEAF, 0xFEB0},
	'س': {0xFEB1, 0xFEB2, 0xFEB3, 0xFEB4},
	'ش': {0xFEB5, 0xFEB6, 0xFEB7, 0xFEB8},
	'ص': {0xFEB9, 0xFEBA, 0xFEBB, 0xFEBC},
	'ض': {0xFEBD, 0xFEBE, 0xFEBF, 0xFEC0},
	'ط': {0xFEC1, 0xFEC2, 0xFEC3, 0xFEC4},
	'ظ': {0xFEC5, 0xFEC6, 0xFEC7, 0xFEC8},
	'ع': {0xFEC9, 0xFECA, 0xFECB, 0xFECC},
	'غ': {0xFECD, 0xFECE, 0xFECF, 0xFED0},
	'ف': {0xFED1, 0xFED2, 0xFED3, 0xFED4},
	'ق': {0xFED5, 0xFED6, 0xFED7, 0xFED8},
	'ك': {0xFED9, 0xFEDA, 0xFEDB, 0xFEDC},
	'ل': {0xFEDD, 0xFEDE, 0xFEDF, 0xFEE0},
	'م': {0xFEE1, 0xFEE2, 0xFEE3, 0xFEE4},
	'ن': {0xFEE5, 0xFEE6, 0xFEE7, 0xFEE8},
	'ه': {0xFEE9, 0xFEEA, 0xFEEB, 0xFEEC},
	'و': {0xFEED, 0xFEEE, 0xFEED, 0xFEEE},
	'ي': {0xFEF1, 0xFEF2, 0xFEF3, 0xFEF4},
	'ى': {0xFEEF, 0xFEF0, 0xFBE8, 0xFBE9},
	'ة': {0xFE93, 0xFE94, 0xFE93, 0xFE94},
	'ئ': {0xFE89, 0xFE8A, 0xFE8B, 0xFE8C},
}

// nonJoining are Arabic letters that do NOT connect to the following character.
var nonJoining = map[rune]bool{
	'ا': true, 'أ': true, 'إ': true, 'آ': true,
	'ؤ': true, 'ء': true,
	'د': true, 'ذ': true,
	'ر': true, 'ز': true,
	'و': true,
}

func isArabicLetter(r rune) bool {
	_, ok := arabicForms[r]
	return ok
}

// canConnectToNext returns true if r can connect to the letter after it.
func canConnectToNext(r rune) bool {
	return isArabicLetter(r) && !nonJoining[r]
}

// shapeWord shapes a single Arabic word (no spaces).
func shapeWord(runes []rune) []rune {
	n := len(runes)
	result := make([]rune, 0, n)

	// Handle Lam + Alef ligature first
	for i := 0; i < n; i++ {
		if runes[i] == 'ل' && i+1 < n && (runes[i+1] == 'ا' || runes[i+1] == 'أ' || runes[i+1] == 'إ' || runes[i+1] == 'آ') {
			// Determine which lam-alef ligature form
			prevConnects := i > 0 && canConnectToNext(runes[i-1])
			if prevConnects {
				result = append(result, 0xFEFC) // final lam-alef
			} else {
				result = append(result, 0xFEFB) // isolated lam-alef
			}
			i++ // skip the alef
			continue
		}

		r := runes[i]
		forms, ok := arabicForms[r]
		if !ok {
			// Non-Arabic character inside word (e.g. tatweel, diacritics)
			result = append(result, r)
			continue
		}

		prevConnects := i > 0 && canConnectToNext(runes[i-1])
		nextConnects := i < n-1 && isArabicLetter(runes[i+1])

		var form int
		switch {
		case prevConnects && nextConnects:
			form = 3 // medial
		case prevConnects:
			form = 1 // final
		case nextConnects:
			form = 2 // initial
		default:
			form = 0 // isolated
		}
		result = append(result, forms[form])
	}
	return result
}

// reverseRunes reverses a rune slice in place.
func reverseRunes(r []rune) {
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
}

// processArabicText shapes and applies BiDi reversal to Arabic text.
// Handles mixed Arabic+Latin+numbers correctly by treating each token separately.
func processArabicText(text string) string {
	if !containsArabic(text) {
		return text
	}

	// Tokenize into Arabic words and non-Arabic segments
	type token struct {
		text    string
		isArabic bool
	}

	var tokens []token
	var current strings.Builder
	currentIsArabic := false

	flush := func() {
		if current.Len() > 0 {
			tokens = append(tokens, token{current.String(), currentIsArabic})
			current.Reset()
		}
	}

	for _, r := range text {
		arabic := isArabicLetter(r) || (r >= 0x0600 && r <= 0x06FF)
		if arabic != currentIsArabic {
			flush()
			currentIsArabic = arabic
		}
		current.WriteRune(r)
	}
	flush()

	// Shape each Arabic token, then reverse ALL tokens for RTL rendering
	shaped := make([]string, len(tokens))
	for i, t := range tokens {
		if t.isArabic {
			// Shape word by word within the token
			words := strings.Fields(t.text)
			shapedWords := make([]string, len(words))
			for j, w := range words {
				runes := []rune(w)
				shapedRunes := shapeWord(runes)
				reverseRunes(shapedRunes)
				shapedWords[j] = string(shapedRunes)
			}
			// Reverse word order within the Arabic token (RTL)
			for a, b := 0, len(shapedWords)-1; a < b; a, b = a+1, b-1 {
				shapedWords[a], shapedWords[b] = shapedWords[b], shapedWords[a]
			}
			shaped[i] = strings.Join(shapedWords, " ")
		} else {
			shaped[i] = t.text
		}
	}

	// Reverse token order for overall RTL layout
	for a, b := 0, len(shaped)-1; a < b; a, b = a+1, b-1 {
		shaped[a], shaped[b] = shaped[b], shaped[a]
	}

	return strings.Join(shaped, "")
}

func containsArabic(s string) bool {
	for _, r := range s {
		if (r >= 0x0600 && r <= 0x06FF) || (r >= 0xFE70 && r <= 0xFEFF) {
			return true
		}
	}
	return false
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

	// Load font directly from bytes — no temp file, no path issues
	fontData, err := os.ReadFile("/app/fonts/NotoSansArabic-Regular.ttf")
	if err != nil {
		return nil, fmt.Errorf("failed to read Arabic font: %w", err)
	}

	pdf.AddUTF8FontFromBytes("NotoArabic", "", fontData)
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

			if containsArabic(s.Text) {
				pdf.SetFont("NotoArabic", "", s.FontSize)
				displayText = processArabicText(s.Text)

				textWidth := pdf.GetStringWidth(displayText)
				pdf.SetXY(s.X-textWidth, s.Y)
			} else {
				pdf.SetFont("Helvetica", "", s.FontSize)
				pdf.SetXY(s.X, s.Y)
			}

			if pdf.Error() != nil {
				return nil, fmt.Errorf("font error: %w", pdf.Error())
			}

			pdf.Cell(0, s.FontSize*0.3528, displayText)
		}
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
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

// unused import guard
var _ = unicode.IsLetter