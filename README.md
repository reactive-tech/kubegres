News: A webinar about Kubegres will be organised on 25 May 2021 hosted by PostgresConf. [Register here!](https://postgresconf.org/conferences/2021_Postgres_Conference_Webinars/program/proposals/creating-a-resilient-postgresql-cluster-with-kubegres)

[Kubegres](https://www.kubegres.io/) is a Kubernetes operator allowing to deploy a cluster of PostgreSql pods with data 
replication enabled out-of-the box. It brings simplicity when using PostgreSql considering how complex managing 
stateful-set's life-cycle and data replication could be with Kubernetes.

**Features**

* It creates a cluster of PostgreSql servers with data replication enabled: it creates a Primary PostgreSql pod and a 
  number of Replica PostgreSql pods and replicates primary's database in real-time to Replica pods.

* It manages fail-over: if a Primary PostgreSql crashes, it automatically promotes a Replica PostgreSql as a Primary.

* It has a data backup option allowing to dump PostgreSql data regularly in a given volume.

* It provides a very simple YAML with properties specialised for PostgreSql.

* It is resilient, has over [55 automatized tests](https://github.com/reactive-tech/kubegres/tree/main/test) cases and 
  has been running in production. 

[Please click here to get started](https://www.kubegres.io/doc/getting-started.html).

**Getting started**

If you would like to install Kubegres, please read the page [Getting started](http://www.kubegres.io/doc/getting-started.html).

**Contribute**

If you would like to contribute to Kubegres, please read the page [How to contribute](http://www.kubegres.io/contribute/).

**More details about the project**

[Kubegres](https://www.kubegres.io/) was developed by [Reactive Tech Limited](https://www.reactive-tech.io/)  and Alex 
Arica as the lead developer. Reactive Tech offers [support services](https://www.kubegres.io/support/) for Kubegres, 
Kubernetes and PostgreSql.

It was developed with the framework [Kubebuilder](https://book.kubebuilder.io/) version 3, an SDK for building Kubernetes 
APIs using CRDs. Kubebuilder is maintained by the official Kubernetes API Machinery Special Interest Group (SIG).
