package secret

import (
	"context"
	"fmt"
	"pandax/base/global"

	v1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"pandax/apps/devops/entity/k8s"
	k8scommon "pandax/apps/devops/services/k8s/common"
	"pandax/apps/devops/services/k8s/dataselect"
)

// SecretSpec is a common interface for the specification of different secrets.
type SecretSpec interface {
	GetName() string
	GetType() v1.SecretType
	GetNamespace() string
	GetData() map[string][]byte
}

// ImagePullSecretSpec is a specification of an image pull secret implements SecretSpec
type ImagePullSecretSpec struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`

	// The value of the .dockercfg property. It must be Base64 encoded.
	Data []byte `json:"data"`
}

// GetName returns the name of the ImagePullSecret
func (spec *ImagePullSecretSpec) GetName() string {
	return spec.Name
}

// GetType returns the type of the ImagePullSecret, which is always api.SecretTypeDockercfg
func (spec *ImagePullSecretSpec) GetType() v1.SecretType {
	return v1.SecretTypeDockercfg
}

// GetNamespace returns the namespace of the ImagePullSecret
func (spec *ImagePullSecretSpec) GetNamespace() string {
	return spec.Namespace
}

// GetData returns the data the secret carries, it is a single key-value pair
func (spec *ImagePullSecretSpec) GetData() map[string][]byte {
	return map[string][]byte{v1.DockerConfigKey: spec.Data}
}

// Secret is a single secret returned to the frontend.
type Secret struct {
	ObjectMeta k8s.ObjectMeta `json:"objectMeta"`
	TypeMeta   k8s.TypeMeta   `json:"typeMeta"`
	Type       v1.SecretType  `json:"type"`
}

// SecretList is a response structure for a queried secrets list.
type SecretList struct {
	k8s.ListMeta `json:"listMeta"`

	// Unordered list of Secrets.
	Secrets []Secret `json:"secrets"`
}

// GetSecretList returns all secrets in the given namespace.
func GetSecretList(client kubernetes.Interface, namespace *k8scommon.NamespaceQuery, dsQuery *dataselect.DataSelectQuery) (*SecretList, error) {
	global.Log.Info(fmt.Sprintf("Getting list of secrets in %s namespace", namespace))
	secretList, err := client.CoreV1().Secrets(namespace.ToRequestParam()).List(context.TODO(), k8s.ListEverything)
	if err != nil {
		return nil, err
	}

	return ToSecretList(secretList.Items, dsQuery), nil
}

// CreateSecret creates a single secret using the cluster API client
func CreateSecret(client kubernetes.Interface, spec SecretSpec) (*Secret, error) {
	namespace := spec.GetNamespace()
	secret := &v1.Secret{
		ObjectMeta: metaV1.ObjectMeta{
			Name:      spec.GetName(),
			Namespace: namespace,
		},
		Type: spec.GetType(),
		Data: spec.GetData(),
	}
	_, err := client.CoreV1().Secrets(namespace).Create(context.TODO(), secret, metaV1.CreateOptions{})
	result := toSecret(secret)
	return &result, err
}

func toSecret(secret *v1.Secret) Secret {
	return Secret{
		ObjectMeta: k8s.NewObjectMeta(secret.ObjectMeta),
		TypeMeta:   k8s.NewTypeMeta(k8s.ResourceKindSecret),
		Type:       secret.Type,
	}
}

func ToSecretList(secrets []v1.Secret, dsQuery *dataselect.DataSelectQuery) *SecretList {
	newSecretList := &SecretList{
		ListMeta: k8s.ListMeta{TotalItems: len(secrets)},
		Secrets:  make([]Secret, 0),
	}

	secretCells, filteredTotal := dataselect.GenericDataSelectWithFilter(toCells(secrets), dsQuery)
	secrets = fromCells(secretCells)
	newSecretList.ListMeta = k8s.ListMeta{TotalItems: filteredTotal}

	for _, secret := range secrets {
		newSecretList.Secrets = append(newSecretList.Secrets, toSecret(&secret))
	}

	return newSecretList
}

func DeleteSecret(client *kubernetes.Clientset, namespace string, name string) error {
	global.Log.Info(fmt.Sprintf("请求删除Secret: %v, namespace: %v", name, namespace))
	return client.CoreV1().Secrets(namespace).Delete(
		context.TODO(),
		name,
		metaV1.DeleteOptions{},
	)
}

func DeleteCollectionSecret(client *kubernetes.Clientset, secretList []k8s.SecretsData) (err error) {
	global.Log.Info("批量删除Secret开始")
	for _, v := range secretList {
		global.Log.Info(fmt.Sprintf("delete Secret：%v, ns: %v", v.Name, v.Namespace))
		err := client.CoreV1().Secrets(v.Namespace).Delete(
			context.TODO(),
			v.Name,
			metaV1.DeleteOptions{},
		)
		if err != nil {
			global.Log.Error(err.Error())
			return err
		}
	}
	global.Log.Info("删除Secret已完成")
	return nil
}
