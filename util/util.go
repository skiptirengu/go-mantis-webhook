package util

import "reflect"

func MapStringSlice(slc interface{}, f func(val interface{}) (string)) ([]string) {
	s := reflect.ValueOf(slc)
	if s.Kind() != reflect.Slice {
		return make([]string, 0)
	}

	res := make([]string, s.Len())
	for i := 0; i < s.Len(); i++ {
		res[i] = f(s.Index(i).Interface())
	}

	return res
}
