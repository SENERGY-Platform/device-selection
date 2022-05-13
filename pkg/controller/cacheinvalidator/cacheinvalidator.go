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

package cacheinvalidator

import (
	"context"
	"device-selection/pkg/configuration"
	"device-selection/pkg/controller/cache"
	"device-selection/pkg/controller/cacheinvalidator/kafka"
	"fmt"
)

func StartCacheInvalidator(ctx context.Context, config configuration.Config, cache cache.Cache) error {
	for _, topic := range config.KafkaTopicsForCacheInvalidation {
		err := kafka.NewConsumer(ctx, config, topic, func(delivery []byte) error {
			cache.Invalidate()
			return nil
		})
		if err != nil {
			return fmt.Errorf("unable to start kafka consumer for cache invalidation on topic %v: %w", topic, err)
		}
	}
	return nil
}
