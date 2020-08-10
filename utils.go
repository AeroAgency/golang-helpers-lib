package helpers

func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
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
