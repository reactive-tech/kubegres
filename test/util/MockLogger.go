/*
Copyright 2021 Reactive Tech Limited.
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
	log2 "reactive-tech.io/kubegres/controllers/ctx/log"
	"strings"
)

type MockLogger struct {
	InfoLogger logr.InfoLogger
	Name       string
}

func (r *MockLogger) Info(msg string, keysAndValues ...interface{}) {
	log.Println(r.constructFullMsg(msg, keysAndValues))
}

func (r *MockLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	log.Println(r.constructFullErrMsg(err, msg, keysAndValues))
}

func (r *MockLogger) Enabled() bool {
	return true
}

func (r *MockLogger) V(level int) logr.InfoLogger {
	return r
}

func (r *MockLogger) WithValues(keysAndValues ...interface{}) logr.Logger {
	return r
}

func (r *MockLogger) WithName(name string) logr.Logger {
	if !strings.Contains(r.Name, name) {
		r.Name += name + " - "
	}
	return r
}

func (r *MockLogger) constructFullErrMsg(err error, msg string, keysAndValues ...interface{}) string {
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

func (r *MockLogger) constructFullMsg(msg string, keysAndValues ...interface{}) string {
	if msg == "" {
		return r.Name + ""
	}

	keysAndValuesStr := log2.InterfacesToStr(keysAndValues...)
	if keysAndValuesStr != "" {
		return r.Name + msg + " " + keysAndValuesStr
	}
	return r.Name + msg
}
