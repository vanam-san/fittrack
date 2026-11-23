package database

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"sort"
	"strconv"
	"strings"
)

func writeArchive(w io.Writer, format string, data records) error {
	switch format {
	case "json":
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		return encoder.Encode(archive{Version: portableVersion, Tables: data})
	case "csv":
		writer := csv.NewWriter(w)
		header := csvHeader()
		if err := writer.Write(header); err != nil {
			return err
		}
		fields := header[3:]
		version := strconv.Itoa(portableVersion)
		count := 0
		for _, table := range portableTables {
			for _, row := range data[table.name] {
				line := make([]string, len(header))
				line[0], line[1] = version, table.name
				for i, field := range fields {
					switch value := row[field].(type) {
					case string:
						line[i+3] = csvEscape.Replace(value)
					case nil:
					default:
						line[i+3] = fmt.Sprint(value)
					}
				}
				if err := writer.Write(line); err != nil {
					return err
				}
				count++
			}
		}
		footer := make([]string, len(header))
		footer[0], footer[1], footer[2] = version, "__end__", strconv.Itoa(count)
		if err := writer.Write(footer); err != nil {
			return err
		}
		writer.Flush()
		return writer.Error()
	default:
		return fmt.Errorf("choose CSV or JSON")
	}
}

func readArchive(r io.Reader, format string) (records, error) {
	switch format {
	case "csv":
		return readCSV(r)
	case "json":
		decoder := json.NewDecoder(r)
		decoder.UseNumber()
		decoder.DisallowUnknownFields()
		var file archive
		if err := decoder.Decode(&file); err != nil {
			return nil, err
		}
		if file.Version != portableVersion {
			return nil, fmt.Errorf("unsupported JSON archive version %d", file.Version)
		}
		if err := decoder.Decode(new(any)); err != io.EOF {
			return nil, fmt.Errorf("unexpected content after JSON archive")
		}
		return file.Tables, nil
	default:
		return nil, fmt.Errorf("choose a FitTrack CSV or JSON export")
	}
}

// encoding/csv normalizes CRLF inside quoted cells on decoding. Escape CR and
// backslash to preserve text entered on any platform.
var csvEscape = strings.NewReplacer("\\", "\\\\", "\r", "\\r")
var csvUnescape = strings.NewReplacer("\\\\", "\\", "\\r", "\r")

func csvHeader() []string {
	seen := make(map[string]bool)
	var fields []string
	for _, table := range portableTables {
		for _, field := range strings.Fields(table.fields) {
			if !seen[field] {
				seen[field] = true
				fields = append(fields, field)
			}
		}
	}
	sort.Strings(fields)
	return append([]string{"fittrack_version", "table", "record_count"}, fields...)
}

func readCSV(r io.Reader) (records, error) {
	reader := csv.NewReader(r)
	header, err := reader.Read()
	if err != nil {
		return nil, err
	}
	header[0] = strings.TrimPrefix(header[0], "\ufeff")
	if !slices.Equal(header, csvHeader()) {
		return nil, fmt.Errorf("expected the FitTrack version 1 CSV header")
	}
	data := make(records, len(portableTables))
	columns := make(map[string][]string, len(portableTables))
	for _, table := range portableTables {
		columns[table.name] = strings.Fields(table.fields)
		data[table.name] = make([]map[string]any, 0)
	}
	version := strconv.Itoa(portableVersion)
	count := 0
	for {
		line, err := reader.Read()
		if err == io.EOF {
			return nil, fmt.Errorf("missing CSV record count")
		}
		if err != nil {
			return nil, err
		}
		lineNumber, _ := reader.FieldPos(0)
		name := line[1]
		if line[0] != version {
			return nil, fmt.Errorf("CSV line %d has an unknown version", lineNumber)
		}
		if name == "__end__" {
			if line[2] != strconv.Itoa(count) || strings.Join(line[3:], "") != "" {
				return nil, fmt.Errorf("CSV record count does not match the archive")
			}
			if _, err := reader.Read(); err != io.EOF {
				return nil, fmt.Errorf("unexpected content after CSV record count")
			}
			return data, nil
		}
		if columns[name] == nil || line[2] != "" {
			return nil, fmt.Errorf("CSV line %d has an unknown version or table", lineNumber)
		}
		row := make(map[string]any, len(columns[name]))
		for i, field := range header[3:] {
			text := line[i+3]
			if !slices.Contains(columns[name], field) {
				if text != "" {
					return nil, fmt.Errorf("CSV line %d has an unexpected %s value", lineNumber, field)
				}
				continue
			}
			if text == "" && field == "food_id" {
				row[field] = nil
				continue
			}
			value, err := parseField(field, csvUnescape.Replace(text))
			if err != nil {
				return nil, fmt.Errorf("CSV line %d, %s: %w", lineNumber, field, err)
			}
			row[field] = value
		}
		data[name] = append(data[name], row)
		count++
	}
}
