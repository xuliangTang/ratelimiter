package lru

import (
	"fmt"
	"testing"
	"time"
)

func TestCache_Add(t *testing.T) {
	type args struct {
		key   string
		value interface{}
		ttl   time.Duration
	}
	tests := []struct {
		name string
		args args
	}{
		{"cache1", args{key: "key1", value: "value1", ttl: 0}},
		{"cache2", args{key: "key2", value: "value2", ttl: 0}},
		{"cache3", args{key: "key3", value: "value3", ttl: time.Second * 1}},
	}
	cache := NewCache(WithSize(2))
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache.Add(tt.args.key, tt.args.value, tt.args.ttl)
		})
	}
	if len(cache.edata) != 2 {
		t.Error("error length")
	}
	time.Sleep(time.Second * 2)
	if len(cache.edata) != 1 {
		fmt.Println(len(cache.edata))
		t.Error("error expire length")
	}
}
