/*
Copyright 2023 Reactive Tech Limited.
"Reactive Tech Limited" is a company located in England, United Kingdom.
https://www.reactive-tech.io

Lead Developer: Alex Arica

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package log

import (
	"fmt"
	"reflect"
)

func InterfacesToStr(keysAndValues ...interface{}) string {

	if len(keysAndValues) == 0 || (len(keysAndValues) == 1 && reflect.ValueOf(keysAndValues[0]).IsNil()) {
		return ""
	}
	return keyValuesPairsToStr(keysAndValues...)
}

func keyValuesPairsToStr(keysAndValues ...interface{}) string {
	keysAndValuesStr := ""
	comma := ""
	for _, keyAndValue := range keysAndValues {
		keysAndValuesStr += comma + interfaceToStr(keyAndValue)
		comma = ", "
	}
	return keysAndValuesStr
}

func interfaceToStr(keyAndValue interface{}) string {

	keysAndValuesSlice := reflect.ValueOf(keyAndValue)
	if keysAndValuesSlice.Kind() != reflect.Slice {
		return ""
	}

	keysAndValuesStr := ""
	comma := ""
	keysAndValuesLength := keysAndValuesSlice.Len()

	for i := 0; i < keysAndValuesLength; i += 2 {
		key := keysAndValuesSlice.Index(i).Interface()

		var value interface{}
		if (i + 1) < keysAndValuesLength {
			value = keysAndValuesSlice.Index(i + 1).Interface()
		}

		keysAndValuesStr += comma + fmt.Sprintf("'%s'", key) + ": " + fmt.Sprintf("%v", value)
		comma = ", "
	}

	return keysAndValuesStr
}
