/*
 * Copyright 2020 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cache

import "errors"

var LocalCacheSize = 40 * 1024 * 1024      //40MB
var LocalCacheExpirationInSec = 600        // 10 min
var GlobalCacheExpirationInSec int32 = 600 // 10 min

var ErrNotFound = errors.New("key not found in cache")

type Cache interface {
	Use(key string, getter func() (interface{}, error), result interface{}) (err error)
	Invalidate()
}

func New(memcachedUrls []string) Cache {
	if len(memcachedUrls) > 0 {
		return NewGlobal(memcachedUrls, GlobalCacheExpirationInSec)
	} else {
		return NewLocal(LocalCacheSize, LocalCacheExpirationInSec)
	}
}
