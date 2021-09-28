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

package states

import (
	core "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"reactive-tech.io/kubegres/controllers/ctx"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	ConfigMapDataKeyPostgresConf      = "postgres.conf"
	ConfigMapDataKeyPrimaryInitScript = "primary_init_script.sh"
	ConfigMapDataKeyPgHbaConf         = "pg_hba.conf"
	ConfigMapDataKeyBackUpScript      = "backup_database.sh"
)

type ConfigStates struct {
	IsBaseConfigDeployed   bool
	BaseConfigName         string
	IsCustomConfigDeployed bool
	CustomConfigName       string
	ConfigLocations        ConfigLocations

	kubegresContext ctx.KubegresContext
}

// Stores as string the volume-name for each config-type which can be either 'base-config' or 'custom-config'
type ConfigLocations struct {
	PostgreConf       string
	PrimaryInitScript string
	BackUpScript      string
	PgHbaConf         string
}

func loadConfigStates(kubegresContext ctx.KubegresContext) (ConfigStates, error) {

	configMapStates := ConfigStates{kubegresContext: kubegresContext}
	configMapStates.BaseConfigName = ctx.BaseConfigMapName
	configMapStates.CustomConfigName = kubegresContext.Kubegres.Spec.CustomConfig

	err := configMapStates.loadStates()

	return configMapStates, err
}

func (r *ConfigStates) loadStates() (err error) {

	r.ConfigLocations.PostgreConf = ctx.BaseConfigMapVolumeName
	r.ConfigLocations.PrimaryInitScript = ctx.BaseConfigMapVolumeName
	r.ConfigLocations.BackUpScript = ctx.BaseConfigMapVolumeName
	r.ConfigLocations.PgHbaConf = ctx.BaseConfigMapVolumeName

	baseConfigMap, err := r.getBaseDeployedConfigMap()
	if err != nil {
		return err
	}

	if r.isBaseConfigMap(baseConfigMap) {
		r.IsBaseConfigDeployed = true
	}

	if r.isBaseConfigAlsoCustomConfig() {
		return nil
	}

	customConfigMap, err := r.getDeployedCustomCustomConfigMap()
	if err != nil {
		return err
	}

	if r.isCustomConfigDeployed(customConfigMap) {

		r.IsCustomConfigDeployed = true

		if customConfigMap.Data[ConfigMapDataKeyPostgresConf] != "" {
			r.ConfigLocations.PostgreConf = ctx.CustomConfigMapVolumeName
		}

		if customConfigMap.Data[ConfigMapDataKeyPrimaryInitScript] != "" {
			r.ConfigLocations.PrimaryInitScript = ctx.CustomConfigMapVolumeName
		}

		if customConfigMap.Data[ConfigMapDataKeyBackUpScript] != "" {
			r.ConfigLocations.BackUpScript = ctx.CustomConfigMapVolumeName
		}

		if customConfigMap.Data[ConfigMapDataKeyPgHbaConf] != "" {
			r.ConfigLocations.PgHbaConf = ctx.CustomConfigMapVolumeName
		}
	}

	return nil
}

func (r *ConfigStates) isBaseConfigAlsoCustomConfig() bool {
	return r.CustomConfigName == r.BaseConfigName
}

func (r *ConfigStates) isBaseConfigMap(configMap *core.ConfigMap) bool {
	return configMap.Name == r.BaseConfigName
}

func (r *ConfigStates) isCustomConfigDeployed(configMap *core.ConfigMap) bool {
	return configMap.Name != "" && configMap.Name != r.BaseConfigName
}

func (r *ConfigStates) getBaseDeployedConfigMap() (*core.ConfigMap, error) {

	namespace := r.kubegresContext.Kubegres.Namespace
	resourceName := r.BaseConfigName
	configMapKey := client.ObjectKey{Namespace: namespace, Name: resourceName}

	return r.getDeployedConfigMap(configMapKey, resourceName, "Base")
}

func (r *ConfigStates) getDeployedCustomCustomConfigMap() (*core.ConfigMap, error) {

	namespace := r.kubegresContext.Kubegres.Namespace
	resourceName := r.CustomConfigName
	configMapKey := client.ObjectKey{Namespace: namespace, Name: resourceName}

	return r.getDeployedConfigMap(configMapKey, resourceName, "Init")
}

func (r *ConfigStates) getDeployedConfigMap(configMapKey client.ObjectKey, resourceName string, logLabel string) (*core.ConfigMap, error) {

	configMap := &core.ConfigMap{}
	err := r.kubegresContext.Client.Get(r.kubegresContext.Ctx, configMapKey, configMap)

	if err != nil {
		if apierrors.IsNotFound(err) {
			err = nil
		} else {
			r.kubegresContext.Log.ErrorEvent("ConfigMapLoadingErr", err, "Unable to load any deployed "+logLabel+" Config.", "Config name", resourceName)
		}
	}

	return configMap, err
}
