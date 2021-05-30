package util

import (
	v12 "k8s.io/api/core/v1"
	postgresv1 "reactive-tech.io/kubegres/api/v1"
)

type TestResourceModifier struct {
}

func (r *TestResourceModifier) AppendAnnotation(annotationKey, annotationValue string, kubegresResource *postgresv1.Kubegres) {

	if kubegresResource.Annotations == nil {
		kubegresResource.Annotations = make(map[string]string)
	}

	kubegresResource.Annotations[annotationKey] = annotationValue
}

func (r *TestResourceModifier) AppendEnvVar(envVarName, envVarVal string, kubegresResource *postgresv1.Kubegres) {

	envVar := v12.EnvVar{
		Name:  envVarName,
		Value: envVarVal,
	}

	kubegresResource.Spec.Env = append(kubegresResource.Spec.Env, envVar)
}

func (r *TestResourceModifier) AppendEnvVarFromSecretKey(envVarName, envVarSecretKeyValueName string, kubegresResource *postgresv1.Kubegres) {

	envVar := v12.EnvVar{
		Name: envVarName,
		ValueFrom: &v12.EnvVarSource{
			SecretKeyRef: &v12.SecretKeySelector{
				Key: envVarSecretKeyValueName,
				LocalObjectReference: v12.LocalObjectReference{
					Name: "my-kubegres-secret",
				},
			},
		},
	}

	kubegresResource.Spec.Env = append(kubegresResource.Spec.Env, envVar)
}
