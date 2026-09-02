package managedimport

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"strconv"
	"strings"
)

func sourceBytes(request PrepareRequest) ([]byte, error) {
	if request.SourceBase64 != "" {
		value, err := base64.StdEncoding.DecodeString(request.SourceBase64)
		if err != nil {
			return nil, fmt.Errorf("INVALID_SOURCE_BASE64: %w", err)
		}
		return value, nil
	}
	return []byte(request.SourceContent), nil
}

func parseSource(request PrepareRequest, data []byte) ([]map[string]any, error) {
	switch request.SourceType {
	case "JSON":
		return parseJSON(data, request.SourceCollection)
	case "CSV":
		return parseCSV(bytes.NewReader(data))
	case "XLSX":
		return parseXLSX(data, request.SourceSheet)
	default:
		return nil, fmt.Errorf("UNSUPPORTED_SOURCE_TYPE: %s", request.SourceType)
	}
}

func parseJSON(data []byte, collection string) ([]map[string]any, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil, fmt.Errorf("INVALID_JSON: %w", err)
	}
	if decoder.Decode(&struct{}{}) != io.EOF {
		return nil, fmt.Errorf("INVALID_JSON: trailing value")
	}
	if object, ok := value.(map[string]any); ok {
		key := collection
		if key == "" {
			key = "records"
		}
		value = object[key]
	}
	items, ok := value.([]any)
	if !ok {
		return nil, fmt.Errorf("INVALID_JSON: source collection is not an array")
	}
	rows := make([]map[string]any, 0, len(items))
	for index, item := range items {
		row, ok := item.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("INVALID_JSON: row %d is not an object", index+1)
		}
		rows = append(rows, row)
	}
	return rows, nil
}

func parseCSV(reader io.Reader) ([]map[string]any, error) {
	parser := csv.NewReader(reader)
	parser.TrimLeadingSpace = true
	records, err := parser.ReadAll()
	if err != nil || len(records) == 0 {
		return nil, fmt.Errorf("INVALID_CSV: %w", err)
	}
	headers := records[0]
	seen := map[string]bool{}
	for _, header := range headers {
		header = strings.TrimSpace(header)
		if header == "" || seen[header] {
			return nil, fmt.Errorf("INVALID_CSV: empty or duplicate header %q", header)
		}
		seen[header] = true
	}
	rows := make([]map[string]any, 0, len(records)-1)
	for index, record := range records[1:] {
		if len(record) != len(headers) {
			return nil, fmt.Errorf("INVALID_CSV: row %d column count", index+2)
		}
		row := map[string]any{}
		for column, value := range record {
			row[strings.TrimSpace(headers[column])] = strings.TrimSpace(value)
		}
		rows = append(rows, row)
	}
	return rows, nil
}

type xlsxSharedStrings struct {
	Items []struct {
		Text string `xml:"t"`
	} `xml:"si"`
}
type xlsxWorksheet struct {
	Rows []struct {
		Cells []struct {
			Ref   string `xml:"r,attr"`
			Type  string `xml:"t,attr"`
			Value string `xml:"v"`
		} `xml:"c"`
	} `xml:"sheetData>row"`
}

func parseXLSX(data []byte, _ string) ([]map[string]any, error) {
	archive, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("INVALID_XLSX: %w", err)
	}
	files := map[string]*zip.File{}
	for _, file := range archive.File {
		files[file.Name] = file
	}
	shared := []string{}
	if file := files["xl/sharedStrings.xml"]; file != nil {
		var values xlsxSharedStrings
		if err := decodeXMLFile(file, &values); err != nil {
			return nil, fmt.Errorf("INVALID_XLSX: %w", err)
		}
		for _, item := range values.Items {
			shared = append(shared, item.Text)
		}
	}
	file := files["xl/worksheets/sheet1.xml"]
	if file == nil {
		return nil, fmt.Errorf("INVALID_XLSX: first worksheet missing")
	}
	var sheet xlsxWorksheet
	if err := decodeXMLFile(file, &sheet); err != nil {
		return nil, fmt.Errorf("INVALID_XLSX: %w", err)
	}
	table := [][]string{}
	for _, sourceRow := range sheet.Rows {
		row := []string{}
		for _, cell := range sourceRow.Cells {
			column := xlsxColumn(cell.Ref)
			for len(row) <= column {
				row = append(row, "")
			}
			value := cell.Value
			if cell.Type == "s" {
				index, _ := strconv.Atoi(value)
				if index >= 0 && index < len(shared) {
					value = shared[index]
				}
			}
			row[column] = value
		}
		table = append(table, row)
	}
	if len(table) == 0 {
		return nil, fmt.Errorf("INVALID_XLSX: worksheet is empty")
	}
	buffer := &bytes.Buffer{}
	writer := csv.NewWriter(buffer)
	width := len(table[0])
	for _, row := range table {
		for len(row) < width {
			row = append(row, "")
		}
		_ = writer.Write(row[:width])
	}
	writer.Flush()
	return parseCSV(buffer)
}

func decodeXMLFile(file *zip.File, target any) error {
	reader, err := file.Open()
	if err != nil {
		return err
	}
	defer reader.Close()
	return xml.NewDecoder(reader).Decode(target)
}

func xlsxColumn(reference string) int {
	column := 0
	for _, char := range reference {
		if char < 'A' || char > 'Z' {
			break
		}
		column = column*26 + int(char-'A'+1)
	}
	if column == 0 {
		return 0
	}
	return column - 1
}
