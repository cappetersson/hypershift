package kcm

import (
	"encoding/json"
	"fmt"
	"path"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	componentbasev1 "k8s.io/component-base/config/v1alpha1"
	"k8s.io/utils/pointer"

	kcpv1 "github.com/openshift/api/kubecontrolplane/v1"

	"github.com/openshift/hypershift/support/certs"
	"github.com/openshift/hypershift/support/config"
)

const (
	KubeControllerManagerConfigKey = "config.json"
	ServiceServingCAKey            = "service-ca.crt"
)

func ReconcileConfig(config, serviceServingCA *corev1.ConfigMap, ownerRef config.OwnerRef) error {
	ownerRef.ApplyTo(config)
	if config.Data == nil {
		config.Data = map[string]string{}
	}
	serializedConfig, err := generateConfig(serviceServingCA)
	if err != nil {
		return fmt.Errorf("failed to create apiserver config: %w", err)
	}
	config.Data[KubeControllerManagerConfigKey] = serializedConfig
	return nil
}

func generateConfig(serviceServingCA *corev1.ConfigMap) (string, error) {
	renewDeadline, err := time.ParseDuration("12s")
	if err != nil {
		return "", err
	}
	retryPeriod, err := time.ParseDuration("3s")
	if err != nil {
		return "", err
	}
	var serviceServingCAPath string
	if serviceServingCA != nil {
		serviceServingCAPath = path.Join(serviceServingCAMount.Path(kcmContainerMain().Name, kcmVolumeServiceServingCA().Name), ServiceServingCAKey)
	}
	config := kcpv1.KubeControllerManagerConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "KubeControllerManagerConfig",
			APIVersion: kcpv1.GroupVersion.String(),
		},
		ExtendedArguments: map[string]kcpv1.Arguments{},
		ServiceServingCert: kcpv1.ServiceServingCert{
			CertFile: serviceServingCAPath,
		},
		LeaderElection: componentbasev1.LeaderElectionConfiguration{
			LeaderElect:   pointer.BoolPtr(true),
			RenewDeadline: metav1.Duration{Duration: renewDeadline},
			RetryPeriod:   metav1.Duration{Duration: retryPeriod},
		},
	}
	b, err := json.Marshal(config)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func ReconcileKCMServiceServingCA(cm, combinedCA *corev1.ConfigMap, ownerRef config.OwnerRef) error {
	ownerRef.ApplyTo(cm)
	if cm.Data == nil {
		cm.Data = map[string]string{}
	}
	if _, hasKey := cm.Data[ServiceServingCAKey]; !hasKey {
		cm.Data[ServiceServingCAKey] = combinedCA.Data[certs.CASignerCertMapKey]
	}
	return nil
}

func ReconcileServiceAccount(sa *corev1.ServiceAccount) error {
	// nothing to reconcile
	return nil
}
