package pkg

// Most of the code in here is from the https://github.com/kubernetes/helm repo and
// changed appropriately for the use-case.

import (
	"encoding/base64"
	"errors"
	"strconv"
	"time"

	"github.com/golang/protobuf/proto"
	rspb "k8s.io/helm/pkg/proto/hapi/release"
	"k8s.io/kubernetes/pkg/api"
	apierrors "k8s.io/kubernetes/pkg/api/errors"
	kcl "k8s.io/kubernetes/pkg/client/unversioned"
)

const tillerNamespace = "kube-system"

var b64 = base64.StdEncoding

// CfgCreate creates a configmap based on the release object
func CfgCreate(key string, rls *rspb.Release, kubeClient *kcl.Client) error {
	// set labels for configmaps object meta data
	lbs := make(map[string]string)

	lbs["CREATED_AT"] = strconv.Itoa(int(time.Now().Unix()))

	// create a new configmap to hold the release
	obj, err := newConfigMapsObject(key, rls, lbs)
	if err != nil {
		return err
	}
	// push the configmap object out into the kubiverse
	if _, err := kubeClient.ConfigMaps(tillerNamespace).Create(obj); err != nil {
		if apierrors.IsAlreadyExists(err) {
			return errors.New("already exists")
		}

		return err
	}
	return nil
}

func newConfigMapsObject(key string, rls *rspb.Release, lbs map[string]string) (*api.ConfigMap, error) {
	const owner = "TILLER"

	// encode the release
	s, err := encodeRelease(rls)
	if err != nil {
		return nil, err
	}

	// apply labels
	lbs["NAME"] = rls.Name
	lbs["OWNER"] = owner
	lbs["STATUS"] = rspb.Status_Code_name[int32(rls.Info.Status.Code)]
	lbs["VERSION"] = strconv.Itoa(int(rls.Version))

	// create and return configmap object
	return &api.ConfigMap{
		ObjectMeta: api.ObjectMeta{
			Name:   key,
			Labels: lbs,
		},
		Data: map[string]string{"release": s},
	}, nil
}

// encodeRelease encodes a release returning a base64 encoded
// binary protobuf encoding representation, or error.
func encodeRelease(rls *rspb.Release) (string, error) {
	b, err := proto.Marshal(rls)
	if err != nil {
		return "", err
	}
	return b64.EncodeToString(b), nil
}
