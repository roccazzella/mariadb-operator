package v1alpha1

import (
	"errors"
	"fmt"

	"github.com/mariadb-operator/mariadb-operator/pkg/webhook"
	cron "github.com/robfig/cron/v3"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	inmutableWebhook = webhook.NewInmutableWebhook(
		webhook.WithTagName("webhook"),
	)
	cronParser = cron.NewParser(
		cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow,
	)
)

type Image struct {
	// +kubebuilder:validation:Required
	Repository string `json:"repository"`
	// +kubebuilder:default=latest
	Tag string `json:"tag,omitempty"`
	// +kubebuilder:default=IfNotPresent
	PullPolicy corev1.PullPolicy `json:"pullPolicy,omitempty"`
}

func (i *Image) String() string {
	return fmt.Sprintf("%s:%s", i.Repository, i.Tag)
}

type MariaDBRef struct {
	// +kubebuilder:validation:Required
	corev1.LocalObjectReference `json:",inline"`
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=true
	WaitForIt bool `json:"waitForIt"`
}

type SecretTemplate struct {
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
	Key         *string           `json:"key,omitempty"`
	UsernameKey *string           `json:"usernameKey,omitempty"`
	PasswordKey *string           `json:"passwordKey,omitempty"`
	HostKey     *string           `json:"hostKey,omitempty"`
	PortKey     *string           `json:"portKey,omitempty"`
	DatabaseKey *string           `json:"databaseKey,omitempty"`
}

type HealthCheck struct {
	Interval      *metav1.Duration `json:"interval,omitempty"`
	RetryInterval *metav1.Duration `json:"retryInterval,omitempty"`
}

type ConnectionTemplate struct {
	SecretName     *string           `json:"secretName,omitempty" webhook:"inmutableinit"`
	SecretTemplate *SecretTemplate   `json:"secretTemplate,omitempty" webhook:"inmutableinit"`
	HealthCheck    *HealthCheck      `json:"healthCheck,omitempty"`
	Params         map[string]string `json:"params,omitempty" webhook:"inmutable"`
	ServiceName    *string           `json:"serviceName,omitempty" webhook:"inmutable"`
}

type RestoreSource struct {
	BackupRef *corev1.LocalObjectReference `json:"backupRef,omitempty" webhook:"inmutableinit"`
	Volume    *corev1.VolumeSource         `json:"volume,omitempty" webhook:"inmutableinit"`
	FileName  *string                      `json:"fileName,omitempty" webhook:"inmutableinit"`
}

func (r *RestoreSource) IsInit() bool {
	return r.Volume != nil
}

func (r *RestoreSource) Init(backup *Backup) {
	if backup.Spec.Storage.Volume != nil {
		r.Volume = backup.Spec.Storage.Volume
	}
	if backup.Spec.Storage.PersistentVolumeClaim != nil {
		r.Volume = &corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: backup.Name,
			},
		}
	}
}

func (r *RestoreSource) Validate() error {
	if r.BackupRef != nil {
		return nil
	}
	if r.Volume == nil {
		return errors.New("unable to determine restore source")
	}
	return nil
}

type Schedule struct {
	// +kubebuilder:validation:Required
	Cron string `json:"cron"`
	// +kubebuilder:default=false
	Suspend bool `json:"suspend,omitempty"`
}

func (s *Schedule) Validate() error {
	_, err := cronParser.Parse(s.Cron)
	return err
}
