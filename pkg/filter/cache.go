package filter

import (
	"fmt"
	log "github.com/mhchlib/logger"
	"github.com/mhchlib/mconfig/pkg/cache"
	"github.com/mhchlib/mconfig/pkg/mconfig"
	"github.com/mhchlib/mconfig/pkg/store"
	"sync"
)

type FilterCacheKey struct {
	appKey mconfig.AppKey
	env    mconfig.ConfigEnv
}

type FilterCacheValue struct {
	weight int
	code   mconfig.FilterVal
	mode   mconfig.FilterMode
}

var filterCache *cache.Cache

func initCache() {
	filterCache = cache.NewCache()
}

func PutFilterToCache(appKey mconfig.AppKey, env mconfig.ConfigEnv, val *mconfig.FilterStoreVal) error {
	key := &FilterCacheKey{
		appKey: appKey,
		env:    env,
	}
	return filterCache.PutCache(*key, &FilterCacheValue{
		weight: val.Weight,
		code:   val.Code,
		mode:   val.Mode,
	})
}

func DeleteFilterFromCacheByApp(appKey mconfig.AppKey) error {
	err := filterCache.ExecuteForEachItem(func(key cache.CacheKey, value cache.CacheValue, param ...interface{}) {
		k := key.(FilterCacheKey)
		if appKey == k.appKey {
			_ = filterCache.DeleteCache(k)
			log.Info("recycle filter cache with app key:", fmt.Sprintf("%+v", k))
		}
	})
	if err != nil {
		return err
	}
	return nil
}

func GetFilterFromCache(appKey mconfig.AppKey) ([]*mconfig.FilterEntity, error) {
	filters := make([]*mconfig.FilterEntity, 0)
	mutex := sync.Mutex{}
	//for key, value := range cacheMap {
	//	k := key.(FilterCacheKey)
	//	v := value.(*FilterCacheValue)
	//	if appKey == k.appKey {
	//		filters = append(filters, &mconfig.FilterEntity{
	//			Env:    k.env,
	//			Weight: v.weight,
	//			Code:   v.code,
	//			Mode:   v.mode,
	//		})
	//	}
	//}
	//if len(filters) == 0 {
	//	return nil, cache.ERROR_CACHE_NOT_FOUND
	//}
	err := filterCache.ExecuteForEachItem(func(key cache.CacheKey, value cache.CacheValue, param ...interface{}) {
		k := key.(FilterCacheKey)
		v := value.(*FilterCacheValue)
		if appKey == k.appKey {
			mutex.Lock()
			filters = append(filters, &mconfig.FilterEntity{
				Env:    k.env,
				Weight: v.weight,
				Code:   v.code,
				Mode:   v.mode,
			})
			mutex.Unlock()
		}
	})
	if err != nil {
		return nil, err
	}
	if len(filters) == 0 {
		return nil, cache.ERROR_CACHE_NOT_FOUND
	}
	return filters, nil
}

func getFilterByAppKey(appKey mconfig.AppKey) ([]*mconfig.FilterEntity, error) {
	var filters []*mconfig.FilterEntity
	filters, _ = GetFilterFromCache(appKey)
	if filters == nil {
		appFilters, err := store.GetCurrentMConfigStore().GetAppFilters(appKey)
		if err != nil {
			return nil, err
		}
		filters = appFilters
		//sync to cache
		for _, filter := range appFilters {
			_ = filterCache.PutCache(FilterCacheKey{
				appKey: appKey,
				env:    filter.Env,
			}, &FilterCacheValue{
				weight: filter.Weight,
				code:   filter.Code,
				mode:   filter.Mode,
			})
		}
	}
	for _, filter := range filters {
		log.Info(fmt.Sprintf("%v", filter))
	}
	return filters, nil
}
