// categories-gen-go reads data/categories.csv and writes categories.gen.go.
// Run via: go generate ./...
package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"regexp"
	"strings"
	"unicode"
)

func main() {
	in, err := os.Open("data/categories.csv")
	if err != nil {
		fmt.Fprintf(os.Stderr, "open data/categories.csv: %v\n", err)
		os.Exit(1)
	}
	defer in.Close() //nolint:errcheck

	r := csv.NewReader(in)
	records, err := r.ReadAll()
	if err != nil {
		fmt.Fprintf(os.Stderr, "read csv: %v\n", err)
		os.Exit(1)
	}

	f, err := os.Create("categories.gen.go")
	if err != nil {
		fmt.Fprintf(os.Stderr, "create categories.gen.go: %v\n", err)
		os.Exit(1)
	}
	defer f.Close() //nolint:errcheck

	w := bufio.NewWriter(f)
	lines := []string{
		"//go:generate go run ./tools/cmd/categories-gen-go",
		"",
		"package foursquare",
		"",
		"// CategoryID is a Foursquare category identifier.",
		"type CategoryID string",
		"",
		"// Category ID constants from the Foursquare category taxonomy.",
		"// See https://docs.foursquare.com/data-products/docs/categories for the full reference.",
		"//",
		"// Generated from data/categories.csv — do not edit by hand.",
		"const (",
	}
	for _, l := range lines {
		fmt.Fprintln(w, l) //nolint:errcheck
	}
	for i, rec := range records {
		if i == 0 {
			continue // skip header
		}
		if len(rec) != 2 {
			continue
		}
		label, id := strings.TrimSpace(rec[0]), strings.TrimSpace(rec[1])
		fmt.Fprintf(w, "\t%s CategoryID = %q\n", constName(label), id) //nolint:errcheck,gosec
	}
	fmt.Fprintln(w, ")") //nolint:errcheck

	if err := w.Flush(); err != nil {
		fmt.Fprintf(os.Stderr, "write categories.go: %v\n", err)
		os.Exit(1)
	}
}

var nonAlnum = regexp.MustCompile(`[^a-zA-Z0-9 ]+`)

// asciiMap folds accented/special characters to ASCII equivalents.
var asciiMap = strings.NewReplacer(
	"ş", "s", "Ş", "S",
	"ç", "c", "Ç", "C",
	"ü", "u", "Ü", "U",
	"ğ", "g", "Ğ", "G",
	"ı", "i", "İ", "I",
	"ö", "o", "Ö", "O",
	"é", "e", "è", "e", "ê", "e",
	"à", "a", "â", "a",
	"ô", "o",
)

func constName(label string) string {
	s := asciiMap.Replace(label)
	s = strings.ReplaceAll(s, "&", "And")
	s = strings.ReplaceAll(s, "/", "Or")
	// Strip apostrophes so "Children's" stays one token rather than splitting into "Children" + "S".
	s = strings.ReplaceAll(s, "'", "")
	s = nonAlnum.ReplaceAllString(s, " ")
	parts := strings.FieldsFunc(s, func(r rune) bool { return unicode.IsSpace(r) })
	var b strings.Builder
	b.WriteString("Category")
	for _, p := range parts {
		if p == "" {
			continue
		}
		r := []rune(p)
		r[0] = unicode.ToUpper(r[0])
		b.WriteString(string(r))
	}
	return b.String()
}
