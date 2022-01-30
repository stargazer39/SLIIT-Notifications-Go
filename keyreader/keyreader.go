package keyreader

import (
	"reflect"
	"strings"
)

type KeyReader struct {
	i   interface{}
	key string
	r   *reflect.Value
}

func NewReader(i interface{}, key string) *KeyReader {
	r := reflect.ValueOf(i)

	return &KeyReader{
		i:   i,
		key: key,
		r:   &r,
	}
}

func (k *KeyReader) Get(s string) string {
	v, err := k.r.Type().FieldByName(s)

	if !err {
		return ""
	}
	t := strings.Split(v.Tag.Get(k.key), ",")

	if len(t) != 0 {
		return ""
	}

	return t[0]
}
