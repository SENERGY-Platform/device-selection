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

import (
	"encoding/json"
	"github.com/coocood/freecache"
	"log"
)

type LocalCache struct {
	l1         *freecache.Cache
	expiration int
}

func NewLocal(cacheSize int, expiration int) *LocalCache {
	return &LocalCache{l1: freecache.NewCache(cacheSize), expiration: expiration}
}

func (this *LocalCache) Get(key string) (value []byte, err error) {
	value, err = this.l1.Get([]byte(key))
	if err == freecache.ErrNotFound {
		err = ErrNotFound
	}
	return
}

func (this *LocalCache) Set(key string, value []byte) {
	err := this.l1.Set([]byte(key), value, this.expiration)
	if err != nil {
		log.Println("WARNING: err in LocalCache::l1.Set()", err)
	}
	return
}

func (this *LocalCache) Use(key string, getter func() (interface{}, error), result interface{}) (err error) {
	value, err := this.Get(key)
	if err == nil {
		err = json.Unmarshal(value, result)
		return
	} else if err != ErrNotFound {
		log.Println("WARNING: err in LocalCache::l1.Get()", err)
	}
	temp, err := getter()
	if err != nil {
		return err
	}
	value, err = json.Marshal(temp)
	if err != nil {
		return err
	}
	this.Set(key, value)
	return json.Unmarshal(value, &result)
}
