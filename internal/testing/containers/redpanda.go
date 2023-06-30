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

package containers

import (
	"context"
	"fmt"
	"github.com/noctarius/timescaledb-event-streamer/internal/supporting/logging"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func SetupRedPandaContainer() (testcontainers.Container, []string, error) {
	containerRequest := testcontainers.ContainerRequest{
		Image:        "redpandadata/redpanda:v23.1.4",
		ExposedPorts: []string{"9092:9092/tcp"},
		Cmd:          []string{"redpanda", "start"},
		WaitingFor:   wait.ForLog("Initialized cluster_id to"),
	}

	logger, err := logging.NewLogger("testcontainers")
	if err != nil {
		return nil, nil, err
	}
	redpandaLogger, err := logging.NewLogger("testcontainers-redpanda")
	if err != nil {
		return nil, nil, err
	}

	container, err := testcontainers.GenericContainer(
		context.Background(),
		testcontainers.GenericContainerRequest{
			ContainerRequest: containerRequest,
			Started:          true,
			Logger:           logger,
		},
	)
	if err != nil {
		return nil, nil, err
	}

	// Collect logs
	container.FollowOutput(newLogConsumer(redpandaLogger))
	container.StartLogProducer(context.Background())

	host, err := container.Host(context.Background())
	if err != nil {
		container.Terminate(context.Background())
		return nil, nil, err
	}

	port, err := container.MappedPort(context.Background(), "9092/tcp")
	if err != nil {
		container.Terminate(context.Background())
		return nil, nil, err
	}

	return container, []string{fmt.Sprintf("%s:%d", host, port.Int())}, nil
}
