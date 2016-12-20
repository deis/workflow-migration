package pkg

// Most of the code in here is from the https://github.com/kubernetes/kubernetes/blob/master/pkg/kubectl/cmd/annotate.go and
// changed appropriately for the use-case.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ghodss/yaml"
	"k8s.io/client-go/1.5/kubernetes"
	apierrors "k8s.io/client-go/1.5/pkg/api/errors"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/meta"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"k8s.io/kubernetes/pkg/kubectl/resource"
	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/util/strategicpatch"
)

// UpdateSecrets updates the secrets by adding the helm pre-install hook annotation
func UpdateSecrets(kubeClient *kubernetes.Clientset, secrets []string) error {
	succChan, errChan := make(chan string), make(chan error)

	for _, secret := range secrets {
		go updateSecret(kubeClient, secret, succChan, errChan)
	}
	for i := 0; i < len(secrets); i++ {
		select {
		case successMsg := <-succChan:
			fmt.Println(successMsg)
		case err := <-errChan:
			fmt.Println(err)
		}
	}
	return nil
}

// updateSecret annotates the secret if its present.
func updateSecret(kubeClient *kubernetes.Clientset, secretName string, succChan chan<- string, errChan chan<- error) {
	b := bytes.NewBuffer(nil)
	// Secrets
	secret, err := kubeClient.Secrets("deis").Get(secretName)
	if err != nil {
		if apierrors.IsNotFound(err) {
			succChan <- fmt.Sprintf("secret %s not found", secretName)
			return
		}
		errChan <- err
		return
	}
	secretNameDet := strings.SplitN(secret.ObjectMeta.Name, "-", 2)
	path := "workflow/charts/" + secretNameDet[1] + "templates/" + secretNameDet[1] + "-secret.yaml"
	b.WriteString("\n---\n# Source: " + path + "\n")
	secret.Kind = "Secret"
	secret.APIVersion = "v1"
	secret.ResourceVersion = ""
	y, err := yaml.Marshal(secret)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		errChan <- err
		return
	}
	b.WriteString(string(y))

	factory := cmdutil.NewFactory(nil)
	current := factory.NewBuilder().ContinueOnError().NamespaceParam("deis").DefaultNamespace().Stream(b, "").Flatten().Do()
	err = current.Visit(func(info *resource.Info, err error) error {
		if err != nil {
			return err
		}

		obj, err := cmdutil.MaybeConvertObject(info.Object, info.Mapping.GroupVersionKind.GroupVersion(), info.Mapping)
		if err != nil {
			return err
		}
		name, namespace := info.Name, info.Namespace
		oldData, err := json.Marshal(obj)
		if err != nil {
			return err
		}

		if err := updateAnnotations(obj); err != nil {
			return err
		}
		newData, err := json.Marshal(obj)
		if err != nil {
			return err
		}
		patchBytes, err := strategicpatch.CreateTwoWayMergePatch(oldData, newData, obj)
		createdPatch := err == nil
		if err != nil {
			fmt.Printf("couldn't compute patch: %v", err)
		}

		mapping := info.ResourceMapping()
		client, err := factory.ClientForMapping(mapping)
		if err != nil {
			return err
		}
		helper := resource.NewHelper(client, mapping)

		if createdPatch {
			_, err = helper.Patch(namespace, name, api.StrategicMergePatchType, patchBytes)
		} else {
			_, err = helper.Replace(namespace, name, false, obj)
		}
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		errChan <- err
		return
	}
	succChan <- fmt.Sprintf("secret %s annotated successfuly", secretName)
}

// updateAnnotations updates annotations of obj
func updateAnnotations(obj runtime.Object) error {
	accessor, err := meta.Accessor(obj)
	if err != nil {
		return err
	}

	annotations := accessor.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}

	annotations["helm.sh/hook"] = "pre-install"

	accessor.SetAnnotations(annotations)

	return nil
}
