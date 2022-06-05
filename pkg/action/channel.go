// Copyright 2022 Wenfeng Zhou (zwf1094646850@gmail.com)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package action

import "sync/atomic"

type ResultMessageChan struct {
	refCount int32
	ch       chan resultMessage
}

func NewResultMessageChan(refCount int) *ResultMessageChan {
	if refCount < 0 {
		panic("ref count must be >= 0")
	}
	return &ResultMessageChan{
		refCount: int32(refCount),
		ch:       make(chan resultMessage, 1),
	}
}

func (c *ResultMessageChan) Close() {
	if atomic.AddInt32(&c.refCount, -1) == 0 {
		close(c.ch)
	}
}

func (c *ResultMessageChan) Channel() chan resultMessage {
	return c.ch
}
