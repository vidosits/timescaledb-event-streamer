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

package statestorage

import (
	"encoding/binary"
	"github.com/noctarius/timescaledb-event-streamer/spi/pgtypes"
	"github.com/samber/lo"
	"time"
)

type Offset struct {
	Timestamp      time.Time   `json:"timestamp"`
	Snapshot       bool        `json:"snapshot"`
	SnapshotName   *string     `json:"snapshot_name,omitempty"`
	SnapshotOffset int         `json:"snapshot_offset"`
	LSN            pgtypes.LSN `json:"lsn"`
}

func (o *Offset) UnmarshalBinary(
	data []byte,
) error {

	o.Timestamp = time.Unix(0, int64(binary.BigEndian.Uint64(data[:8]))).In(time.UTC)
	o.Snapshot = data[8] == 1
	o.SnapshotOffset = int(binary.BigEndian.Uint32(data[9:]))
	o.LSN = pgtypes.LSN(binary.BigEndian.Uint64(data[13:]))
	if o.Snapshot {
		snapshotNameLength := int(data[21])
		if snapshotNameLength > 0 {
			o.SnapshotName = lo.ToPtr(string(data[22 : 22+snapshotNameLength]))
		}
	}
	return nil
}

func (o *Offset) MarshalBinary() ([]byte, error) {
	size := 21
	if o.SnapshotName != nil {
		size++
		size += len([]byte(*o.SnapshotName))
	}
	data := make([]byte, size)
	binary.BigEndian.PutUint64(data[:8], uint64(o.Timestamp.UnixNano()))
	data[8] = 0
	if o.Snapshot {
		data[8] = 1
	}
	binary.BigEndian.PutUint32(data[9:], uint32(o.SnapshotOffset))
	binary.BigEndian.PutUint64(data[13:], uint64(o.LSN))
	if o.SnapshotName != nil {
		snapshotName := []byte(*o.SnapshotName)
		snapshotNameLength := len(snapshotName)
		data[21] = byte(snapshotNameLength)
		copy(data[22:], snapshotName)
	}
	return data, nil
}

func (o *Offset) Equal(
	other *Offset,
) bool {

	if !o.Timestamp.Equal(other.Timestamp) {
		return false
	}

	if o.Snapshot != other.Snapshot {
		return false
	}

	if o.SnapshotOffset != other.SnapshotOffset {
		return false
	}

	if (o.SnapshotName == nil && other.SnapshotName != nil) ||
		(o.SnapshotName != nil && other.SnapshotName == nil) ||
		(o.SnapshotName != nil && other.SnapshotName != nil && *o.SnapshotName != *other.SnapshotName) {
		return false
	}

	return o.LSN == other.LSN
}
