package helpers

import (
	uuid "github.com/satori/go.uuid"
	"strconv"
	"time"
)

func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func SliceIndex(element string, data []string) int {
	for k, v := range data {
		if element == v {
			return k
		}
	}
	return -1 //not found.
}

// Возвращает размер файла в байтах
func CalcOrigBinaryLength(fileBase64String string) int {
	l := len(fileBase64String)
	// count how many trailing '=' there are (if any)
	eq := 0
	if l >= 2 {
		if fileBase64String[l-1] == '=' {
			eq++
		}
		if fileBase64String[l-2] == '=' {
			eq++
		}
		l -= eq
	}
	return (l*3 - eq) / 4
}

// Возвращает uuid из переданной строки
func GetUuidByString(input string) uuid.UUID {
	uuid, _ := uuid.FromString(input)
	return uuid
}

// Возвращает timestamp для переданной даты в формате "dd.mm.yyyy"
func GetDateTimeTs(date string) int64 {
	layout := "02.01.2006"
	t, _ := time.Parse(layout, date)
	return t.Unix()
}

// Возвращает timestamp для переданной даты в формате "dd.mm.yyyy hh:mm:ss"
func GetDateWithTimeTs(date string) int64 {
	layout := "02.01.2006 15:04:05"
	t, _ := time.Parse(layout, date)
	return t.Unix()
}

// Возвращает bool по строковому значению 1 / 0
func GetBoolFromString(value string) bool {
	intValue, _ := strconv.Atoi(value)
	return intValue == 1
}
