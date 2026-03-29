package read

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/xuri/excelize/v2"
)

type XlsxData struct {
	BaseText        string
	PosText         string
	PronText        string
	DefinitionText  string
	LearnerExamples string
	SeriesName      string
}

type ResourceReader struct {
}

func NewResourceReader() *ResourceReader {
	return &ResourceReader{}
}

func (r *ResourceReader) getResourceDir() string {
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return filepath.Join(wd, "resources")
}

func (r *ResourceReader) OpenDirectory() ([]string, error) {
	fileNames := make([]string, 0)
	dirPath := r.getResourceDir()
	dir, err := os.Open(dirPath)
	if err != nil {
		return nil, err
	}
	defer dir.Close()
	if stat, err := dir.Stat(); err != nil || !stat.IsDir() {
		return nil, err
	}

	files, err := dir.Readdirnames(-1)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		fileNames = append(fileNames, file)
	}
	return fileNames, nil
}

func (r *ResourceReader) getCell(row []string, index int) string {
	if index < len(row) {
		return row[index]
	}
	return ""
}

func (r *ResourceReader) OpenXlsxFile(filePath, seriresName string) ([]XlsxData, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("没有找到工作表")
	}

	sheetName := sheets[0]
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, err
	}
	result := make([]XlsxData, 0)

	for idx, row := range rows {
		if idx == 0 {
			continue
		}
		base_text := r.getCell(row, 0)
		pos_text := r.getCell(row, 1)
		pron_text := r.getCell(row, 3)
		definition_text := r.getCell(row, 4)
		learnerexamples_text := r.getCell(row, 5)
		data := XlsxData{
			SeriesName:      seriresName,
			BaseText:        base_text,
			PosText:         pos_text,
			PronText:        pron_text,
			DefinitionText:  definition_text,
			LearnerExamples: learnerexamples_text,
		}
		result = append(result, data)
	}
	return result, nil
}

func (r *ResourceReader) removeExtension(fileName string) string {
	return fileName[:len(fileName)-len(filepath.Ext(fileName))]
}

func (r *ResourceReader) Run() (map[string][]XlsxData, error) {
	result := make(map[string][]XlsxData)
	baseDir := r.getResourceDir()
	fileNames, err := r.OpenDirectory()
	if err != nil {
		return nil, err
	}
	for _, fileName := range fileNames {
		name := r.removeExtension(fileName)
		data, err := r.OpenXlsxFile(filepath.Join(baseDir, fileName), name)
		if err != nil {
			return nil, err
		}
		result[name] = data
	}
	return result, nil
}
