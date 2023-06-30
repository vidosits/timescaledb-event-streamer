/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements. See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License. You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package sysconfig

import (
	"github.com/jackc/pgx/v5"
	"github.com/noctarius/timescaledb-event-streamer/internal/sysconfig/defaultproviders"
	spiconfig "github.com/noctarius/timescaledb-event-streamer/spi/config"
	"github.com/noctarius/timescaledb-event-streamer/spi/sink"
	"github.com/noctarius/timescaledb-event-streamer/spi/statestorage"
	"github.com/noctarius/timescaledb-event-streamer/spi/topic/namingstrategy"
)

type SystemConfig struct {
	*spiconfig.Config

	PgxConfig              *pgx.ConnConfig
	SinkProvider           sink.Provider
	EventEmitterProvider   defaultproviders.EventEmitterProvider
	NamingStrategyProvider namingstrategy.Provider
	StateStorageProvider   statestorage.Provider
}

func NewSystemConfig(config *spiconfig.Config) *SystemConfig {
	sc := &SystemConfig{
		Config: config,
	}
	sc.SinkProvider = defaultproviders.DefaultSinkProvider
	sc.EventEmitterProvider = defaultproviders.DefaultEventEmitterProvider
	sc.NamingStrategyProvider = defaultproviders.DefaultNamingStrategyProvider
	sc.StateStorageProvider = defaultproviders.DefaultStateStorageProvider
	return sc
}
