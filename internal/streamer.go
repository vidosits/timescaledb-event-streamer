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

package internal

import (
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/noctarius/timescaledb-event-streamer/internal/erroring"
	"github.com/noctarius/timescaledb-event-streamer/internal/replication"
	"github.com/noctarius/timescaledb-event-streamer/internal/sysconfig"
	spiconfig "github.com/noctarius/timescaledb-event-streamer/spi/config"
	"github.com/noctarius/timescaledb-event-streamer/spi/plugins"
	"github.com/samber/lo"
	"github.com/urfave/cli"

	// Register built-in naming strategies
	_ "github.com/noctarius/timescaledb-event-streamer/spi/namingstrategy"

	// Register built-in offset storages
	_ "github.com/noctarius/timescaledb-event-streamer/spi/statestorage"

	// Register built-in sinks
	_ "github.com/noctarius/timescaledb-event-streamer/internal/eventing/sink/awskinesis"
	_ "github.com/noctarius/timescaledb-event-streamer/internal/eventing/sink/awssqs"
	_ "github.com/noctarius/timescaledb-event-streamer/internal/eventing/sink/kafka"
	_ "github.com/noctarius/timescaledb-event-streamer/internal/eventing/sink/nats"
	_ "github.com/noctarius/timescaledb-event-streamer/internal/eventing/sink/redis"
	_ "github.com/noctarius/timescaledb-event-streamer/internal/eventing/sink/stdout"
)

const publicationName = "pg_ts_streamer"

type Streamer struct {
	replicator *replication.Replicator
}

func NewStreamer(
	config *sysconfig.SystemConfig,
) (*Streamer, *cli.ExitError) {

	if config.PgxConfig == nil {
		connection := spiconfig.GetOrDefault(
			config.Config, spiconfig.PropertyPostgresqlConnection, "host=localhost user=repl_user",
		)

		connConfig, err := pgx.ParseConfig(connection)
		if err != nil {
			return nil, cli.NewExitError(
				fmt.Sprintf("PostgreSQL connection string failed to parse: %s", err.Error()), 6)
		}

		pgPassword := spiconfig.GetOrDefault(config.Config, spiconfig.PropertyPostgresqlPassword, "")
		if pgPassword != "" {
			connConfig.Password = pgPassword
		}

		config.PgxConfig = connConfig
	}

	pgPublication := spiconfig.GetOrDefault(config.Config, spiconfig.PropertyPostgresqlPublicationName, "")
	if pgPublication == "" {
		config.PostgreSQL.Publication.Name = publicationName
	}

	if config.Topic.Prefix == "" {
		config.Topic.Prefix = lo.RandomString(20, lo.LowerCaseLettersCharset)
	}

	// Start all potential plugins, to make sure they're registered
	// before we try to access any of the interface implementations
	if err := plugins.LoadPlugins(config.Config); err != nil {
		return nil, erroring.AdaptError(err, 50)
	}

	replicator, err := replication.NewReplicator(config)
	if err != nil {
		return nil, erroring.AdaptError(err, 21)
	}

	return &Streamer{
		replicator: replicator,
	}, nil
}

func (s *Streamer) Start() *cli.ExitError {
	return s.replicator.StartReplication()
}

func (s *Streamer) Stop() *cli.ExitError {
	return s.replicator.StopReplication()
}
