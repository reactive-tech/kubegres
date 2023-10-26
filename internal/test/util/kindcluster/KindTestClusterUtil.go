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

package kindcluster

import (
	"bytes"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const clusterName = "kubegres"

type KindTestClusterUtil struct {
	kindExecPath string
}

func (r *KindTestClusterUtil) StartCluster() {

	log.Println("STARTING KIND CLUSTER '" + clusterName + "'")

	// start kind cluster
	var err error
	r.kindExecPath, err = exec.LookPath("kind")
	if err != nil {
		log.Fatal("We cannot find the executable 'kind'. " +
			"Make sure 'kind' is installed and the executable 'kind' " +
			"is in the classpath before running the tests.")
	}

	if r.isClusterRunning() {
		log.Println("Cluster is already running. No need to restart it.")
		r.installOperator()
		return
	}

	clusterConfigFilePath := r.getClusterConfigFilePath()

	log.Println("Running 'kind create cluster --name " + clusterName + " --config " + clusterConfigFilePath + "'")

	var out bytes.Buffer
	cmdStartCluster := &exec.Cmd{
		Path:   r.kindExecPath,
		Args:   []string{r.kindExecPath, "create", "cluster", "--name", clusterName, "--config", clusterConfigFilePath},
		Stdout: &out,
		Stderr: os.Stdout,
	}

	err = cmdStartCluster.Run()
	if err != nil {
		log.Fatal("Unable to execute the command 'kind create cluster --name "+clusterName+" --config "+clusterConfigFilePath+"'", err)
	} else {
		log.Println("CLUSTER STARTED")
		r.installOperator()
	}
}

func (r *KindTestClusterUtil) DeleteCluster() bool {

	log.Println("DELETING KIND CLUSTER '" + clusterName + "'")

	log.Println("Running 'kind delete cluster --name " + clusterName + "'")

	var out bytes.Buffer
	cmdDeleteCluster := &exec.Cmd{
		Path:   r.kindExecPath,
		Args:   []string{r.kindExecPath, "delete", "cluster", "--name", clusterName},
		Stdout: &out,
		Stderr: os.Stdout,
	}

	err := cmdDeleteCluster.Run()
	if err != nil {
		log.Fatal("Unable to execute the command 'kind delete cluster --name "+clusterName+"'", err)
		return false
	}

	return true
}

func (r *KindTestClusterUtil) isClusterRunning() bool {

	var out bytes.Buffer
	cmdGetClusters := &exec.Cmd{
		Path:   r.kindExecPath,
		Args:   []string{r.kindExecPath, "get", "clusters"},
		Stdout: &out,
		Stderr: os.Stdout,
	}

	err := cmdGetClusters.Run()
	if err != nil {
		log.Fatal("Unable to execute the command 'kind get clusters'", err)
	}

	outputLines := strings.Split(out.String(), "\n")
	for _, cluster := range outputLines {
		if cluster == clusterName {
			return true
		}
		log.Println("cluster: '" + cluster + "'")
	}
	return false
}

func (r *KindTestClusterUtil) installOperator() bool {

	log.Println("Installing Kubegres operator in Cluster")

	makeFilePath := r.getMakeFilePath()
	makeFileFolder := r.getMakeFileFolder(makeFilePath)
	log.Println("Running 'make install -f " + makeFilePath + " -C " + makeFileFolder + "'")

	// -C /home/alex/source/kubegres

	makeExecPath, err := exec.LookPath("make")
	if err != nil {
		log.Fatal("We cannot find the executable 'make'. " +
			"Make sure 'make' is installed and the executable 'make' " +
			"is in the classpath before running the tests.")
	}

	var out bytes.Buffer
	cmdMakeInstallClusters := &exec.Cmd{
		Path:   makeExecPath,
		Args:   []string{makeExecPath, "install", "-f", makeFilePath, "-C", makeFileFolder},
		Stdout: &out,
		Stderr: os.Stdout,
	}

	err = cmdMakeInstallClusters.Run()
	if err != nil {
		log.Fatal("Unable to execute the command 'make install -f "+makeFilePath+" -C "+makeFileFolder+"'", err)
		return false
	}

	return true
}

func (r *KindTestClusterUtil) getClusterConfigFilePath() string {
	fileAbsolutePath, err := filepath.Abs("util/kindcluster/kind-cluster-config.yaml")
	if err != nil {
		log.Fatal("Error while trying to get the absolute-path of 'kind-cluster-config.yaml': ", err)
	}
	return fileAbsolutePath
}

func (r *KindTestClusterUtil) getMakeFilePath() string {
	fileAbsolutePath, err := filepath.Abs("../../Makefile")
	if err != nil {
		log.Fatal("Error while trying to get the absolute-path of 'MakeFile': ", err)
	}
	return fileAbsolutePath
}

func (r *KindTestClusterUtil) getMakeFileFolder(makeFilePath string) string {
	return filepath.Dir(makeFilePath)
}
