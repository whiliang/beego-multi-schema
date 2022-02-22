package ssdb

import (
	"github.com/whiliang/beego-multi-schema/adapter/cache"
	ssdb2 "github.com/whiliang/beego-multi-schema/client/cache/ssdb"
)

// NewSsdbCache create new ssdb adapter.
func NewSsdbCache() cache.Cache {
	return cache.CreateNewToOldCacheAdapter(ssdb2.NewSsdbCache())
}

func init() {
	cache.Register("ssdb", NewSsdbCache)
}
