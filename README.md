# Workflow Migration

Deis (pronounced DAY-iss) Workflow is an open source Platform as a Service (PaaS) that adds a developer-friendly layer to any [Kubernetes](http://kubernetes.io) cluster, making it easy to deploy and manage applications on your own servers.

For more information about the Deis Workflow, please visit the main project page at https://github.com/deis/workflow.

We welcome your input! If you have feedback, please [submit an issue][issues]. If you'd like to participate in development, please read the "Development" section below and [submit a pull request][prs].

# About
The Workflow Migration service is used to migrate from an helm classic install of the workflow to helm without destroying the existing cluster or having any downtime for the apps. It does so by first checking the current install of workflow and creating a release artifact similar to the one helm creates during an install thereby making helm think that the current install is actually created by it. Then workflow can be simply upgraded whenever needed using the helm charts.

# Usage
1) Run the migration service to create a helm release object.
```
$ git clone https://github.com/deis/workflow-migration.git
$ cd workflow-migration
$ helm install ./charts/workflow-migration/ --set release_name=<optional release name for the helm>,workflow_version=<optional current version of workflow>
```
or
```
$ helm repo add deismigration https://charts.deis.com/migration
$ helm install deismigraton/workflow-migration --set release_name=<optional release name for the helm>,workflow_version=<optional current version of workflow>
```

2) Check that the job ran successfully. Name will the release_name provided which default to `deis-workflow` if not provided and chart version will default to `v2.7.0` if workflow_version is not provided during the install.
```
$ helm list
NAME    	     REVISION	  UPDATED                 	STATUS  	CHART          
deis-workflow	  1       	 Tue Nov  1 11:09:54 2016	DEPLOYED	workflow-v2.7.0
```

3) Upgrade to a new workflow release using the new helm. A values file with the details of external configuration used during the install is needed during an upgrade from helm classic to helm.
```
$ helm repo add deis https://charts.deis.com/workflow
$ helm upgrade <release_name> deis/workflow --version=<desired version> -f <path to the values.yaml>
```

# License

Copyright 2013, 2014, 2015, 2016 Engine Yard, Inc.

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at <http://www.apache.org/licenses/LICENSE-2.0>

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.

[issues]: https://github.com/deis/workflow/issues
[prs]: https://github.com/deis/workflow/pulls
