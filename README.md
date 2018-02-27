
|![](https://upload.wikimedia.org/wikipedia/commons/thumb/1/17/Warning.svg/156px-Warning.svg.png) | Deis Workflow is no longer maintained.<br />Please [read the announcement](https://deis.com/blog/2017/deis-workflow-final-release/) for more detail. |
|---:|---|
| 09/07/2017 | Deis Workflow [v2.18][] final release before entering maintenance mode |
| 03/01/2018 | End of Workflow maintenance: critical patches no longer merged |
| | [Hephy](https://github.com/teamhephy/workflow) is a fork of Workflow that is actively developed and accepts code contributions. |

# Workflow Migration

Deis (pronounced DAY-iss) Workflow is an open source Platform as a Service (PaaS) that adds a developer-friendly layer to any [Kubernetes](http://kubernetes.io) cluster, making it easy to deploy and manage applications on your own servers.

For more information about the Deis Workflow, please visit the main project page at https://github.com/deis/workflow.

We welcome your input! If you have feedback, please [submit an issue][issues]. If you'd like to participate in development, please read the "Development" section below and [submit a pull request][prs].

# About
The Workflow Migration service is used to migrate from a [helm-classic](https://github.com/helm/helm-classic) install of Workflow to [Kubernetes Helm](https://github.com/kubernetes/helm) without destroying the existing cluster or having any downtime for the apps. It does so by first checking the current install of Workflow and creating a release artifact similar to the one Kubernetes helm creates during an install thereby making Kubernetes Helm think that the current install is actually created by it. Then Workflow can be simply upgraded whenever needed using the Kubernetes Helm charts.

> Warning: Only workflow install on or after v2.6.0 can be upgraded using this migration service.

# Usage
1) Check that kubernetes helm and its corresponding server component tiller are [installed](https://github.com/kubernetes/helm/blob/master/docs/install.md). Be sure that the helm version is `v2.1.3` or later because earlier versions have issues that may prevent upgrading successfully.

```shell
$ helm version
Client: &version.Version{SemVer:"v2.1.3", GitCommit:"5cbc48fb305ca4bf68c26eb8d2a7eb363227e973", GitTreeState:"clean"}
Server: &version.Version{SemVer:"v2.1.3", GitCommit:"5cbc48fb305ca4bf68c26eb8d2a7eb363227e973", GitTreeState:"clean"}
```

2) Fetch the registry and controller deployment objects just to make sure that the existing install state can achieved if the deis migration service fails. If you are using the off-cluster registry then there won't be any registry deployment and no need to fetch it. Deis migration service deletes the registry and controller deployment objects because of an [issue](https://github.com/kubernetes/kubernetes/pull/35071) in kubernetes with the patching.

```shell
$ kubectl --namespace=deis get deployment deis-registry -o yaml > ~/active-deis-registry-deployment.yaml
$ kubectl --namespace=deis get deployment deis-controller -o yaml > ~/active-deis-controller-deployment.yaml
```

3) Run the migration service to create a helm release object based on the current workflow install. If not otherwise specified, the workflow_release_name will be `deis-workflow` and workflow_version will be `v2.7.0`.

```shell
$ git clone https://github.com/deis/workflow-migration.git
$ cd workflow-migration
$ helm install ./charts/workflow-migration/ --set workflow_release_name=<optional release name for the helm>,workflow_version=<optional current version of workflow>
```

or

```shell
$ helm repo add workflow-migration https://charts.deis.com/workflow-migration
$ helm install workflow-migration/workflow-migration --set workflow_release_name=<optional release name for the helm>,workflow_version=<optional current version of workflow>
```

4) Check that the job ran successfully. Also check that helm release is created for the current workflow install using `helm list` where Name will be the workflow_release_name and chart version will be the workflow_version.

```shell
$ kubectl get jobs
NAME                 DESIRED   SUCCESSFUL   AGE
workflow-migration   1         1            48s

$ helm list
NAME    	     REVISION	  UPDATED                 	 STATUS  	  CHART
erstwhile-oran   1            Wed Nov  1 11:09:34 2016   DEPLOYED     workflow-migration-v1.0.0
deis-workflow    1            Tue Nov  1 11:09:54 2016   DEPLOYED     workflow-v2.7.0
```

5) Upgrade to a new workflow release using the kubernetes helm. All the configuration used during install of workflow will be preserved over the update. You can check the configuration before upgrading to the new release.

```shell
$ helm get values <workflow_release_name>  ## will print the configuration values

$ helm repo add deis https://charts.deis.com/workflow
$ helm upgrade <workflow_release_name> deis/workflow --version=<desired version>
```

6) Verify that all components have started and passed their readiness checks:

```shell
$ kubectl --namespace=deis get pods
NAME                                     READY     STATUS    RESTARTS   AGE
deis-builder-2448122224-3cibz            1/1       Running   0          5m
deis-controller-1410285775-ipc34         1/1       Running   3          5m
deis-database-e7c5z                      1/1       Running   0          5m
deis-logger-cgjup                        1/1       Running   3          5m
deis-logger-fluentd-45h7j                1/1       Running   0          5m
deis-logger-fluentd-4z7lw                1/1       Running   0          5m
deis-logger-fluentd-k2wsw                1/1       Running   0          5m
deis-logger-fluentd-skdw4                1/1       Running   0          5m
deis-logger-redis-8nazu                  1/1       Running   0          5m
deis-monitor-grafana-tm266               1/1       Running   0          5m
deis-monitor-influxdb-ah8io              1/1       Running   0          5m
deis-monitor-telegraf-51zel              1/1       Running   1          5m
deis-monitor-telegraf-cdasg              1/1       Running   0          5m
deis-monitor-telegraf-hea6x              1/1       Running   0          5m
deis-monitor-telegraf-r7lsg              1/1       Running   0          5m
deis-nsqd-3yrg2                          1/1       Running   0          5m
deis-registry-1814324048-yomz5           1/1       Running   0          5m
deis-registry-proxy-4m3o4                1/1       Running   0          5m
deis-registry-proxy-no3r1                1/1       Running   0          5m
deis-registry-proxy-ou8is                1/1       Running   0          5m
deis-registry-proxy-zyajl                1/1       Running   0          5m
deis-router-1357759721-a3ard             1/1       Running   0          5m
deis-workflow-manager-2654760652-kitf9   1/1       Running   0          5m
```

[issues]: https://github.com/deis/workflow/issues
[prs]: https://github.com/deis/workflow/pulls
[v2.18]: https://github.com/deis/workflow/releases/tag/v2.18.0
