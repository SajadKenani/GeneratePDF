package utils

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/pdfcpu/pdfcpu/pkg/api"
)

type ContractData struct {
	NumID          int       `json:"numId"`
	CreatedAt      time.Time `json:"createdAt"`
	ItemAmount     float64   `json:"itemAmount"`
	PaymentMethod  string    `json:"paymentMethod"`
	Realm          string    `json:"realm"`
	Currency       string    `json:"currency"`
	CurrencyAmount float64   `json:"currencyAmount"`
	UserName       string    `json:"userName"`
	UserPhone      string    `json:"userPhone"`
	UserEmail      string    `json:"userEmail"`
	Item           string    `json:"item"`
	Insurance      bool      `json:"insurance"`
}

type AttachmentProps struct {
	Filename    string `json:"filename"`
	Content     string `json:"content"`
	Encoding    string `json:"encoding"`
	ContentType string `json:"contentType"`
}

type pdfOptions struct {
	paperWidthIn  float64
	paperHeightIn float64
	marginIn      float64
	printBG       bool
}

func debugEnabled() bool {
	return strings.TrimSpace(os.Getenv("CHROMEDP_DEBUG")) != ""
}

func debugLogf(enabled bool, format string, args ...any) {
	if enabled {
		log.Printf("[pdf] "+format, args...)
	}
}

var defaultPDFOptions = pdfOptions{
	// A4 in inches
	paperWidthIn:  8.27,
	paperHeightIn: 11.69,
	marginIn:      0,
	printBG:       true,
}

// GetDepositAttachments mirrors the JS logic, but always returns attachments.
// - If realm == "romania" and item contains "cattle" => single PDF attachment
// - Otherwise => 6 attachments: rental, care, sales, auth, withdraw, receipt
func GetDepositAttachments(ctx context.Context, contractData ContractData) ([]AttachmentProps, error) {
	debug := debugEnabled()
	itemLower := strings.ToLower(strings.TrimSpace(contractData.Item))

	// If item mentions cattle or rooster we generate a single robust PDF (merged header/content/footer)
	if strings.Contains(itemLower, "cattle") || strings.Contains(itemLower, "rooster") {
		return generatePdfRobust(ctx, contractData)
	}

	return generateSixAttachments(ctx, contractData, debug)
}

// generatePdfRobust creates a browser, opens a single tab and renders a merged contract PDF
// for items like cattle/rooster.
func generatePdfRobust(ctx context.Context, contractData ContractData) ([]AttachmentProps, error) {
	debug := debugEnabled()
	browserCtx, cancelBrowser, err := newBrowserContext(ctx)
	if err != nil {
		return nil, err
	}
	defer cancelBrowser()

	// create a tab (child context) from the browser context
	tabCtx, cancelTab := chromedp.NewContext(browserCtx)
	defer cancelTab()

	// Determine contractType key
	itemLower := strings.ToLower(strings.TrimSpace(contractData.Item))
	contractType := "cattle"
	if strings.Contains(itemLower, "rooster") {
		contractType = "rooster"
	}

	if debug {
		debugLogf(debug, "generatePdfRobust contractType=%s", contractType)
	}

	pdf, err := generateContractPDF(tabCtx, contractType, contractData)
	if err != nil {
		return nil, err
	}

	return []AttachmentProps{pdfToAttachment(contractData.Item+"-contract", contractData.NumID, pdf)}, nil
}

func pdfToAttachment(prefix string, numID int, pdf []byte) AttachmentProps {
	return AttachmentProps{
		Filename:    fmt.Sprintf("%s-%d.pdf", prefix, numID),
		Content:     base64.StdEncoding.EncodeToString(pdf),
		Encoding:    "base64",
		ContentType: "application/pdf",
	}
}

func generateSixAttachments(ctx context.Context, contractData ContractData, debug bool) ([]AttachmentProps, error) {
	contractTypes := []string{"sales", "rental", "care", "auth", "withdraw", "receipt"}

	browser, cancelBrowser, err := newBrowserContext(ctx)
	if err != nil {
		return nil, err
	}
	defer cancelBrowser()
	if debug {
		debugLogf(debug, "browser context created: err=%v cause=%v", browser.Err(), context.Cause(browser))
		_ = context.AfterFunc(browser, func() {
			debugLogf(debug, "browser context canceled: err=%v cause=%v", browser.Err(), context.Cause(browser))
		})
	}

	var (
		mu          sync.Mutex
		attachments []AttachmentProps
	)

	renderBatch := func(types []string) error {
		for _, ct := range types {
			if debug {
				debugLogf(debug, "start contract type=%s", ct)
			}
			// Create a lightweight tab context from the shared browser context to reuse the browser process.
			tabCtx, cancelTab := chromedp.NewContext(browser)
			if cancelTab == nil {
				return fmt.Errorf("failed to create tab context")
			}
			tabCtx, cancelTimeout := context.WithTimeout(tabCtx, 75*time.Second)
			if debug {
				debugLogf(debug, "tab context after timeout wrap: err=%v cause=%v", tabCtx.Err(), context.Cause(tabCtx))
				_ = context.AfterFunc(tabCtx, func() {
					debugLogf(debug, "tab context canceled: type=%s err=%v cause=%v", ct, tabCtx.Err(), context.Cause(tabCtx))
				})
			}
			if debug {
				debugLogf(debug, "tab context before generateContractPDF: err=%v cause=%v", tabCtx.Err(), context.Cause(tabCtx))
			}
			pdf, err := generateContractPDF(tabCtx, ct, contractData)
			if err != nil {
				cancelTimeout()
				cancelTab()
				return err
			}
			cancelTimeout()
			cancelTab()

			mu.Lock()
			attachments = append(attachments, pdfToAttachment(ct+"-contract", contractData.NumID, pdf))
			mu.Unlock()
		}
		return nil
	}

	if err := renderBatch(contractTypes[:3]); err != nil {
		return nil, err
	}
	if err := renderBatch(contractTypes[3:]); err != nil {
		return nil, err
	}

	ordered := make([]AttachmentProps, 0, len(contractTypes))
	for _, ct := range contractTypes {
		name := fmt.Sprintf("%s-contract-%d.pdf", ct, contractData.NumID)
		for _, a := range attachments {
			if a.Filename == name {
				ordered = append(ordered, a)
				break
			}
		}
	}

	return ordered, nil
}

func newBrowserContext(parent context.Context) (context.Context, context.CancelFunc, error) {
	debug := strings.TrimSpace(os.Getenv("CHROMEDP_DEBUG")) != ""
	execPath := chromeExecPath()
	if debug {
		fmt.Printf("[chromedp] using execPath=%q\n", execPath)
	}
	logFile := chromeLogFile()
	profileDir, _ := os.MkdirTemp("", "chromedp-profile-")
	baseOpts := []chromedp.ExecAllocatorOption{
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-setuid-sandbox", true),
		chromedp.Flag("single-process", true),
		chromedp.Flag("no-zygote", true),
		chromedp.Flag("disable-software-rasterizer", true),
		chromedp.Flag("enable-logging", true),
		chromedp.Flag("v", "1"),
		chromedp.Flag("log-file", logFile),
		chromedp.Flag("crash-dumps-dir", "/tmp/chrome-crash"),
		chromedp.UserDataDir(profileDir),
	}
	opts := append([]chromedp.ExecAllocatorOption{}, baseOpts...)
	if execPath != "" {
		opts = append(opts, chromedp.ExecPath(execPath))
	}

	allocCtx, cancelAlloc := chromedp.NewExecAllocator(parent, opts...)
	if debug {
		_ = context.AfterFunc(allocCtx, func() {
			log.Printf("[pdf] alloc context canceled: err=%v cause=%v parentErr=%v parentCause=%v", allocCtx.Err(), context.Cause(allocCtx), parent.Err(), context.Cause(parent))
		})
	}

	var browserCtx context.Context
	var cancelCtx context.CancelFunc
	if debug {
		logger := func(format string, args ...any) { log.Printf("[chromedp] "+format, args...) }
		browserCtx, cancelCtx = chromedp.NewContext(allocCtx, chromedp.WithLogf(logger), chromedp.WithDebugf(logger), chromedp.WithErrorf(logger))
	} else {
		browserCtx, cancelCtx = chromedp.NewContext(allocCtx)
	}
	if debug {
		_ = context.AfterFunc(browserCtx, func() {
			log.Printf("[pdf] browserCtx AfterFunc: err=%v cause=%v", browserCtx.Err(), context.Cause(browserCtx))
		})
	}

	cancel := func() {
		cancelCtx()
		cancelAlloc()
		if profileDir != "" {
			_ = os.RemoveAll(profileDir)
		}
	}
	return browserCtx, cancel, nil
}

func generatePDF(ctx context.Context, html string, opt pdfOptions) ([]byte, error) {
	debug := debugEnabled()
	if strings.TrimSpace(html) == "" {
		return nil, errors.New("empty html")
	}
	if debug {
		debugLogf(debug, "generatePDF context state: err=%v cause=%v", ctx.Err(), context.Cause(ctx))
	}

	if err := chromedp.Run(ctx,
		page.Enable(),
		dom.Enable(),
		emulation.SetDeviceMetricsOverride(1200, 1600, 1.0, false),
		emulation.SetEmulatedMedia().WithMedia("screen"),
	); err != nil {
		if debug {
			debugLogf(debug, "chromedp init failed: err=%v cause=%v", err, context.Cause(ctx))
		}
		return nil, fmt.Errorf("chromedp init failed: %w", err)
	}

	if err := chromedp.Run(ctx,
		chromedp.Navigate("about:blank"),
		chromedp.WaitReady("html", chromedp.ByQuery),
	); err != nil {
		return nil, fmt.Errorf("chromedp navigate failed: %w", err)
	}

	if err := chromedp.Run(ctx,
		setDocumentHTML(html),
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.Sleep(600*time.Millisecond),
	); err != nil {
		return nil, fmt.Errorf("chromedp set html failed: %w", err)
	}

	var pdf []byte
	if err := chromedp.Run(ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			buf, _, err := page.PrintToPDF().
				WithPrintBackground(opt.printBG).
				WithPaperWidth(opt.paperWidthIn).
				WithPaperHeight(opt.paperHeightIn).
				WithMarginTop(opt.marginIn).
				WithMarginBottom(opt.marginIn).
				WithMarginLeft(opt.marginIn).
				WithMarginRight(opt.marginIn).
				Do(ctx)
			if err != nil {
				return err
			}
			pdf = buf
			return nil
		}),
	); err != nil {
		return nil, fmt.Errorf("chromedp printToPDF failed: %w", err)
	}

	return pdf, nil
}

func setDocumentHTML(html string) chromedp.Action {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		var lastErr error
		for i := 0; i < 10; i++ {
			tree, err := page.GetFrameTree().Do(ctx)
			if err == nil && tree != nil && tree.Frame != nil {
				return page.SetDocumentContent(tree.Frame.ID, html).Do(ctx)
			}
			if err != nil {
				lastErr = err
			} else {
				lastErr = errors.New("no frame tree")
			}
			select {
			case <-time.After(50 * time.Millisecond):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		return lastErr
	})
}

func isThreePageContractType(contractType string) bool {
	switch contractType {
	case "sales", "care", "rental", "auth", "withdraw", "cattle", "rooster":
		return true
	default:
		return false
	}
}

// mergePDFBuffers merges multiple PDF byte slices into one PDF using pdfcpu.
func mergePDFBuffers(parts [][]byte) ([]byte, error) {
	if len(parts) == 0 {
		return nil, errors.New("no pdf parts to merge")
	}
	// Write parts to temporary files because the pdfcpu Merge API expects file paths.
	tmpFiles := make([]string, 0, len(parts))
	for i, p := range parts {
		f, err := os.CreateTemp("", fmt.Sprintf("pdfpart-%d-*.pdf", i))
		if err != nil {
			// cleanup any created files
			for _, tf := range tmpFiles {
				_ = os.Remove(tf)
			}
			return nil, fmt.Errorf("failed to create temp file: %w", err)
		}
		if _, err := f.Write(p); err != nil {
			_ = f.Close()
			_ = os.Remove(f.Name())
			for _, tf := range tmpFiles {
				_ = os.Remove(tf)
			}
			return nil, fmt.Errorf("failed to write temp pdf part: %w", err)
		}
		_ = f.Close()
		tmpFiles = append(tmpFiles, f.Name())
	}

	var out bytes.Buffer
	// api.Merge signature varies; call the file-based variant: api.Merge(outFile string, inFiles []string, w io.Writer, conf *model.Configuration, validate bool)
	if err := api.Merge("", tmpFiles, &out, nil, false); err != nil {
		for _, tf := range tmpFiles {
			_ = os.Remove(tf)
		}
		return nil, fmt.Errorf("pdf merge failed: %w", err)
	}

	for _, tf := range tmpFiles {
		_ = os.Remove(tf)
	}
	return out.Bytes(), nil
}

// generateContractPDF produces a PDF for a given contract type. For multi-page contracts
// (header/content/footer) it renders each template separately and merges them.
func generateContractPDF(ctx context.Context, contractType string, contractData ContractData) ([]byte, error) {
	// three-page style
	if isThreePageContractType(contractType) {
		headerHTML, err := RenderTemplateHTML("header", contractData, contractType)
		if err != nil {
			return nil, err
		}
		contentHTML, err := RenderTemplateHTML("content", contractData, contractType)
		if err != nil {
			return nil, err
		}
		footerHTML, err := RenderTemplateHTML("footer", contractData, contractType)
		if err != nil {
			return nil, err
		}

		headerPdf, err := generatePDF(ctx, headerHTML, defaultPDFOptions)
		if err != nil {
			return nil, err
		}
		contentPdf, err := generatePDF(ctx, contentHTML, defaultPDFOptions)
		if err != nil {
			return nil, err
		}
		footerPdf, err := generatePDF(ctx, footerHTML, defaultPDFOptions)
		if err != nil {
			return nil, err
		}

		merged, err := mergePDFBuffers([][]byte{headerPdf, contentPdf, footerPdf})
		if err != nil {
			return nil, err
		}
		return merged, nil
	}

	// receipt -> use shared/receipt
	if contractType == "receipt" {
		html, err := RenderTemplateHTML("shared/receipt", contractData, contractType)
		if err != nil {
			return nil, err
		}
		return generatePDF(ctx, html, defaultPDFOptions)
	}

	// fallback: render using the normal contract TSX path (RenderContractHTML)
	html, err := RenderContractHTML(contractType, contractData)
	if err != nil {
		return nil, err
	}
	return generatePDF(ctx, html, defaultPDFOptions)
}

func chromeExecPath() string {
	if p := strings.TrimSpace(os.Getenv("CHROME_PATH")); p != "" {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	paths := []string{
		"/headless-shell/headless-shell",
		"/headless-shell",
		"/usr/bin/headless-shell",
	}
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	candidates := []string{"chromium", "chromium-browser", "google-chrome", "chrome", "msedge"}
	for _, name := range candidates {
		if p, err := exec.LookPath(name); err == nil {
			return p
		}
	}

	if runtime.GOOS == "windows" {
		wins := []string{
			filepath.Join(os.Getenv("ProgramFiles"), "Google", "Chrome", "Application", "chrome.exe"),
			filepath.Join(os.Getenv("ProgramFiles(x86)"), "Google", "Chrome", "Application", "chrome.exe"),
			filepath.Join(os.Getenv("LOCALAPPDATA"), "Google", "Chrome", "Application", "chrome.exe"),
			filepath.Join(os.Getenv("LOCALAPPDATA"), "Chromium", "Application", "chrome.exe"),
			filepath.Join(os.Getenv("ProgramFiles"), "Microsoft", "Edge", "Application", "msedge.exe"),
			filepath.Join(os.Getenv("ProgramFiles(x86)"), "Microsoft", "Edge", "Application", "msedge.exe"),
			filepath.Join(os.Getenv("LOCALAPPDATA"), "Microsoft", "Edge", "Application", "msedge.exe"),
		}
		for _, p := range wins {
			if p != "" {
				if _, err := os.Stat(p); err == nil {
					return p
				}
			}
		}
	}

	return ""
}

// chromeLogFile returns the log file path for Chrome, defaulting to /tmp/chrome.log.
// It also ensures the directory exists.
func chromeLogFile() string {
	if p := strings.TrimSpace(os.Getenv("CHROME_LOG_FILE")); p != "" {
		dir := filepath.Dir(p)
		_ = os.MkdirAll(dir, 0o755)
		return p
	}
	_ = os.MkdirAll("/tmp", 0o755)
	return "/tmp/chrome.log"
}

func calcUnitPriceRounded(currencyAmount float64, itemAmount float64) string {
	if itemAmount == 0 {
		return "0"
	}
	return fmt.Sprintf("%.0f", math.Round(currencyAmount/itemAmount))
}

func ParseContractDataFromRequest(r *http.Request, rawCreatedAt string) (time.Time, error) {
	_ = r
	if rawCreatedAt == "" {
		return time.Time{}, errors.New("createdAt is required")
	}
	if t, err := time.Parse(time.RFC3339Nano, rawCreatedAt); err == nil {
		return t, nil
	}
	if t, err := time.Parse(time.RFC3339, rawCreatedAt); err == nil {
		return t, nil
	}
	return time.Time{}, fmt.Errorf("invalid createdAt: %q", rawCreatedAt)
}
