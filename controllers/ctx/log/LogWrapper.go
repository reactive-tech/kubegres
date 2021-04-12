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

package log

import (
	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
	postgresV1 "reactive-tech.io/kubegres/api/v1"
)

type LogWrapper struct {
	Kubegres *postgresV1.Kubegres
	Logger   logr.Logger
	Recorder record.EventRecorder
}

func (r *LogWrapper) WithValues(keysAndValues ...interface{}) {
	r.Logger = r.Logger.WithValues(keysAndValues...)
}

func (r *LogWrapper) WithName(name string) {
	r.Logger = r.Logger.WithName(name)
}

func (r *LogWrapper) Info(msg string, keysAndValues ...interface{}) {
	r.Logger.Info(msg, keysAndValues...)
}

func (r *LogWrapper) Error(err error, msg string, keysAndValues ...interface{}) {
	r.Logger.Error(err, msg, keysAndValues...)
}

func (r *LogWrapper) InfoEvent(eventReason string, msg string, keysAndValues ...interface{}) {
	r.Info(msg, keysAndValues...)
	r.Recorder.Eventf(r.Kubegres, v1.EventTypeNormal, eventReason, r.constructFullMsg(msg, keysAndValues))
}

func (r *LogWrapper) ErrorEvent(eventReason string, err error, msg string, keysAndValues ...interface{}) {
	r.Error(err, msg, keysAndValues...)
	r.Recorder.Eventf(r.Kubegres, v1.EventTypeWarning, eventReason, r.constructFullErrMsg(err, msg, keysAndValues))
}

func (r *LogWrapper) constructFullMsg(msg string, keysAndValues ...interface{}) string {
	if msg == "" {
		return ""
	}

	keysAndValuesStr := InterfacesToStr(keysAndValues...)
	if keysAndValuesStr != "" {
		return msg + " " + keysAndValuesStr
	}
	return msg
}

func (r *LogWrapper) constructFullErrMsg(err error, msg string, keysAndValues ...interface{}) string {
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
