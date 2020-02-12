/*
   Copyright 2020 T-Systems GmbH

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
package services

import (
	"math/rand"
	"time"
	"unsafe"
)

// Random generation for tests

// DataRandCS is charset for randomizing
const DataRandCS = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ-"

var src = rand.NewSource(time.Now().UnixNano())

func randomByteSlice(size int, prefix string, charset string) []byte {
	csLen := len(charset)
	prefLen := len(prefix)
	result := make([]byte, size)
	copy(result, prefix)
	for i := prefLen; i < size; i++ {
		result[i] = charset[src.Int63()%int64(csLen)]
	}
	return result
}

// RandomString generates random string
func RandomString(size int, prefix string, charset ...string) string {
	cs := DataRandCS
	if len(charset) > 0 {
		cs = charset[0]
	}
	result := randomByteSlice(size, prefix, cs)
	return *(*string)(unsafe.Pointer(&result)) // faster way to convert big slice to string
}
