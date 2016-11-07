package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/deis/workflow-migration/pkg"
	"github.com/ghodss/yaml"
	"k8s.io/helm/pkg/proto/hapi/chart"
	rspb "k8s.io/helm/pkg/proto/hapi/release"
	"k8s.io/helm/pkg/timeconv"
	"k8s.io/kubernetes/pkg/api"
	apierrors "k8s.io/kubernetes/pkg/api/errors"
	kcl "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/labels"
)

const apiVersion = "v1"

func main() {
	kubeClient, err := kcl.NewInCluster()
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	raw, err := pkg.GetValues(kubeClient)
	if err != nil {
		log.Fatalf("Failed to get values: %v", err)
	}
	fmt.Println(raw)

	secrets := []string{"builder-key-auth", "builder-ssh-private-keys", "database-creds", "django-secret-key", "logger-redis-creds"}

	// Adding the annotation as pre-install hooks will make sure that they don't change
	// during the upgrade from helm classic to helm.
	pkg.UpdateSecrets(kubeClient, secrets)

	// Deployments needs to be deleted because of the issue in kubernetes patching
	// https://github.com/kubernetes/kubernetes/issues/35134.
	err = deleteDeployments(kubeClient)
	if err != nil && !apierrors.IsNotFound(err) {
		log.Fatalf("failed to delete the deployment: %v", err)
	}

	ts := timeconv.Now()
	workflowVersion := getenv("WORKFLOW_VERSION", "v2.7.0")
	config := &chart.Config{Raw: raw}
	chartmetadata := &chart.Metadata{Name: "workflow", Version: workflowVersion}

	// Get the manifest based on the current workflow install which are identfied
	// by the label `heritage: deis`.
	manifestDoc, err := getManifest(kubeClient, secrets)
	if err != nil {
		log.Fatal("get manifest error", err)
	}
	log.Println("generated manifest")
	log.Println(manifestDoc.String())

	releaseName := getenv("RELEASE_NAME", "deis-workflow")
	actualrel := &rspb.Release{
		Name:      releaseName,
		Namespace: "deis",
		Version:   1,
		Config:    config,
		Chart:     &chart.Chart{Metadata: chartmetadata, Values: config},
		Info: &rspb.Info{
			FirstDeployed: ts,
			LastDeployed:  ts,
			Status:        &rspb.Status{Code: rspb.Status_DEPLOYED},
		},
		Manifest: manifestDoc.String(),
	}
	cfgName := fmt.Sprintf("%s.v%d", releaseName, 1)
	err = pkg.CfgCreate(cfgName, actualrel, kubeClient)
	if err != nil {
		log.Fatalf("Failed to create configMap: %v", err)
	}
}

func deleteDeployments(kubeClient *kcl.Client) error {
	deployments := [3]string{"deis-builder", "deis-controller", "deis-registry"}
	for _, deployment := range deployments {
		err := kubeClient.ExtensionsClient.Deployments("deis").Delete(deployment, &api.DeleteOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}

func getManifest(kubeClient *kcl.Client, secretsArray []string) (*bytes.Buffer, error) {
	b := bytes.NewBuffer(nil)
	labelMap := labels.Set{"heritage": "deis"}
	var y []byte

	// ServiceAccounts
	serviceAccounts, err := kubeClient.ServiceAccounts("deis").List(api.ListOptions{LabelSelector: labelMap.AsSelector(), FieldSelector: fields.Everything()})
	if err != nil {
		return nil, err
	}
	for _, serviceAccount := range serviceAccounts.Items {
		serviceAccountNameDet := strings.SplitN(serviceAccount.ObjectMeta.Name, "-", 2)
		path := "workflow/charts/" + serviceAccountNameDet[1] + "templates/" + serviceAccountNameDet[1] + "-service-account.yaml"
		b.WriteString("\n---\n# Source: " + path + "\n")
		serviceAccount.Kind = "ServiceAccount"
		serviceAccount.APIVersion = apiVersion
		serviceAccount.ResourceVersion = ""
		serviceAccount.Secrets = nil
		y, err = yaml.Marshal(serviceAccount)
		if err != nil {
			fmt.Printf("err: %v\n", err)
			return nil, err
		}
		b.WriteString(string(y))
	}

	// Secrets
	secretsMap := make(map[string]struct{})
	for _, secret := range secretsArray {
		secretsMap[secret] = struct{}{}
	}
	secrets, err := kubeClient.Secrets("deis").List(api.ListOptions{LabelSelector: labelMap.AsSelector(), FieldSelector: fields.Everything()})
	if err != nil {
		return nil, err
	}
	for _, secret := range secrets.Items {
		if _, ok := secretsMap[secret.ObjectMeta.GetName()]; ok {
			continue
		}
		secretNameDet := strings.SplitN(secret.ObjectMeta.Name, "-", 2)
		path := "workflow/charts/" + secretNameDet[1] + "templates/" + secretNameDet[1] + "-secret.yaml"
		b.WriteString("\n---\n# Source: " + path + "\n")
		secret.Kind = "Secret"
		secret.APIVersion = apiVersion
		secret.ResourceVersion = ""
		y, err = yaml.Marshal(secret)
		if err != nil {
			fmt.Printf("err: %v\n", err)
			return nil, err
		}
		b.WriteString(string(y))
	}

	// Services
	services, err := kubeClient.Services("deis").List(api.ListOptions{LabelSelector: labelMap.AsSelector(), FieldSelector: fields.Everything()})
	if err != nil {
		return nil, err
	}
	for _, service := range services.Items {
		serviceNameDet := strings.SplitN(service.ObjectMeta.Name, "-", 2)
		path := "workflow/charts/" + serviceNameDet[1] + "templates/" + serviceNameDet[1] + "-service.yaml"
		b.WriteString("\n---\n# Source: " + path + "\n")
		service.Kind = "Service"
		service.APIVersion = apiVersion
		service.ResourceVersion = ""
		service.Spec.ClusterIP = ""
		y, err = yaml.Marshal(service)
		if err != nil {
			fmt.Printf("err: %v\n", err)
			return nil, err
		}
		b.WriteString(string(y))
	}
	// deis-logger-redis service has label `heritage: helm` and hence needs to be manually queried.
	service, err := kubeClient.Services("deis").Get("deis-logger-redis")
	if err != nil && !apierrors.IsNotFound(err) {
		return nil, err
	}
	if err == nil {
		serviceNameDet := strings.SplitN(service.ObjectMeta.Name, "-", 2)
		path := "workflow/charts/" + serviceNameDet[1] + "templates/" + serviceNameDet[1] + "-service.yaml"
		b.WriteString("\n---\n# Source: " + path + "\n")
		service.Kind = "Service"
		service.APIVersion = apiVersion
		service.ResourceVersion = ""
		service.Spec.ClusterIP = ""
		y, err = yaml.Marshal(service)
		if err != nil {
			fmt.Printf("err: %v\n", err)
			return nil, err
		}
		b.WriteString(string(y))
	}

	// Deployments
	deployments, err := kubeClient.Extensions().Deployments("deis").List(api.ListOptions{LabelSelector: labelMap.AsSelector(), FieldSelector: fields.Everything()})
	if err != nil {
		return nil, err
	}
	for _, deployment := range deployments.Items {
		deploymentNameDet := strings.SplitN(deployment.ObjectMeta.Name, "-", 2)
		path := "workflow/charts/" + deploymentNameDet[1] + "templates/" + deploymentNameDet[1] + "-deployment.yaml"
		b.WriteString("\n---\n# Source: " + path + "\n")
		deployment.Kind = "Deployment"
		deployment.APIVersion = "extensions/v1beta1"
		deployment.ResourceVersion = ""
		deployment.ObjectMeta.Annotations = nil
		y, err = yaml.Marshal(deployment)
		if err != nil {
			fmt.Printf("err: %v\n", err)
			return nil, err
		}
		b.WriteString(string(y))
	}

	// DaemonSets
	daemonsets, err := kubeClient.Extensions().DaemonSets("deis").List(api.ListOptions{LabelSelector: labelMap.AsSelector(), FieldSelector: fields.Everything()})
	if err != nil {
		return nil, err
	}
	for _, daemonset := range daemonsets.Items {
		daemonsetNameDet := strings.SplitN(daemonset.ObjectMeta.Name, "-", 2)
		path := "workflow/charts/" + daemonsetNameDet[1] + "templates/" + daemonsetNameDet[1] + "-daemonset.yaml"
		b.WriteString("\n---\n# Source: " + path + "\n")
		daemonset.Kind = "DaemonSet"
		daemonset.APIVersion = "extensions/v1beta1"
		daemonset.ResourceVersion = ""
		y, err = yaml.Marshal(daemonset)
		if err != nil {
			fmt.Printf("err: %v\n", err)
			return nil, err
		}
		b.WriteString(string(y))
	}

	return b, err
}

func getenv(name, dfault string) string {
	value := os.Getenv(name)
	if value == "" {
		value = dfault
	}
	return value
}
