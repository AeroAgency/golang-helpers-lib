package helpers

import (
	"encoding/base64"
	"strconv"
	"strings"
)

type FileChecker struct {
	errorFormatter *ErrorFormatter
	validator      *Validator
	rules          map[string]string
}

func NewFileChecker(rules map[string]string) *FileChecker {
	errorFormatter := &ErrorFormatter{}
	validator := &Validator{}
	return &FileChecker{
		errorFormatter: errorFormatter,
		validator:      validator,
		rules:          rules,
	}
}

//  Возвращает максимальное количество файлов для загрузки
func (f *FileChecker) GetMaxFilesCount() int {
	maxFilesCount := 100
	if filesLimitRule, ok := f.rules["files_limit"]; ok {
		maxFilesCount, _ = strconv.Atoi(filesLimitRule)
	}
	return maxFilesCount
}

// Возвращает корректный ли uuid файла
func (f *FileChecker) IsValidFileId(fileId string) bool {
	err := f.validator.ValidateProto(
		&map[string]string{"id": fileId},
		map[string][]string{"id": {"uuid_v4"}},
	)
	if err != nil {
		return false
	}
	return true
}

// Возвращает признак корректности количества загруженных файлов
func (f *FileChecker) IsValidCount(count int) bool {
	maxFilesCount := f.GetMaxFilesCount()
	actualFilesCount := count
	if actualFilesCount > maxFilesCount {
		return false
	}
	return true
}

// Возвращает признак корректности расширения файла
func (f *FileChecker) IsValidExt(ext string) bool {
	if filesExtensionRule, ok := f.rules["file_ext"]; ok {
		extList := strings.Split(filesExtensionRule, ",")
		res := StringInSlice(ext, extList)
		if res != true {
			return false
		}
	}
	return true
}

// Возвращает признак корректности размера файла
func (f *FileChecker) IsValidSize(fileBase64String string) bool {
	if fileSizeRule, ok := f.rules["file_size"]; ok {
		fileBase64DecodeString, _ := base64.StdEncoding.DecodeString(fileBase64String[strings.IndexByte(fileBase64String, ',')+1:])
		fileSize := len(fileBase64DecodeString)
		maxFileSize, _ := strconv.Atoi(fileSizeRule)
		if fileSize > maxFileSize {
			return false
		}
	}
	return true
}
