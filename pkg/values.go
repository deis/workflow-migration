package pkg

import (
	"bytes"
	"errors"
	"text/template"

	"k8s.io/client-go/1.5/kubernetes"
	apierrors "k8s.io/client-go/1.5/pkg/api/errors"
)

const (
	offCluster = "off-cluster"
	onCluster  = "on-cluster"
)

type valuesConfig struct {
	StorageLocation       string
	DatabaseLocation      string
	RedisLocation         string
	InfluxDBLocation      string
	GrafanaLocation       string
	RegistryLocation      string
	RegistryHostPort      string
	ImagePullSecretPrefix string
	S3                    s3
	GCS                   gcs
	Azure                 azure
	Swift                 swift
	Postgres              postgres
	Controller            controller
	Redis                 redis
	Grafana               grafana
	InfluxDB              influxDB
	ECR                   ecr
	GCR                   gcr
	OffClusterRegistry    offClusterRegistry
	Router                router
}

type s3 struct {
	AccessKey      string
	SecretKey      string
	Region         string
	RegistryBucket string
	DatabaseBucket string
	BuilderBucket  string
}

type gcs struct {
	KeyJSON        string
	RegistryBucket string
	DatabaseBucket string
	BuilderBucket  string
}

type azure struct {
	AccountName       string
	AccountKey        string
	RegistryContainer string
	DatabaseContainer string
	BuilderContainer  string
}

type swift struct {
	UserName          string
	Password          string
	Tenant            string
	AuthURL           string
	AuthVersion       string
	RegistryContainer string
	DatabaseContainer string
	BuilderContainer  string
}

type controller struct {
	AppPullPolicy    string
	RegistrationMode string
}

type postgres struct {
	Name     string
	UserName string
	Password string
	Host     string
	Port     string
}

type redis struct {
	DB       string
	Host     string
	Port     string
	Password string
}

type grafana struct {
	User     string
	Password string
}

type influxDB struct {
	Database string
	URL      string
	User     string
	Password string
}

type ecr struct {
	AccessKey  string
	SecretKey  string
	Region     string
	RegistryID string
	HostName   string
}

type gcr struct {
	KeyJSON  string
	HostName string
}

type offClusterRegistry struct {
	HostName     string
	Organization string
	UserName     string
	Password     string
}

type router struct {
	DHParam string
}

const (
	valuesTemplate = `#
# This is the main configuration file for Deis object storage. The values in
# this file are passed into the appropriate services so that they can configure
# themselves for persisting data in object storage.
#
# In general, all object storage credentials must be able to read and write to
# the container or bucket they are configured to use.
#

global:
  # Set the storage backend
  #
  # Valid values are:
  # - s3: Store persistent data in AWS S3 (configure in S3 section)
  # - azure: Store persistent data in Azure's object storage
  # - gcs: Store persistent data in Google Cloud Storage
  # - minio: Store persistent data on in-cluster Minio server
  storage: "{{ .StorageLocation }}"
  # Set the location of Workflow's PostgreSQL database
  #
  # Valid values are:
  # - on-cluster: Run PostgreSQL within the Kubernetes cluster (credentials are generated
  #   automatically; backups are sent to object storage
  #   configured above)
  # - off-cluster: Run PostgreSQL outside the Kubernetes cluster (configure in database section)
  database_location: "{{ .DatabaseLocation }}"
  # Set the location of Workflow's logger-specific Redis instance
  #
  # Valid values are:
  # - on-cluster: Run Redis within the Kubernetes cluster
  # - off-cluster: Run Redis outside the Kubernetes cluster (configure in loggerRedis section)
  logger_redis_location: "{{ .RedisLocation }}"

  # Set the location of Workflow's influxdb cluster
  #
  # Valid values are:
  # - on-cluster: Run Influxdb within the Kubernetes cluster
  # - off-cluster: Influxdb is running outside of the cluster and credentials and connection information will be provided.
  influxdb_location: "{{ .InfluxDBLocation }}"
  # Set the location of Workflow's grafana instance
  #
  # Valid values are:
  # - on-cluster: Run Grafana within the Kubernetes cluster
  # - off-cluster: Grafana is running outside of the cluster
  grafana_location: "{{ .GrafanaLocation }}"

  # Set the location of Workflow's Registry
  #
  # Valid values are:
  # - on-cluster: Run registry within the Kubernetes cluster
  # - off-cluster: Use registry outside the Kubernetes cluster (example: dockerhub,quay.io,self-hosted)
  # - ecr: Use Amazon's ECR
  # - gcr: Use Google's GCR
  registry_location: "{{ .RegistryLocation }}"
  # The host port to which registry proxy binds to
  host_port: {{ .RegistryHostPort }}
  # Prefix for the imagepull secret created when using private registry
  secret_prefix: "{{ .ImagePullSecretPrefix }}"

{{ if ne .S3.Region "" }}
s3:
  # Your AWS access key. Leave it empty if you want to use IAM credentials.
  accesskey: "{{ .S3.AccessKey }}"
  # Your AWS secret key. Leave it empty if you want to use IAM credentials.
  secretkey: "{{ .S3.SecretKey }}"
  # Any S3 region
  region: "{{ .S3.Region }}"
  # Your buckets.
  registry_bucket: "{{ .S3.RegistryBucket }}"
  database_bucket: "{{ .S3.DatabaseBucket }}"
  builder_bucket: "{{ .S3.BuilderBucket }}"
{{ end }}
{{ if ne .Azure.AccountName "" }}
azure:
  accountname: "{{ .Azure.AccountName }}"
  accountkey: "{{ .Azure.AccountKey }}"
  registry_container: "{{ .Azure.RegistryContainer }}"
  database_container: "{{ .Azure.DatabaseContainer }}"
  builder_container: "{{ .Azure.BuilderContainer }}"{{ end }}
{{ if ne .GCS.KeyJSON "" }}
gcs:
  # key_json is expanded into a JSON file on the remote server. It must be
  # well-formatted JSON data.
  key_json: '{{ .GCS.KeyJSON }}'
  registry_bucket: "{{ .GCS.RegistryBucket }}"
  database_bucket: "{{ .GCS.DatabaseBucket }}"
  builder_bucket: "{{ .GCS.BuilderBucket }}"{{ end }}
{{ if ne .Swift.UserName "" }}
swift:
  username: "{{ .Swift.UserName }}"
  password: "{{ .Swift.Password }}"
  authurl: "{{ .Swift.AuthURL }}"
  # Your OpenStack tenant name if you are using auth version 2 or 3.
  tenant: "{{ .Swift.Tenant }}"
  authversion: "{{ .Swift.AuthVersion }}"
  registry_container: "{{ .Swift.RegistryContainer }}"
  database_container: "{{ .Swift.DatabaseContainer }}"
  builder_container: "{{ .Swift.BuilderContainer }}"{{ end }}

# Set the default (global) way of how Application (your own) images are
# pulled from within the Controller.
# This can be configured per Application as well in the Controller.
#
# This affects pull apps and git push (slugrunner images) apps
#
# Values values are:
# - Always
# - IfNotPresent
controller:
  app_pull_policy: "{{ .Controller.AppPullPolicy }}"
  # Possible values are:
  # enabled - allows for open registration
  # disabled - turns off open registration
  # admin_only - allows for registration by an admin only.
  registration_mode: "{{ .Controller.RegistrationMode }}"
{{ if ne .Postgres.Name "" }}
database:
  # Configure the following ONLY if using an off-cluster PostgreSQL database
  postgres:
    name: "{{ .Postgres.Name }}"
    username: "{{ .Postgres.UserName }}"
    password: "{{ .Postgres.Password }}"
    host: "{{ .Postgres.Host }}"
    port: "{{ .Postgres.Port }}"{{ end }}
{{ if ne .Redis.DB "" }}
logger:
  redis:
    # Configure the following ONLY if using an off-cluster Redis instance for logger
    db: "{{ .Redis.DB }}"
    host: "{{ .Redis.Host }}"
    port: "{{ .Redis.Port }}"
    password: "{{ .Redis.Password }}"{{ end }}
{{ if ne .Grafana.User "" }}
monitor:
  grafana:
    user: "{{ .Grafana.User }}"
    password: "{{ .Grafana.Password }}"
  # Configure the following ONLY if using an off-cluster Influx database
  influxdb:
    url: "{{ .InfluxDB.URL }}"
    database: "{{ .InfluxDB.Database }}"
    user: "{{ .InfluxDB.User }}"
    password: "{{ .InfluxDB.Password }}"{{ end }}

registry-token-refresher:
  # Time in minutes after which the token should be refreshed.
  # Leave it empty to use the default provider time.
  token_refresh_time: ""{{ if ne .OffClusterRegistry.UserName "" }}
  off_cluster_registry:
    hostname: "{{ .OffClusterRegistry.HostName }}"
    organization: "{{ .OffClusterRegistry.Organization }}"
    username: "{{ .OffClusterRegistry.UserName }}"
    password: "{{ .OffClusterRegistry.Password }}"{{ end }}{{ if ne .ECR.Region "" }}
  ecr:
    # Your AWS access key. Leave it empty if you want to use IAM credentials.
    accesskey: "{{ .ECR.AccessKey }}"
    # Your AWS secret key. Leave it empty if you want to use IAM credentials.
    secretkey: "{{ .ECR.SecretKey }}"
    # Any S3 region
    region: "{{ .ECR.Region }}"
    registryid: "{{ .ECR.RegistryID }}"
    hostname: "{{ .ECR.HostName }}"{{ end }}{{ if ne .GCR.KeyJSON "" }}
  gcr:
    key_json: '{{ .GCR.KeyJSON }}'
    hostname: "{{ .GCR.HostName }}"{{ end }}
{{ if ne .Router.DHParam "" }}
router:
  dhparam: "{{ .Router.DHParam }}"{{ end }}
`
)

func (v *valuesConfig) updateStorageparams(kubeClient *kubernetes.Clientset) error {
	objSecret, err := kubeClient.Secrets("deis").Get("objectstorage-keyfile")
	if err != nil {
		return err
	}
	val, ok := objSecret.GetAnnotations()["deis.io/objectstorage"]
	if !ok {
		return errors.New("storage type can't be found")
	}
	v.StorageLocation = val
	switch val {
	case "s3":
		v.S3 = s3{
			AccessKey:      string(objSecret.Data["accesskey"]),
			SecretKey:      string(objSecret.Data["secretkey"]),
			Region:         string(objSecret.Data["region"]),
			RegistryBucket: string(objSecret.Data["registry-bucket"]),
			DatabaseBucket: string(objSecret.Data["database-bucket"]),
			BuilderBucket:  string(objSecret.Data["builder-bucket"]),
		}
	case "gcs":
		v.GCS = gcs{
			KeyJSON:        string(objSecret.Data["key.json"]),
			RegistryBucket: string(objSecret.Data["registry-bucket"]),
			DatabaseBucket: string(objSecret.Data["database-bucket"]),
			BuilderBucket:  string(objSecret.Data["builder-bucket"]),
		}
	case "azure":
		v.Azure = azure{
			AccountName:       string(objSecret.Data["accountname"]),
			AccountKey:        string(objSecret.Data["accountkey"]),
			RegistryContainer: string(objSecret.Data["registry-container"]),
			DatabaseContainer: string(objSecret.Data["database-container"]),
			BuilderContainer:  string(objSecret.Data["builder-container"]),
		}
	case "swift":
		v.Swift = swift{
			UserName:          string(objSecret.Data["username"]),
			Password:          string(objSecret.Data["password"]),
			Tenant:            string(objSecret.Data["tenant"]),
			AuthURL:           string(objSecret.Data["authurl"]),
			AuthVersion:       string(objSecret.Data["authversion"]),
			RegistryContainer: string(objSecret.Data["registry-container"]),
			DatabaseContainer: string(objSecret.Data["database-container"]),
			BuilderContainer:  string(objSecret.Data["builder-container"]),
		}
	default:
		return errors.New("Not a valid storage type")
	}
	return nil
}

func (v *valuesConfig) updateRegistryparams(kubeClient *kubernetes.Clientset) error {
	v.RegistryLocation = onCluster
	objSecret, err := kubeClient.Secrets("deis").Get("registry-secret")
	if err != nil && !apierrors.IsNotFound(err) {
		return err
	}
	if err == nil {

		val, ok := objSecret.GetAnnotations()["deis.io/registry-location"]
		if !ok {
			return errors.New("Registry Location can't be found")
		}
		v.RegistryLocation = val
		switch val {
		case "ecr":
			v.ECR = ecr{
				AccessKey:  string(objSecret.Data["accesskey"]),
				SecretKey:  string(objSecret.Data["secretkey"]),
				Region:     string(objSecret.Data["region"]),
				RegistryID: string(objSecret.Data["registryid"]),
				HostName:   string(objSecret.Data["hostname"]),
			}
		case "gcr":
			v.GCR = gcr{
				KeyJSON:  string(objSecret.Data["key.json"]),
				HostName: string(objSecret.Data["hostname"]),
			}
		case "off-cluster":
			v.OffClusterRegistry = offClusterRegistry{
				HostName:     string(objSecret.Data["hostname"]),
				Organization: string(objSecret.Data["organization"]),
				UserName:     string(objSecret.Data["username"]),
				Password:     string(objSecret.Data["password"]),
			}
		}
	}
	v.RegistryHostPort = "5555"
	v.ImagePullSecretPrefix = ""
	controllerDeployment, err := kubeClient.Deployments("deis").Get("deis-controller")
	if err != nil {
		return err
	}
	envs := controllerDeployment.Spec.Template.Spec.Containers[0].Env
	for _, env := range envs {
		if env.Name == "DEIS_REGISTRY_SERVICE_PORT" {
			v.RegistryHostPort = env.Value
		}
		if env.Name == "DEIS_REGISTRY_SECRET_PREFIX" {
			v.ImagePullSecretPrefix = env.Value
		}
	}

	return nil
}

func (v *valuesConfig) updateRedisparams(kubeClient *kubernetes.Clientset) error {
	v.RedisLocation = onCluster
	v.Redis = redis{}
	loggerDeployment, err := kubeClient.Deployments("deis").Get("deis-logger")
	if err != nil {
		return err
	}
	envs := loggerDeployment.Spec.Template.Spec.Containers[0].Env
	for _, env := range envs {
		if env.Name == "DEIS_LOGGER_REDIS_DB" {
			v.Redis.DB = env.Value
		}
		if env.Name == "DEIS_LOGGER_REDIS_SERVICE_HOST" {
			v.Redis.Host = env.Value
		}
		if env.Name == "DEIS_LOGGER_REDIS_SERVICE_PORT" {
			v.Redis.Port = env.Value
		}
	}
	if v.Redis.Host != "" {
		redisSecret, err := kubeClient.Secrets("deis").Get("logger-redis-creds")
		if err != nil {
			return err
		}
		v.Redis.Password = string(redisSecret.Data["password"])
		v.RedisLocation = offCluster
		// Update the redis secret as the secret template updated in the new helm charts.
		// `helm upgrade` doesn't upgrade as this set as pre-install hook.
		redisSecret.Data["db"] = []byte(v.Redis.DB)
		redisSecret.Data["host"] = []byte(v.Redis.Host)
		redisSecret.Data["port"] = []byte(v.Redis.Port)
		if _, err := kubeClient.Secrets("deis").Update(redisSecret); err != nil {
			return err
		}
	}

	return nil
}

func (v *valuesConfig) updateDatabaseParams(kubeClient *kubernetes.Clientset) error {
	v.DatabaseLocation = onCluster
	controllerDeployment, err := kubeClient.Deployments("deis").Get("deis-controller")
	if err != nil && !apierrors.IsNotFound(err) {
		return err
	}
	if err == nil {

		postgresDetails := postgres{}
		envs := controllerDeployment.Spec.Template.Spec.Containers[0].Env
		for _, env := range envs {
			if env.Name == "DEIS_DATABASE_NAME" {
				postgresDetails.Name = env.Value
			}
			if env.Name == "DEIS_DATABASE_SERVICE_HOST" {
				postgresDetails.Host = env.Value
			}
			if env.Name == "DEIS_DATABASE_SERVICE_PORT" {
				postgresDetails.Port = env.Value
			}
		}
		if postgresDetails.Name != "" {
			postgresSecret, err := kubeClient.Secrets("deis").Get("database-creds")
			if err != nil {
				return err
			}
			postgresDetails.UserName = string(postgresSecret.Data["user"])
			postgresDetails.Password = string(postgresSecret.Data["password"])
			v.Postgres = postgresDetails
			v.DatabaseLocation = offCluster
			// Update the database secret as the secret template updated in the new helm charts.
			// `helm upgrade` doesn't upgrade as this set as pre-install hook.
			postgresSecret.Data["name"] = []byte(postgresDetails.Name)
			postgresSecret.Data["host"] = []byte(postgresDetails.Host)
			postgresSecret.Data["port"] = []byte(postgresDetails.Port)
			if _, err := kubeClient.Secrets("deis").Update(postgresSecret); err != nil {
				return err
			}
		}
	}
	return nil
}

func (v *valuesConfig) updateInfluxparams(kubeClient *kubernetes.Clientset) error {
	v.InfluxDBLocation = onCluster
	telegrafDaemonSet, err := kubeClient.DaemonSets("deis").Get("deis-monitor-telegraf")
	if err != nil {
		return err
	}
	influxDetails := influxDB{}
	envs := telegrafDaemonSet.Spec.Template.Spec.Containers[0].Env
	for _, env := range envs {
		if env.Name == "INFLUXDB_USERNAME" {
			influxDetails.User = env.Value
		}
		if env.Name == "INFLUXDB_PASSWORD" {
			influxDetails.Password = env.Value
		}
		if env.Name == "INFLUXDB_URLS" {
			influxDetails.URL = env.Value
		}
		if env.Name == "INFLUXDB_DATABASE" {
			influxDetails.Database = env.Value
		}
	}
	if influxDetails.User != "" {
		v.InfluxDBLocation = offCluster
	}
	return nil
}

func (v *valuesConfig) updateGrafanaparams(kubeClient *kubernetes.Clientset) error {
	v.GrafanaLocation = onCluster
	_, err := kubeClient.Deployments("deis").Get("deis-monitor-grafana")
	if err != nil {
		if apierrors.IsNotFound(err) {
			v.GrafanaLocation = offCluster
		}
		return err
	}
	return nil
}

func (v *valuesConfig) updateControllerparams(kubeClient *kubernetes.Clientset) error {
	v.Controller = controller{
		AppPullPolicy:    "IfNotPresent",
		RegistrationMode: "enabled",
	}
	controllerDeployment, err := kubeClient.Deployments("deis").Get("deis-controller")
	if err != nil {
		return err
	}
	envs := controllerDeployment.Spec.Template.Spec.Containers[0].Env
	for _, env := range envs {
		if env.Name == "REGISTRATION_MODE" {
			v.Controller.RegistrationMode = env.Value
		}
		if env.Name == "IMAGE_PULL_POLICY" {
			v.Controller.AppPullPolicy = env.Value
		}
	}

	return nil
}

// GetValues gets the values used for cluster configuration
func GetValues(kubeClient *kubernetes.Clientset) (string, error) {
	workflowConfig := &valuesConfig{}
	err := workflowConfig.updateStorageparams(kubeClient)
	if err != nil {
		return "", err
	}
	err = workflowConfig.updateDatabaseParams(kubeClient)
	if err != nil {
		return "", err
	}
	err = workflowConfig.updateGrafanaparams(kubeClient)
	if err != nil {
		return "", err
	}
	err = workflowConfig.updateInfluxparams(kubeClient)
	if err != nil {
		return "", err
	}
	err = workflowConfig.updateRedisparams(kubeClient)
	if err != nil {
		return "", err
	}
	err = workflowConfig.updateRegistryparams(kubeClient)
	if err != nil {
		return "", err
	}
	err = workflowConfig.updateControllerparams(kubeClient)
	if err != nil {
		return "", err
	}

	tmpl, err := template.New("values").Parse(valuesTemplate)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, workflowConfig)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
