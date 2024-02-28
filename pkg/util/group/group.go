package group

import gocache "github.com/patrickmn/go-cache"

var cache gocache.Cache

func init() {
	cache = *gocache.New(gocache.NoExpiration, gocache.NoExpiration)
}
