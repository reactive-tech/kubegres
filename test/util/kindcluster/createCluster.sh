#!/bin/bash

clusterName="kubegres"

kind get clusters | grep clusterName && doesClusterExist=1 || doesClusterExist=0

if [ doesClusterExist == 0 ]; then

  kind create cluster --name clusterName --config "+clusterConfigFilePath+"'

fi


