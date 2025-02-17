package server

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"

	"github.com/VictoriaMetrics/fastcache"
	"github.com/pkg/errors"

	"github.com/bytebase/bytebase/api"
)

var (
	// 32 MiB.
	cacheSize                  = 1024 * 1024 * 32
	_         api.CacheService = (*CacheService)(nil)
)

// CacheService implements a cache.
type CacheService struct {
	cache *fastcache.Cache
}

// NewCacheService creates a cache service.
func NewCacheService() *CacheService {
	return &CacheService{
		cache: fastcache.New(cacheSize),
	}
}

// FindCache finds the value in cache.
func (s *CacheService) FindCache(namespace api.CacheNamespace, id int, entry interface{}) (bool, error) {
	buf1 := []byte{0, 0, 0, 0, 0, 0, 0, 0}
	binary.LittleEndian.PutUint64(buf1, uint64(id))

	buf2, has := s.cache.HasGet(nil, append([]byte(namespace), buf1...))
	if has {
		dec := gob.NewDecoder(bytes.NewReader(buf2))
		if err := dec.Decode(entry); err != nil {
			return false, errors.Wrapf(err, "failed to decode entry for cache namespace: %s", namespace)
		}
		return true, nil
	}

	return false, nil
}

// UpsertCache upserts the value to cache.
func (s *CacheService) UpsertCache(namespace api.CacheNamespace, id int, entry interface{}) error {
	buf1 := []byte{0, 0, 0, 0, 0, 0, 0, 0}
	binary.LittleEndian.PutUint64(buf1, uint64(id))

	var buf2 bytes.Buffer
	enc := gob.NewEncoder(&buf2)
	if err := enc.Encode(entry); err != nil {
		return errors.Wrapf(err, "failed to encode entry for cache namespace: %s", namespace)
	}
	s.cache.Set(append([]byte(namespace), buf1...), buf2.Bytes())

	return nil
}

// DeleteCache deletes the key from cache.
func (s *CacheService) DeleteCache(namespace api.CacheNamespace, id int) {
	buf1 := []byte{0, 0, 0, 0, 0, 0, 0, 0}
	binary.LittleEndian.PutUint64(buf1, uint64(id))
	s.cache.Del(append([]byte(namespace), buf1...))
}
