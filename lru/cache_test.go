package lru

import (
	"testing"
)

func TestCache_Add(t *testing.T) {
	type args struct {
		key   string
		value interface{}
	}
	tests := []struct {
		name string
		args args
	}{
		{"cache1", args{key: "key1", value: "value1"}},
		{"cache2", args{key: "key2", value: "value2"}},
		{"cache3", args{key: "key3", value: "value3"}},
	}
	cache := NewCache(WithSize(2))
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache.Add(tt.args.key, tt.args.value)
		})
	}
	if len(cache.edata) != 2 {
		t.Error("error length")
	}
}
