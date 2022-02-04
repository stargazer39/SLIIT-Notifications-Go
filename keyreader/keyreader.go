package keyreader

import (
	"log"
	"reflect"
	"strings"
)

type KeyReader struct {
	i   interface{}
	key string
	r   *reflect.Value
	m   map[string]string
}

func NewReader(i interface{}, key string) *KeyReader {
	r := reflect.ValueOf(i)

	return &KeyReader{
		i:   i,
		key: key,
		r:   &r,
		m:   make(map[string]string),
	}
}

func (k *KeyReader) Get(s string) string {
	if str, ok := k.m[s]; ok {
		return str
	}

	v, err := k.r.Type().FieldByName(s)

	if !err {
		log.Panic(err)
	}

	t := strings.Split(v.Tag.Get(k.key), ",")

	if len(t) == 0 {
		return ""
	}

	k.m[s] = t[0]
	return t[0]
}
