package main

// Read CSV as input, write a plaintext table output.

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

func main() {

	truncateValues := flag.Int("truncate", 0, "truncate values")

	isTab := flag.Bool("tab", false, "process .tsv instead of .csv")
	flag.BoolVar(isTab, "t", false, "process .tsv instead of .csv")
	flag.BoolVar(isTab, "tsv", false, "process .tsv instead of .csv")

	doSort := flag.Bool("sort", false, "sort by first column")
	flag.BoolVar(doSort, "s", false, "sort by first column")

	flag.Parse()

	r := csv.NewReader(os.Stdin)
	r.FieldsPerRecord = -1 // don't freak out about diff len lines
	if *isTab {
		r.Comma = '\t'
	}

	data, err := r.ReadAll()
	if err != nil {
		log.Fatalf("got error reading csv: %v", err)
	}

	maxcols := 0
	for _, rec := range data {
		if maxcols < len(rec) {
			maxcols = len(rec)
		}
	}

	sizes := make([]int, maxcols)
	for _, rec := range data {
		for i := 0; i < maxcols; i++ {
			if i >= len(rec) {
				break
			}
			rec[i] = utf8Truncate(rec[i], *truncateValues)
			if sizes[i] < utf8.RuneCount([]byte(rec[i])) {
				sizes[i] = utf8.RuneCount([]byte(rec[i]))
			}
		}
	}

	if len(data) == 0 {
		return
	}

	showCol := map[int]bool{}
	for i, h := range data[0] {
		showCol[i] = includeColumn(h, flag.Args())
	}

	if *doSort {
		sort.Slice(data, func(i, j int) bool {
			if i == 0 || j == 0 {
				// keep header on top
				return i == 0
			}
			return strings.ToLower(data[i][0]) < strings.ToLower(data[j][0])
		})
	}

	for _, rec := range data {
		for i := 0; i < maxcols; i++ {
			if i >= len(rec) {
				break
			}
			if showCol[i] {
				// fmt.Printf("%-*s  ", sizes[i], rec[i])
				printField(rec[i], sizes[i])
			}
		}
		fmt.Println()
	}
}

func printField(s string, width int) {
	spaces := strings.Repeat(" ", width-utf8.RuneCount([]byte(s)))
	if looksNumeric(s) {
		fmt.Printf("%s%s  ", spaces, s)
	} else {
		fmt.Printf("%s%s  ", s, spaces)
	}
}

func looksNumeric(s string) bool {
	for _, r := range s {
		if !(unicode.IsDigit(r) || r == '.' || r == '-') {
			return false
		}
	}
	return true
}

func utf8Truncate(s string, n int) string {
	if n <= 0 {
		return s
	}

	if utf8.RuneCount([]byte(s)) <= n {
		return s
	}

	sb := strings.Builder{}
	for i, r := range s {
		if i >= n {
			return sb.String()
		}
		sb.WriteRune(r)
	}

	return sb.String()
}

func includeColumn(header string, includeList []string) bool {
	if len(includeList) == 0 {
		return true
	}

	for _, each := range includeList {
		if strings.EqualFold(header, each) {
			return true
		}
	}

	return false
}
