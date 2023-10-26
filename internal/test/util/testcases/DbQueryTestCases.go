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

package testcases

import (
	. "github.com/onsi/gomega"
	"log"
	"reactive-tech.io/kubegres/internal/test/resourceConfigs"
	util2 "reactive-tech.io/kubegres/internal/test/util"
)

type DbQueryTestCases struct {
	connectionPrimaryDb util2.DbConnectionDbUtil
	connectionReplicaDb util2.DbConnectionDbUtil
}

func InitDbQueryTestCases(resourceCreator util2.TestResourceCreator, kubegresName string) DbQueryTestCases {
	return InitDbQueryTestCasesWithNodePorts(resourceCreator, kubegresName, resourceConfigs.ServiceToSqlQueryPrimaryDbNodePort, resourceConfigs.ServiceToSqlQueryReplicaDbNodePort)
}

func InitDbQueryTestCasesWithNodePorts(resourceCreator util2.TestResourceCreator, kubegresName string, primaryServiceNodePort, replicaServiceNodePort int) DbQueryTestCases {
	return DbQueryTestCases{
		connectionPrimaryDb: util2.InitDbConnectionDbUtil(resourceCreator, kubegresName, primaryServiceNodePort, true),
		connectionReplicaDb: util2.InitDbConnectionDbUtil(resourceCreator, kubegresName, replicaServiceNodePort, false),
	}
}

func (r *DbQueryTestCases) ThenWeCanSqlQueryPrimaryDb() {

	Eventually(func() bool {

		isInserted := r.connectionPrimaryDb.InsertUser()
		if !isInserted {
			return false
		}

		users := r.connectionPrimaryDb.GetUsers()
		r.connectionPrimaryDb.Close()

		if r.connectionPrimaryDb.LastInsertedUserId != "" {
			if !r.isLastInsertedUserInDb(users) {
				log.Println("Does not contain the last inserted userId: '" + r.connectionPrimaryDb.LastInsertedUserId + "'")
				return false
			}
		}

		return len(users) == r.connectionPrimaryDb.NbreInsertedUsers

	}, resourceConfigs.TestTimeout, resourceConfigs.TestRetryInterval).Should(BeTrue())
}

func (r *DbQueryTestCases) ThenWeCanSqlQueryReplicaDb() {
	Eventually(func() bool {
		users := r.connectionReplicaDb.GetUsers()
		r.connectionReplicaDb.Close()
		return len(users) == r.connectionPrimaryDb.NbreInsertedUsers

	}, resourceConfigs.TestTimeout, resourceConfigs.TestRetryInterval).Should(BeTrue())
}

func (r *DbQueryTestCases) isLastInsertedUserInDb(users []util2.AccountUser) bool {
	for _, user := range users {
		if user.UserId == r.connectionPrimaryDb.LastInsertedUserId {
			return true
		}
	}
	return false
}
