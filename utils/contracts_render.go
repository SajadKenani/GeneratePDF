package utils

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strings"
	"sync"
	"time"
)

type tsxVars struct {
	DepositNumID      string
	UserPhone         string
	UserName          string
	UserEmail         string
	Item              string
	ItemAmount        string
	Currency          string
	CurrencyUpper     string
	CurrencyAmount    string
	PaymentMethod     string
	Realm             string
	DateNumeric       string
	DateShort         string
	InsuranceYesNoEn  string
	InsuranceYesNoAr  string
	InsuranceYears010 string
	UnitPriceRounded  string
}

var (
	tsxReturnCacheMu  sync.Mutex
	tsxReturnCache    = map[string]string{} // tsxPath -> extracted return(...) JSX
	transparentPngB64 = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/xcAAn8B9oBs8LkAAAAASUVORK5CYII="
)

func imageBase64(name string) string {
	name = strings.ToUpper(strings.TrimSpace(name))
	if name == "" {
		return transparentPngB64
	}

	base := name
	alt := strings.ReplaceAll(base, "_", "")

	tryEnv := func(keys ...string) string {
		for _, k := range keys {
			if v := strings.TrimSpace(os.Getenv(k)); v != "" {
				return v
			}
		}
		return ""
	}

	loadPath := func(path string) string {
		data, err := os.ReadFile(path)
		if err != nil {
			return ""
		}
		trimmed := strings.TrimSpace(string(data))
		if trimmed != "" {
			if _, err := base64.StdEncoding.DecodeString(trimmed); err == nil {
				return trimmed
			}
		}
		return base64.StdEncoding.EncodeToString(data)
	}

	if p := tryEnv(base+"_B64_PATH", alt+"_B64_PATH"); p != "" {
		if v := loadPath(p); v != "" {
			return v
		}
	}

	if v := tryEnv(base+"_B64", alt+"_B64"); v != "" {
		return v
	}

	if p := tryEnv(base, alt); p != "" {
		if v := loadPath(p); v != "" {
			return v
		}
	}

	if base == "LOGO" {
		if p := tryEnv("MAIN_LOGO_B64_PATH", "MAINLOGO_B64_PATH"); p != "" {
			if v := loadPath(p); v != "" {
				return v
			}
		}
		if v := tryEnv("MAIN_LOGO_B64", "MAINLOGO_B64"); v != "" {
			return v
		}
		if p := tryEnv("MAIN_LOGO", "MAINLOGO"); p != "" {
			if v := loadPath(p); v != "" {
				return v
			}
		}
	}
	if base == "MAINLOGO" {
		if p := tryEnv("MAIN_LOGO_B64_PATH"); p != "" {
			if v := loadPath(p); v != "" {
				return v
			}
		}
		if v := tryEnv("MAIN_LOGO_B64"); v != "" {
			return v
		}
		if p := tryEnv("MAIN_LOGO"); p != "" {
			if v := loadPath(p); v != "" {
				return v
			}
		}
	}

	return transparentPngB64
}

// RenderContractHTML renders a contract by converting existing TSX templates in contracts/ into plain HTML.
// This stays Go-only (no Node/React execution) by extracting the JSX returned from the component and
// applying a targeted conversion + placeholder substitution.
func RenderContractHTML(contractType string, c ContractData) (string, error) {
	tsxPath := contractTSXPath(contractType)
	if tsxPath == "" {
		// fallback (only for missing templates like withdraw/receipt)
		return minimalFallbackHTML(contractType, c), nil
	}

	jsx, err := loadTSXReturnBlock(tsxPath)
	if err != nil {
		return "", err
	}

	vars := buildTSXVars(c)
	_, title := generateContentHTML(c, vars, contractType)
	htmlOut := tsxToHTML(title, jsx, vars)
	if strings.TrimSpace(htmlOut) == "" {
		return "", fmt.Errorf("tsx conversion produced empty html for %q", tsxPath)
	}
	return htmlOut, nil
}

func contractTSXPath(contractType string) string {
	rel := ""
	if contractType == "receipt" {
		rel = filepath.Join("contracts", "templates", "shared", contractType+".tsx")
	} else {
		rel = filepath.Join("contracts", "templates", "content.tsx")
	}
	if rel == "" {
		return ""
	}

	// Resolve relative to current working directory; in containers this is project root.
	return filepath.Clean(rel)
}

// RenderTemplateHTML renders a TSX template file located under contracts/templates.
// `relPath` may be a simple name like "header" or a nested path like "shared/receipt" (without .tsx).
func RenderTemplateHTML(relPath string, c ContractData, contractType string) (string, error) {
	rel := relPath
	if !strings.HasSuffix(rel, ".tsx") {
		rel = rel + ".tsx"
	}
	tsxPath := filepath.Join("contracts", "templates", filepath.FromSlash(rel))
	// If file doesn't exist, return fallback
	if _, err := os.Stat(tsxPath); err != nil {
		return minimalFallbackHTML(relPath, c), nil
	}

	jsx, err := loadTSXReturnBlock(tsxPath)
	if err != nil {
		return "", err
	}

	vars := buildTSXVars(c)

	// If template expects `content` injected via dangerouslySetInnerHTML or {content},
	// build the content HTML and replace occurrences before conversion.
	if strings.Contains(jsx, "dangerouslySetInnerHTML") || strings.Contains(jsx, "{content}") {
		contentHTML, title := generateContentHTML(c, vars, contractType)
		if contentHTML == "" {
			// fallback to normal path
			htmlOut := tsxToHTML(title, jsx, vars)
			if strings.TrimSpace(htmlOut) == "" {
				return "", fmt.Errorf("tsx conversion produced empty html for %q", tsxPath)
			}
			return htmlOut, nil
		}

		// Insert a unique marker and replace after conversion
		marker := "__INLINE_CONTENT_MARKER__"
		jsx = strings.ReplaceAll(jsx, "dangerouslySetInnerHTML={{ __html: content }}", marker)
		jsx = strings.ReplaceAll(jsx, "dangerouslySetInnerHTML={{__html:content}}", marker)
		jsx = strings.ReplaceAll(jsx, "{content}", marker)

		htmlOut := tsxToHTML(title, jsx, vars)
		htmlOut = strings.ReplaceAll(htmlOut, marker, contentHTML)
		if strings.TrimSpace(htmlOut) == "" {
			return "", fmt.Errorf("tsx conversion produced empty html for %q", tsxPath)
		}
		return htmlOut, nil
	}
	_, title := generateContentHTML(c, vars, contractType)
	htmlOut := tsxToHTML(title, jsx, vars)
	if strings.TrimSpace(htmlOut) == "" {
		return "", fmt.Errorf("tsx conversion produced empty html for %q", tsxPath)
	}
	return htmlOut, nil
}

func buildTSXVars(c ContractData) tsxVars {
	dateNumeric := c.CreatedAt.Format("02/01/2006")
	dateShort := c.CreatedAt.Format("02 Jan 2006")

	insuranceYears := "0"
	if c.Insurance {
		insuranceYears = "1"
	}

	insuranceEn := "No"
	insuranceAr := "لا"
	if c.Insurance {
		insuranceEn = "Yes"
		insuranceAr = "نعم"
	}

	itemAmount := fmt.Sprintf("%.0f", c.ItemAmount)
	currencyAmount := fmt.Sprintf("%.2f", c.CurrencyAmount)
	currencyAmount = strings.TrimSuffix(currencyAmount, ".00")

	return tsxVars{
		DepositNumID:      fmt.Sprintf("%d", c.NumID),
		UserPhone:         c.UserPhone,
		UserName:          c.UserName,
		UserEmail:         c.UserEmail,
		Item:              c.Item,
		ItemAmount:        itemAmount,
		Currency:          c.Currency,
		CurrencyUpper:     strings.ToUpper(c.Currency),
		CurrencyAmount:    currencyAmount,
		PaymentMethod:     c.PaymentMethod,
		Realm:             c.Realm,
		DateNumeric:       dateNumeric,
		DateShort:         dateShort,
		InsuranceYesNoEn:  insuranceEn,
		InsuranceYesNoAr:  insuranceAr,
		InsuranceYears010: insuranceYears,
		UnitPriceRounded:  calcUnitPriceRounded(c.CurrencyAmount, c.ItemAmount),
	}
}

func loadTSXReturnBlock(tsxPath string) (string, error) {
	// Cache by path to avoid reparsing on every request.
	tsxReturnCacheMu.Lock()
	if cached, ok := tsxReturnCache[tsxPath]; ok {
		tsxReturnCacheMu.Unlock()
		return cached, nil
	}
	tsxReturnCacheMu.Unlock()

	b, err := os.ReadFile(tsxPath)
	if err != nil {
		return "", err
	}
	s := string(b)

	jsx, err := extractReturnJSX(s)
	if err != nil {
		return "", fmt.Errorf("extract TSX return block from %q: %w", tsxPath, err)
	}

	tsxReturnCacheMu.Lock()
	tsxReturnCache[tsxPath] = jsx
	tsxReturnCacheMu.Unlock()
	return jsx, nil
}

// extractReturnJSX extracts the JSX between `return (` and the matching `)`.
// It is a small, TSX-specific parser that ignores parentheses inside strings and template literals.
func extractReturnJSX(tsx string) (string, error) {
	idx := strings.Index(tsx, "return (")
	if idx < 0 {
		return "", errors.New("missing `return (`")
	}
	// Start scanning right after "return (".
	i := idx + len("return (")
	depth := 1
	inSingle := false
	inDouble := false
	inBacktick := false
	escaped := false

	for ; i < len(tsx); i++ {
		ch := tsx[i]

		if escaped {
			escaped = false
			continue
		}
		if ch == '\\' {
			// escapes in strings
			if inSingle || inDouble || inBacktick {
				escaped = true
			}
			continue
		}

		if inSingle {
			if ch == '\'' {
				inSingle = false
			}
			continue
		}
		if inDouble {
			if ch == '"' {
				inDouble = false
			}
			continue
		}
		if inBacktick {
			if ch == '`' {
				inBacktick = false
			}
			continue
		}

		switch ch {
		case '\'':
			inSingle = true
		case '"':
			inDouble = true
		case '`':
			inBacktick = true
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				jsx := tsx[idx+len("return (") : i]
				return strings.TrimSpace(jsx), nil
			}
		}
	}

	return "", errors.New("unterminated return(...) block")
}

func tsxToHTML(title string, jsx string, v tsxVars) string {
	s := jsx
	s = regexp.MustCompile(`\{\s*/\*[\s\S]*?\*/\s*\}`).ReplaceAllString(s, "")
	s = strings.ReplaceAll(s, "{\" \"}", " ")
	s = strings.ReplaceAll(s, "{\"\n\"}", "")
	s = strings.ReplaceAll(s, "<Html", "<html")
	s = strings.ReplaceAll(s, "</Html>", "</html>")
	s = strings.ReplaceAll(s, "<Head>", "<head>")
	s = strings.ReplaceAll(s, "</Head>", "</head>")
	s = regexp.MustCompile(`(?is)<Preview[\s\S]*?</Preview>`).ReplaceAllString(s, "")
	s = strings.ReplaceAll(s, "<Heading", "<h2")
	s = strings.ReplaceAll(s, "</Heading>", "</h2>")
	s = strings.ReplaceAll(s, "<Text", "<p")
	s = strings.ReplaceAll(s, "</Text>", "</p>")
	s = strings.ReplaceAll(s, "<Img", "<img")
	s = strings.ReplaceAll(s, "className=", "class=")
	s = regexp.MustCompile(`colSpan=\{(\d+)\}`).ReplaceAllString(s, `colspan="$1"`)
	s = regexp.MustCompile(`rowSpan=\{(\d+)\}`).ReplaceAllString(s, `rowspan="$1"`)
	s = convertStyleObjects(s)
	s = regexp.MustCompile(`src=\{\s*`+"`"+`data:image/png;base64,\$\{(mainLogo|logo|signature)\}`+"`"+`\s*\}`).ReplaceAllStringFunc(s, func(m string) string {
		matches := regexp.MustCompile(`\{\s*` + "`" + `data:image/png;base64,\$\{(mainLogo|logo|signature)\}` + "`" + `\s*\}`).FindStringSubmatch(m)
		if len(matches) != 2 {
			return " src=\"\""
		}
		b64 := imageBase64(matches[1])
		return fmt.Sprintf("src=\"data:image/png;base64,%s\"", html.EscapeString(b64))
	})
	s = regexp.MustCompile(`\s+src=\{[^}]*\}`).ReplaceAllString(s, ` src=""`)
	s = strings.ReplaceAll(s, "{currency.toUpperCase()}", html.EscapeString(v.CurrencyUpper))
	s = strings.ReplaceAll(s, "{currency.toLowerCase()}", html.EscapeString(strings.ToLower(v.Currency)))
	s = strings.ReplaceAll(s, "// ", "")
	s = regexp.MustCompile(`\{\s*new\s+Decimal\(\s*currencyAmount\s*/\s*itemAmount\s*\)\.round\(\)\.toString\(\)\s*\}`).ReplaceAllString(s, html.EscapeString(v.UnitPriceRounded))
	s = regexp.MustCompile(`\{\s*insurance\s*\?\s*1\s*:\s*0\s*\}`).ReplaceAllString(s, html.EscapeString(v.InsuranceYears010))
	s = regexp.MustCompile("\\{`([^`]*)\\$\\{realm\\}([^`]*)`\\}").ReplaceAllString(s, "${1}"+html.EscapeString(v.Realm)+"${2}")
	s = replaceSimpleTSXVar(s, "depositNumId", v.DepositNumID)
	s = replaceSimpleTSXVar(s, "userPhone", v.UserPhone)
	s = replaceSimpleTSXVar(s, "userName", v.UserName)
	s = replaceSimpleTSXVar(s, "userEmail", v.UserEmail)
	s = replaceSimpleTSXVar(s, "item", v.Item)
	s = replaceSimpleTSXVar(s, "itemAmount", v.ItemAmount)
	s = replaceSimpleTSXVar(s, "currency", v.Currency)
	s = replaceSimpleTSXVar(s, "currencyAmount", v.CurrencyAmount)
	s = replaceSimpleTSXVar(s, "paymentMethod", v.PaymentMethod)
	s = replaceSimpleTSXVar(s, "realm", v.Realm)
	s = replaceSimpleTSXVar(s, "title", title)
	// Date variable inside templates is called {date}
	// We pick a reasonable default (numeric); cattle uses short month, but that's also fine.
	dateVal := v.DateNumeric
	if strings.Contains(s, "month: \"short\"") || strings.Contains(strings.ToLower(s), "cow-romanian") {
		dateVal = v.DateShort
	}
	s = replaceSimpleTSXVar(s, "date", dateVal)
	// Some templates precompute insurance string and use {insurance}
	// We don't execute TS; replace with Arabic Yes/No if dir="rtl" is present.
	insuranceVal := v.InsuranceYesNoEn
	if strings.Contains(s, "dir=\"rtl\"") {
		insuranceVal = v.InsuranceYesNoAr
	}
	s = replaceSimpleTSXVar(s, "insurance", insuranceVal)

	// Replace commonProps.* simple references (e.g. {commonProps.userName})
	commonMap := map[string]string{
		"depositNumId":   v.DepositNumID,
		"userName":       v.UserName,
		"userEmail":      v.UserEmail,
		"userPhone":      v.UserPhone,
		"item":           v.Item,
		"itemAmount":     v.ItemAmount,
		"currency":       v.Currency,
		"currencyAmount": v.CurrencyAmount,
		"paymentMethod":  v.PaymentMethod,
		"realm":          v.Realm,
	}

	// commonProps.property
	reCommon := regexp.MustCompile(`\{\s*commonProps\.([a-zA-Z0-9_]+)\s*\}`)
	s = reCommon.ReplaceAllStringFunc(s, func(m string) string {
		sub := reCommon.FindStringSubmatch(m)
		if len(sub) != 2 {
			return ""
		}
		if val, ok := commonMap[sub[1]]; ok {
			return html.EscapeString(val)
		}
		return ""
	})

	// commonProps.property.toUpperCase()
	reCommonUpper := regexp.MustCompile(`\{\s*commonProps\.([a-zA-Z0-9_]+)\.toUpperCase\(\)\s*\}`)
	s = reCommonUpper.ReplaceAllStringFunc(s, func(m string) string {
		sub := reCommonUpper.FindStringSubmatch(m)
		if len(sub) != 2 {
			return ""
		}
		if val, ok := commonMap[sub[1]]; ok {
			return html.EscapeString(strings.ToUpper(val))
		}
		return ""
	})

	// Replace commonProps.insurance ternary expressions with localized value
	reIns := regexp.MustCompile(`\{[^}]*commonProps\.insurance[^}]*\}`)
	s = reIns.ReplaceAllString(s, html.EscapeString(insuranceVal))
	if strings.Contains(s, "dir=\"rtl\"") {
		insuranceVal = v.InsuranceYesNoAr
	}
	s = replaceSimpleTSXVar(s, "insurance", insuranceVal)

	// Remove remaining JSX expression braces to avoid broken HTML.
	// Before stripping generic expressions, handle in-template content arrays like
	// {content?.content?.map((term, index) => ( ... ))} by loading the JSON
	// `contracts/data/{realm}/{contractType}.json` and rendering matching HTML.
	s = inlineContentArray(s, v)

	s = regexp.MustCompile(`\{[^{}]*\}`).ReplaceAllString(s, "")

	// Ensure we return a full HTML document. TSX already includes <html>, <head>, <body>.
	// If it doesn't, wrap it.
	lower := strings.ToLower(s)
	if !strings.Contains(lower, "<html") {
		s = "<html><head><meta charset=\"utf-8\" /></head><body>" + s + "</body></html>"
	}

	return s
}

func replaceSimpleTSXVar(s string, name string, value string) string {
	// Replace { name } occurrences
	re := regexp.MustCompile(`\{\s*` + regexp.QuoteMeta(name) + `\s*\}`)
	return re.ReplaceAllString(s, html.EscapeString(value))
}

// inlineContentArray looks for a common JSX pattern that maps over `content.content`
// and replaces it with HTML generated from the JSON content file for the contract.
func inlineContentArray(s string, v tsxVars) string {
	// quick check
	if !strings.Contains(s, "content?.content?.map") && !strings.Contains(s, "content.content.map") {
		return s
	}

	re := regexp.MustCompile(`(?s)\{\s*content\?\.?content\?\.?\.map\([^}]*?\)\s*\)\s*\}`)
	return re.ReplaceAllStringFunc(s, func(_ string) string {
		// Load JSON file for this contract type and realm
		// try multiple candidate locations to mirror the JS CONTENT_MAP
		dataPath := findDataFile(v.Realm, v.Item)

		if dataPath == "" {
			return ""
		}

		b, err := os.ReadFile(dataPath)
		if err != nil {
			return ""
		}

		var root struct {
			Title   string `json:"title"`
			Content []struct {
				Title   string              `json:"title"`
				Content []map[string]string `json:"content"`
			} `json:"content"`
		}
		if err := json.Unmarshal(b, &root); err != nil {
			return ""
		}

		var out strings.Builder
		for _, term := range root.Content {
			out.WriteString(`<div class="border-b border-gray-200 py-2">`)
			out.WriteString(`<p class="m-0 font-bold text-gray-700" style="margin:0">`)
			out.WriteString(html.EscapeString(term.Title))
			out.WriteString(`</p>`)

			for _, lineObj := range term.Content {
				// Each lineObj has keys like ".1", ".2" etc. We render them in numeric order.
				// Collect and sort keys
				keys := make([]string, 0, len(lineObj))
				for k := range lineObj {
					keys = append(keys, k)
				}
				slices.Sort(keys)
				for _, k := range keys {
					txt := lineObj[k]
					// replace placeholders
					rendered := interpolateJsonPlaceholders(txt, v.CurrencyAmount, v.DateNumeric, v.ItemAmount)
					out.WriteString(`<p class="m-0 p-0 text-sm text-gray-600" style="margin:0;padding:0;line-height:1.4">`)
					// prefix number if possible
					num := strings.TrimPrefix(k, ".")
					out.WriteString(html.EscapeString(fmt.Sprintf("%s- %s", num, rendered)))
					out.WriteString(`</p>`)
				}
			}
			out.WriteString(`</div>`)
		}

		return out.String()
	})
}

func interpolateJsonPlaceholders(text string, currencyAmount string, date string, itemAmount string) string {
	if text == "" {
		return ""
	}
	out := text
	out = strings.ReplaceAll(out, "{currencyAmount}", currencyAmount)
	out = strings.ReplaceAll(out, "{date}", date)
	out = strings.ReplaceAll(out, "{itemAmount}", itemAmount)
	return out
}

// findDataFile searches recursively for {item}.json in contracts/data/{realm}
func findDataFile(realm string, item string) string {
	// Try direct path first: contracts/data/{realm}/{item}.json
	directPath := filepath.Join("contracts", "data", realm, item+".json")
	if _, err := os.Stat(directPath); err == nil {
		return directPath
	}

	// Try nested structure: contracts/data/{realm}/{subdir}/{item}.json
	realmPath := filepath.Join("contracts", "data", realm)
	var foundPath string

	filepath.WalkDir(realmPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, filepath.Join(item+".json")) {
			foundPath = path
			return filepath.SkipDir // Stop searching once found
		}
		return nil
	})

	return foundPath
}

// generateContentHTML produces the HTML used when templates inject `content` via
// dangerouslySetInnerHTML or {content}.
func generateContentHTML(c ContractData, v tsxVars, contractType string) (string, string) {
	// Use contractType to find the correct data file (e.g., "sales", "auth", etc.)
	// If contractType is empty, fall back to using Item
	item := contractType
	if item == "" {
		item = c.Item
	}

	dataPath := findDataFile(c.Realm, item)
	if dataPath == "" {
		return "", ""
	}

	b, err := os.ReadFile(dataPath)
	if err != nil {
		return "", ""
	}

	var root struct {
		Title   string `json:"title"`
		Content []struct {
			Title   string              `json:"title"`
			Content []map[string]string `json:"content"`
		} `json:"content"`
	}
	if err := json.Unmarshal(b, &root); err != nil {
		return "", ""
	}

	var out strings.Builder
	for _, term := range root.Content {
		out.WriteString(`<div class="border-b border-gray-200 py-2">`)
		out.WriteString(`<p class="m-0 font-bold text-gray-700" style="margin:0">`)
		out.WriteString(html.EscapeString(term.Title))
		out.WriteString(`</p>`)

		for _, lineObj := range term.Content {
			keys := make([]string, 0, len(lineObj))
			for k := range lineObj {
				keys = append(keys, k)
			}
			slices.Sort(keys)
			for _, k := range keys {
				txt := lineObj[k]
				rendered := interpolateJsonPlaceholders(txt, v.CurrencyAmount, v.DateNumeric, v.ItemAmount)
				out.WriteString(`<p class="m-0 p-0 text-sm text-gray-600" style="margin:0;padding:0;line-height:1.4">`)
				num := strings.TrimPrefix(k, ".")
				out.WriteString(html.EscapeString(fmt.Sprintf("%s- %s", num, rendered)))
				out.WriteString(`</p>`)
			}
		}

		out.WriteString(`</div>`)
	}

	return out.String(), root.Title
}

func convertStyleObjects(s string) string {
	// This handles style={{ fontFamily: "...", fontSize: "12px" }} by converting to style="font-family: ...; font-size: 12px;"
	re := regexp.MustCompile(`style=\{\{([^}]*)\}\}`)
	return re.ReplaceAllStringFunc(s, func(match string) string {
		m := re.FindStringSubmatch(match)
		if len(m) != 2 {
			return ""
		}
		body := strings.TrimSpace(m[1])
		// try to turn a JS object into a map by a very small heuristic
		// replace single quotes with double quotes and quote keys.
		// Example: fontFamily: "'Tajawal', sans-serif" -> {"fontFamily":"'Tajawal', sans-serif"}
		parts := strings.Split(body, ",")
		obj := map[string]string{}
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			kv := strings.SplitN(p, ":", 2)
			if len(kv) != 2 {
				continue
			}
			k := strings.TrimSpace(kv[0])
			val := strings.TrimSpace(kv[1])
			val = strings.TrimSuffix(strings.TrimPrefix(val, "\""), "\"")
			val = strings.TrimSuffix(strings.TrimPrefix(val, "'"), "'")
			obj[k] = val
		}

		if len(obj) == 0 {
			return ""
		}

		css := ""
		for k, val := range obj {
			cssKey := camelToKebab(k)
			css += fmt.Sprintf("%s: %s; ", cssKey, val)
		}
		css = strings.TrimSpace(css)
		return fmt.Sprintf(`style="%s"`, html.EscapeString(css))
	})
}

func camelToKebab(s string) string {
	var b strings.Builder
	b.Grow(len(s) + 8)
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			if i > 0 {
				b.WriteByte('-')
			}
			b.WriteByte(byte(c + 32))
			continue
		}
		b.WriteByte(c)
	}
	return b.String()
}

func minimalFallbackHTML(contractType string, c ContractData) string {
	// Only used when the repo has no TSX contract template (withdraw/receipt).
	dateStr := c.CreatedAt.Format(time.RFC3339)
	data, _ := json.MarshalIndent(map[string]any{
		"numId":          c.NumID,
		"createdAt":      dateStr,
		"item":           c.Item,
		"itemAmount":     c.ItemAmount,
		"currency":       c.Currency,
		"currencyAmount": c.CurrencyAmount,
		"paymentMethod":  c.PaymentMethod,
		"realm":          c.Realm,
		"userName":       c.UserName,
		"userPhone":      c.UserPhone,
		"userEmail":      c.UserEmail,
		"insurance":      c.Insurance,
	}, "", "  ")

	title := fmt.Sprintf("%s-contract-%d", contractType, c.NumID)
	return fmt.Sprintf(`<!doctype html>
	<html lang="en">
		<head>
		<meta charset="utf-8" />
		<meta name="viewport" content="width=device-width, initial-scale=1" />
		<title>%s</title>
			<style>
				@page { size: A4; margin: 14mm; }
				body { font-family: Arial, sans-serif; color: #111; }
				pre { white-space: pre-wrap; word-break: break-word; font-size: 12px; }
			</style>
		</head>
		<body>
			<h2>%s</h2>
				<p>Template not found in contracts/. Using fallback.</p>
			<pre>%s</pre>
		</body>
	</html>`, html.EscapeString(title), html.EscapeString(title), html.EscapeString(string(data)))
}

func init() {
	// Ensure filepath separator logic is consistent.
	_ = runtime.GOOS
}
