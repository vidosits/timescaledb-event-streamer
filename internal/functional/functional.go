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

package functional

import "sort"

func Zero[T any]() (t T) {
	return
}

func MappingTransformer[T, V any](
	transformer func(T) V,
) func(T, int) V {

	return func(value T, _ int) V {
		return transformer(value)
	}
}

func Sort[I any](
	collection []I, less func(this, other I) bool,
) []I {

	sort.Slice(collection, func(i, j int) bool {
		return less(collection[i], collection[j])
	})
	return collection
}

func ArrayEqual[T comparable](
	this, that []T,
) bool {

	if (this == nil && that != nil) || (this != nil && that == nil) {
		return false
	}
	if len(this) != len(that) {
		return false
	}
	for i := 0; i < len(this); i++ {
		if this[i] != that[i] {
			return false
		}
	}
	return true
}
