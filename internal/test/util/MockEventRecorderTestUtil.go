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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"log"
	log2 "reactive-tech.io/kubegres/internal/controller/ctx/log"
)

type MockEventRecorderTestUtil struct {
	eventRecords []EventRecord
}

type EventRecord struct {
	Eventtype string
	Reason    string
	Message   string
}

func (r *MockEventRecorderTestUtil) CheckEventExist(eventRecordToSearch EventRecord) bool {

	for _, eventRecord := range r.eventRecords {
		if eventRecord == eventRecordToSearch {
			return true
		}
	}
	return false
}

func (r *MockEventRecorderTestUtil) RemoveAllEvents() {
	r.eventRecords = []EventRecord{}
}

func (r *MockEventRecorderTestUtil) Event(object runtime.Object, eventtype, reason, message string) {
	r.Eventf(object, eventtype, reason, message)
}

func (r *MockEventRecorderTestUtil) Eventf(object runtime.Object, eventtype, reason, messageFmt string, args ...interface{}) {
	messageFmt = r.constructFullMsg(messageFmt, args)
	log.Println("Event - eventtype: '" + eventtype + "', reason: '" + reason + "', message: '" + messageFmt + "'")
	r.eventRecords = append(r.eventRecords, EventRecord{Eventtype: eventtype, Reason: reason, Message: messageFmt})
}

func (r *MockEventRecorderTestUtil) PastEventf(object runtime.Object, timestamp metav1.Time, eventtype, reason, messageFmt string, args ...interface{}) {
	r.Eventf(object, eventtype, reason, messageFmt, args)
}

func (r *MockEventRecorderTestUtil) AnnotatedEventf(object runtime.Object, annotations map[string]string, eventtype, reason, messageFmt string, args ...interface{}) {
	r.Eventf(object, eventtype, reason, messageFmt, args)
}

func (r *MockEventRecorderTestUtil) constructFullMsg(msg string, keysAndValues ...interface{}) string {
	keysAndValuesStr := log2.InterfacesToStr(keysAndValues...)
	if keysAndValuesStr != "" {
		return msg + " " + keysAndValuesStr
	}
	return msg
}
