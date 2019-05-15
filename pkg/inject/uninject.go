package inject

import (
	"strings"

	"github.com/linkerd/linkerd2/pkg/k8s"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Uninject removes from the workload in conf the init and proxy containers,
// the TLS volumes and the extra annotations/labels that were added
func (conf *ResourceConfig) Uninject(report *Report) ([]byte, error) {
	if conf.pod.spec == nil {
		return nil, nil
	}

	conf.uninjectPodSpec(report)

	if conf.workload.Meta != nil {
		uninjectObjectMeta(conf.workload.Meta)
	}

	uninjectObjectMeta(conf.pod.meta)
	return conf.YamlMarshalObj()
}

// Given a PodSpec, update the PodSpec in place with the sidecar
// and init-container uninjected
func (conf *ResourceConfig) uninjectPodSpec(report *Report) {
	t := conf.pod.spec

	if conf.pod.meta.Annotations[k8s.AutomountServiceAccountTokenAnnotation] == k8s.AutomountServiceAccountTokenEnabled {
		var disableAutomountServiceAccountToken bool
		t.AutomountServiceAccountToken = &disableAutomountServiceAccountToken
	}

	initContainers := []v1.Container{}
	for _, container := range t.InitContainers {
		if container.Name != k8s.InitContainerName {
			initContainers = append(initContainers, container)
		} else {
			report.Uninjected.ProxyInit = true
		}
	}
	t.InitContainers = initContainers

	containers := []v1.Container{}
	for _, container := range t.Containers {
		if container.Name != k8s.ProxyContainerName {
			containers = append(containers, container)
		} else {
			report.Uninjected.Proxy = true
		}
	}
	t.Containers = containers

	volumes := []v1.Volume{}
	for _, volume := range t.Volumes {
		if volume.Name != k8s.IdentityEndEntityVolumeName {
			volumes = append(volumes, volume)
		}
	}
	t.Volumes = volumes
}

func uninjectObjectMeta(t *metav1.ObjectMeta) {
	newAnnotations := make(map[string]string)
	for key, val := range t.Annotations {
		if !strings.HasPrefix(key, k8s.Prefix) || key == k8s.ProxyInjectAnnotation {
			newAnnotations[key] = val
		}
	}
	t.Annotations = newAnnotations

	labels := make(map[string]string)
	for key, val := range t.Labels {
		if !strings.HasPrefix(key, k8s.Prefix) {
			labels[key] = val
		}
	}
	t.Labels = labels
}
