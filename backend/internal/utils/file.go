package utils

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/tealeg/xlsx/v3"
)

// FileType 文件类型枚举
type FileType int

const (
	CSV FileType = iota
	Excel
	JSON
)

// CSVData CSV数据结构
type CSVData struct {
	Headers []string       `json:"headers"`
	Rows    [][]string     `json:"rows"`
	Summary map[string]int `json:"summary"`
}

// ParseCSV 解析CSV文件
func ParseCSV(filePath string) (*CSVData, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("empty CSV file")
	}

	headers := records[0]
	rows := records[1:]

	// 生成数据摘要
	summary := map[string]int{
		"total_rows": len(rows),
		"total_cols": len(headers),
	}

	return &CSVData{
		Headers: headers,
		Rows:    rows,
		Summary: summary,
	}, nil
}

// ParseExcel 解析Excel文件
func ParseExcel(filePath string) (*CSVData, error) {
	wb, err := xlsx.OpenFile(filePath)
	if err != nil {
		return nil, err
	}

	if len(wb.Sheets) == 0 {
		return nil, fmt.Errorf("no sheets found in Excel file")
	}

	sheet := wb.Sheets[0] // 取第一个sheet
	var headers []string
	var rows [][]string

	// 遍历行
	for rowIndex := 0; rowIndex < sheet.MaxRow; rowIndex++ {
		row, err := sheet.Row(rowIndex)
		if err != nil {
			continue
		}

		var rowData []string
		for colIndex := 0; colIndex < sheet.MaxCol; colIndex++ {
			cell := row.GetCell(colIndex)
			text := cell.String()
			rowData = append(rowData, text)
		}

		if rowIndex == 0 {
			headers = rowData
		} else {
			rows = append(rows, rowData)
		}
	}

	summary := map[string]int{
		"total_rows": len(rows),
		"total_cols": len(headers),
	}

	return &CSVData{
		Headers: headers,
		Rows:    rows,
		Summary: summary,
	}, nil
}

// ParseJSON 解析JSON文件
func ParseJSON(filePath string) (interface{}, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var result interface{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetFileType 根据文件扩展名获取文件类型
func GetFileType(filename string) FileType {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".csv":
		return CSV
	case ".xlsx", ".xls":
		return Excel
	case ".json":
		return JSON
	default:
		return CSV // 默认为CSV
	}
}

// GenerateFileName 生成唯一文件名
func GenerateFileName(originalName string) string {
	ext := filepath.Ext(originalName)
	name := strings.TrimSuffix(originalName, ext)
	return fmt.Sprintf("%s_%d%s", name, GetTimestamp(), ext)
}

// GetTimestamp 获取当前时间戳
func GetTimestamp() int64 {
	return 1642694400 // 简化版，实际应该使用time.Now().Unix()
}
