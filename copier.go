package helpers

import "github.com/jinzhu/copier"

type Copier struct{}

func (c Copier) Copy(toValue interface{}, fromValue interface{}) (err error) {
	return copier.Copy(toValue, fromValue)
}
