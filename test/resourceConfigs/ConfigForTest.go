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

package resourceConfigs

import "time"

const (
	TestTimeout            = time.Second * 240
	TestRetryInterval      = time.Second * 5
	DefaultNamespace       = "default"
	PrimaryReplicationRole = "primary"
	DbPort                 = 5432
	DbUser                 = "postgres"
	DbPassword             = "postgresSuperUserPsw"
	DbName                 = "postgres"
	TableName              = "account"
	DbHost                 = "172.18.0.3"

	KubegresYamlFile     = "resourceConfigs/kubegres.yaml"
	KubegresResourceName = "my-kubegres"

	SecretYamlFile     = "resourceConfigs/secret.yaml"
	SecretResourceName = "my-kubegres-secret"

	ServiceToSqlQueryPrimaryDbYamlFile     = "resourceConfigs/primaryService.yaml"
	ServiceToSqlQueryPrimaryDbResourceName = "test-kubegres-primary"
	ServiceToSqlQueryPrimaryDbNodePort     = 30007

	ServiceToSqlQueryReplicaDbServiceYamlFile     = "resourceConfigs/replicaService.yaml"
	ServiceToSqlQueryReplicaDbServiceResourceName = "test-kubegres-replica"
	ServiceToSqlQueryReplicaDbNodePort            = 30008

	BackUpPvcResourceName  = "test-pvc-for-backup"
	BackUpPvcResourceName2 = "test-pvc-for-backup-2"
	BackUpPvcYamlFile      = "resourceConfigs/backupPvc.yaml"

	CustomConfigMapEmptyResourceName = "config-empty"
	CustomConfigMapEmptyYamlFile     = "resourceConfigs/customConfig/configMap_empty.yaml"

	CustomConfigMapWithAllConfigsResourceName = "config-with-all-configs"
	CustomConfigMapWithAllConfigsYamlFile     = "resourceConfigs/customConfig/configMap_with_all_configs.yaml"

	CustomConfigMapWithBackupDatabaseScriptResourceName = "config-with-backup-database-script"
	CustomConfigMapWithBackupDatabaseScriptYamlFile     = "resourceConfigs/customConfig/configMap_with_backup_database_script.yaml"

	CustomConfigMapWithPgHbaConfResourceName = "config-with-pg-hba-conf"
	CustomConfigMapWithPgHbaConfYamlFile     = "resourceConfigs/customConfig/configMap_with_pg_hba_conf.yaml"

	CustomConfigMapWithPostgresConfResourceName = "config-with-postgres-conf"
	CustomConfigMapWithPostgresConfYamlFile     = "resourceConfigs/customConfig/configMap_with_postgres_conf.yaml"

	CustomConfigMapWithPrimaryInitScriptResourceName = "config-with-primary-init-script"
	CustomConfigMapWithPrimaryInitScriptYamlFile     = "resourceConfigs/customConfig/configMap_with_primary_init_script.yaml"

	CustomConfigMapWithPostgresConfAndWalLevelSetToLogicalResourceName = "config-with-postgres-conf-wal-level-to-logical"
	CustomConfigMapWithPostgresConfAndWalLevelSetToLogicalYamlFile     = "resourceConfigs/customConfig/configMap_with_postgres_conf_and_wal_level_to_logical.yaml"
)
