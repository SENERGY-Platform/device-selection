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
	"github.com/bradfitz/gomemcache/memcache"
	"log"
)

type GlobalCache struct {
	l1         *memcache.Client
	expiration int32
}

func NewGlobal(urls []string, expiration int32) *GlobalCache {
	return &GlobalCache{l1: memcache.New(urls...), expiration: expiration}
}

func (this *GlobalCache) Get(key string) (value []byte, err error) {
	var temp *memcache.Item
	temp, err = this.l1.Get(key)
	if temp != nil {
		value = temp.Value
	}
	if err == memcache.ErrCacheMiss {
		err = ErrNotFound
	}
	return
}

func (this *GlobalCache) Invalidate() {
	this.l1.DeleteAll()
}

func (this *GlobalCache) Set(key string, value []byte) {
	err := this.l1.Set(&memcache.Item{
		Key:        key,
		Value:      value,
		Expiration: this.expiration,
	})
	if err != nil {
		log.Println("WARNING: err in LocalCache::l1.Set()", err)
	}
	return
}

func (this *GlobalCache) Use(key string, getter func() (interface{}, error), result interface{}) (err error) {
	value, err := this.Get(key)
	if err == nil {
		err = json.Unmarshal(value, result)
		return
	} else if err != ErrNotFound {
		log.Println("WARNING: err in GlobalCache::l1.Get()", err)
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
