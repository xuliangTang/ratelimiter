package lru

import (
	"container/list"
	"fmt"
	"sync"
)

type cacheData struct {
	key   string
	value interface{}
}

func newCacheData(key string, value interface{}) *cacheData {
	return &cacheData{key: key, value: value}
}

type CacheOption func(cache *Cache)
type CacheOptions []CacheOption

// 应用可变参数
func (this CacheOptions) apply(cache *Cache) {
	for _, fn := range this {
		fn(cache)
	}
}

func WithSize(size int64) CacheOption {
	return func(cache *Cache) {
		cache.size = size
	}
}

type Cache struct {
	mu    sync.Mutex
	size  int64
	elist *list.List // 双向链表
	edata map[string]*list.Element
}

func NewCache(opts ...CacheOption) *Cache {
	cache := &Cache{elist: list.New(), edata: make(map[string]*list.Element)}
	// 应用可变参数
	CacheOptions(opts).apply(cache)
	return cache
}

func (this *Cache) Add(key string, value interface{}) {
	this.mu.Lock()
	defer this.mu.Unlock()

	newCache := newCacheData(key, value)

	if getV, ok := this.edata[key]; ok {
		getV.Value = newCache
		this.elist.MoveToFront(getV)
	} else {
		v := this.elist.PushFront(newCache)
		this.edata[key] = v

		// 判断容量是否溢出
		if this.size > 0 && int64(len(this.edata)) > this.size {
			this.removeOldest()
		}
	}
}

func (this *Cache) Get(key string) interface{} {
	this.mu.Lock()
	defer this.mu.Unlock()

	if getV, ok := this.edata[key]; ok {
		this.elist.MoveToFront(getV)
		return getV.Value.(*cacheData).value
	}
	return nil
}

// 末位淘汰一个缓存
func (this *Cache) removeOldest() {
	back := this.elist.Back()
	if back == nil {
		return
	}
	this.removeItem(back)
}

func (this *Cache) removeItem(ele *list.Element) {
	key := ele.Value.(*cacheData).key
	delete(this.edata, key) // 删除map元素
	this.elist.Remove(ele)  // 删除链表元素
}

func (this *Cache) Print() {
	elist := this.elist.Front()
	for elist != nil {
		fmt.Println(elist.Value.(*cacheData).value)
		elist = elist.Next()
	}
}
