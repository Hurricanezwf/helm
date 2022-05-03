/*
 Copyright 2022 Hurricanezwf

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

      https://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package helm

import "log"

type (
	DebugLogFunc LogFunc
	WarnLogFunc  LogFunc
	LogFunc      func(format string, args ...interface{})
)

func DefaultDebugLogFunc() DebugLogFunc {
	return func(format string, args ...interface{}) {
		log.Printf("[DEBUG] "+format, args...)
	}
}

func DefaultWarnLogFunc() WarnLogFunc {
	return func(format string, args ...interface{}) {
		log.Printf("[WARN] "+format, args...)
	}
}

func DefaultDiscardLogFunc() LogFunc {
	return func(format string, args ...interface{}) {}
}
