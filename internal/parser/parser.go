package parser

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type ParsedTable struct {
	Headers []string
	Rows    []map[string]string
}

func ParseBestTable(html string) (*ParsedTable, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, err
	}

	bestTable := findBestTable(doc)
	if bestTable == nil || bestTable.Length() == 0 {
		return nil, fmt.Errorf("no suitable table found in html")
	}

	headers := extractHeaders(bestTable)
	if len(headers) == 0 {
		return nil, fmt.Errorf("could not extract headers")
	}

	rows := extractRows(bestTable, headers)
	if len(rows) == 0 {
		return nil, fmt.Errorf("no data rows found in table")
	}

	return &ParsedTable{
		Headers: headers,
		Rows:    rows,
	}, nil
}

func findBestTable(doc *goquery.Document) *goquery.Selection {
	var best *goquery.Selection
	maxRows := 0

	doc.Find("table.wikitable, table.sortable").Each(func(_ int, table *goquery.Selection) {
		rowCount := table.Find("tr").Length()
		if rowCount > maxRows {
			maxRows = rowCount
			best = table
		}
	})

	if best != nil {
		return best
	}

	first := doc.Find("table").First()
	return first
}

func extractHeaders(table *goquery.Selection) []string {
	var headers []string

	headerRow := table.Find("tr").First()
	headerRow.Find("th").Each(func(_ int, th *goquery.Selection) {
		text := cleanText(th.Text())
		if text != "" {
			headers = append(headers, text)
		}
	})

	if len(headers) == 0 {
		headerRow.Find("td").Each(func(_ int, td *goquery.Selection) {
			text := cleanText(td.Text())
			if text != "" {
				headers = append(headers, text)
			}
		})
	}

	return headers
}

func extractRows(table *goquery.Selection, headers []string) []map[string]string {
	var rows []map[string]string

	table.Find("tr").Each(func(i int, tr *goquery.Selection) {
		if i == 0 {
			return
		}

		var values []string
		tr.Find("td").Each(func(_ int, td *goquery.Selection) {
			values = append(values, cleanText(td.Text()))
		})

		if len(values) == 0 {
			return
		}

		row := make(map[string]string)
		for j, header := range headers {
			if j < len(values) {
				row[header] = values[j]
			} else {
				row[header] = ""
			}
		}

		rows = append(rows, row)
	})

	return rows
}

func cleanText(s string) string {
	s = strings.Join(strings.Fields(strings.TrimSpace(s)), " ")
	s = strings.ReplaceAll(s, "\u00a0", " ")
	return s
}
