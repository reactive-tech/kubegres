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

package util

import (
	"github.com/go-logr/logr"
	"log"
	log2 "reactive-tech.io/kubegres/internal/controller/ctx/log"
	"strings"
)

type MockLogSink struct {
	name string
}

func CreateMockLogger() logr.Logger {
	logSink := &MockLogSink{}
	logger := logr.New(logSink)
	return logger
}

func (r *MockLogSink) Init(info logr.RuntimeInfo) {
}

func (r *MockLogSink) Enabled(level int) bool {
	return true
}

func (r *MockLogSink) Info(level int, msg string, keysAndValues ...interface{}) {
	log.Println(r.constructFullMsg(msg, keysAndValues))
}

func (r *MockLogSink) Error(err error, msg string, keysAndValues ...interface{}) {
	log.Println(r.constructFullErrMsg(err, msg, keysAndValues))
}

func (r *MockLogSink) WithValues(keysAndValues ...interface{}) logr.LogSink {
	return r
}

func (r *MockLogSink) WithName(name string) logr.LogSink {
	if !strings.Contains(r.name, name) {
		r.name += name + " - "
	}
	return r
}

func (r *MockLogSink) constructFullErrMsg(err error, msg string, keysAndValues ...interface{}) string {
	msgToReturn := ""
	separator := ""

	customErrMsg := r.constructFullMsg(msg, keysAndValues...)
	if customErrMsg != "" {
		msgToReturn = customErrMsg
		separator = " - "
	}

	msgFromErr := err.Error()
	if msgFromErr != "" {
		msgToReturn += separator + msgFromErr
	}

	return msgToReturn
}

func (r *MockLogSink) constructFullMsg(msg string, keysAndValues ...interface{}) string {
	if msg == "" {
		return r.name + ""
	}

	keysAndValuesStr := log2.InterfacesToStr(keysAndValues...)
	if keysAndValuesStr != "" {
		return r.name + msg + " " + keysAndValuesStr
	}
	return r.name + msg
}
