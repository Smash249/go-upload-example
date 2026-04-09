package read

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xuri/excelize/v2"
)

type XlsxData struct {
	BaseText        string
	PosText         string
	PronText        string
	DefinitionText  string
	LearnerExamples string
	SeriesName      string
	ChineseMeaning  string
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
		if strings.HasPrefix(file, "~$") {
			continue
		}
		if strings.ToLower(filepath.Ext(file)) != ".xlsx" {
			continue
		}
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

	for _, row := range rows {
		baseText := r.getCell(row, 0)
		posText := r.getCell(row, 1)
		pronText := r.getCell(row, 3)
		definitionText := r.getCell(row, 4)
		learnerexamplesText := r.getCell(row, 5)
		chineseMeaning := r.getCell(row, 6)
		data := XlsxData{
			SeriesName:      seriresName,
			BaseText:        baseText,
			PosText:         posText,
			PronText:        pronText,
			DefinitionText:  definitionText,
			LearnerExamples: learnerexamplesText,
			ChineseMeaning:  chineseMeaning,
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
