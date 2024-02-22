package lru

import (
	"container/list"
	"fmt"
	"sync"
	"time"
)

// InfDuration is the duration returned by Delay when a Reservation is not OK.
const InfDuration = time.Duration(1<<63 - 1)

type cacheData struct {
	key      string
	value    interface{}
	expireAt time.Time
}

func newCacheData(key string, value interface{}, expireAt time.Time) *cacheData {
	return &cacheData{key: key, value: value, expireAt: expireAt}
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
	cache.clearExpired()
	return cache
}

func (this *Cache) Add(key string, value interface{}, ttl time.Duration) {
	this.mu.Lock()
	defer this.mu.Unlock()

	var setExpire time.Time
	if ttl == 0 { // 0代表不过期
		setExpire = time.Now().Add(InfDuration)
	} else {
		setExpire = time.Now().Add(ttl)
	}
	newCache := newCacheData(key, value, setExpire)

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
		// 判断是否过期
		if time.Now().After(getV.Value.(*cacheData).expireAt) {
			this.removeItem(getV)
			return nil
		}
		this.elist.MoveToFront(getV)
		return getV.Value.(*cacheData).value
	}
	return nil
}

func (this *Cache) Length() int {
	return len(this.edata)
}

// 末位淘汰一个缓存
func (this *Cache) removeOldest() {
	back := this.elist.Back()
	if back == nil {
		return
	}
	this.removeItem(back)
}

// 定时清理过期的缓存
func (this *Cache) clearExpired() {
	go func() {
		for {
			for _, ele := range this.edata {
				if ele.Value.(*cacheData).expireAt.Before(time.Now()) {
					this.removeItem(ele)
				}
			}
			time.Sleep(time.Second * 1)
		}
	}()
}

func (this *Cache) removeItem(ele *list.Element) {
	key := ele.Value.(*cacheData).key
	delete(this.edata, key) // 删除map元素
	this.elist.Remove(ele)  // 删除链表元素
}

func (this *Cache) Print() {
	elist := this.elist.Front()
	for elist != nil {
		// fmt.Println(elist.Value.(*cacheData).value)
		getKey := elist.Value.(*cacheData).key
		fmt.Println(getKey, this.Get(getKey))
		//fmt.Println(this.Get(elist.Value.(*cacheData).key))
		elist = elist.Next()
		fmt.Println("------|||-----", elist)
		time.Sleep(time.Second * 1)
	}
}
