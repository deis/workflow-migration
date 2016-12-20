package pkg

// Most of the code in here is from the https://github.com/kubernetes/helm repo and
// changed appropriately for the use-case.

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"errors"
	"strconv"
	"time"

	"github.com/golang/protobuf/proto"
	"k8s.io/client-go/1.5/kubernetes"
	apierrors "k8s.io/client-go/1.5/pkg/api/errors"
	"k8s.io/client-go/1.5/pkg/api/v1"
	rspb "k8s.io/helm/pkg/proto/hapi/release"
)

const tillerNamespace = "kube-system"

var b64 = base64.StdEncoding

// CfgCreate creates a configmap based on the release object
func CfgCreate(key string, rls *rspb.Release, clientset *kubernetes.Clientset) error {
	// set labels for configmaps object meta data
	lbs := make(map[string]string)

	lbs["CREATED_AT"] = strconv.Itoa(int(time.Now().Unix()))

	// create a new configmap to hold the release
	obj, err := newConfigMapsObject(key, rls, lbs)
	if err != nil {
		return err
	}
	// push the configmap object out into the kubiverse
	if _, err := clientset.ConfigMaps(tillerNamespace).Create(obj); err != nil {
		if apierrors.IsAlreadyExists(err) {
			return errors.New("already exists")
		}

		return err
	}
	return nil
}

func newConfigMapsObject(key string, rls *rspb.Release, lbs map[string]string) (*v1.ConfigMap, error) {
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
	return &v1.ConfigMap{
		ObjectMeta: v1.ObjectMeta{
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
	var buf bytes.Buffer
	w, err := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	if err != nil {
		return "", err
	}
	if _, err = w.Write(b); err != nil {
		return "", err
	}
	w.Close()
	return b64.EncodeToString(buf.Bytes()), nil
}
