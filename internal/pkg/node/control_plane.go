/*
Copyright 2020 Rafael Fernández López <ereslibre@ereslibre.es>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package node

import (
	"fmt"

	"oneinfra.ereslibre.es/m/internal/pkg/cluster"
	"oneinfra.ereslibre.es/m/internal/pkg/infra"
)

const (
	dqliteImage                = "oneinfra/dqlite:latest"
	kineImage                  = "oneinfra/kine:latest"
	kubeAPIServerImage         = "k8s.gcr.io/kube-apiserver:v1.17.0"
	kubeControllerManagerImage = "k8s.gcr.io/kube-controller-manager:v1.17.0"
	kubeSchedulerImage         = "k8s.gcr.io/kube-scheduler:v1.17.0"
)

// ControlPlane represents a complete control plane instance,
// including: etcd, API server, controller-manager and scheduler
type ControlPlane struct{}

// Reconcile reconciles the kube-apiserver
func (controlPlane *ControlPlane) Reconcile(hypervisor *infra.Hypervisor, cluster *cluster.Cluster, node *Node) error {
	if err := hypervisor.PullImages(kineImage, kubeAPIServerImage, kubeControllerManagerImage, kubeSchedulerImage); err != nil {
		return err
	}
	controllerManagerKubeConfig, err := cluster.KubeConfig("https://127.0.0.1:6443")
	if err != nil {
		return err
	}
	schedulerKubeConfig, err := cluster.KubeConfig("https://127.0.0.1:6443")
	if err != nil {
		return err
	}
	err = hypervisor.UploadFiles(
		map[string]string{
			// API server secrets
			secretsPathFile(cluster, "apiserver-client-ca.crt"): cluster.CertificateAuthorities.APIServerClient.Certificate,
			secretsPathFile(cluster, "apiserver.crt"):           cluster.APIServer.TLSCert,
			secretsPathFile(cluster, "apiserver.key"):           cluster.APIServer.TLSPrivateKey,
			// controller-manager secrets
			secretsPathFile(cluster, "controller-manager.kubeconfig"): controllerManagerKubeConfig,
			// scheduler secrets
			secretsPathFile(cluster, "scheduler.kubeconfig"): schedulerKubeConfig,
		},
	)
	if err != nil {
		return err
	}
	_, err = hypervisor.RunPod(
		cluster,
		infra.NewPod(
			fmt.Sprintf("kube-apiserver-%s", cluster.Name),
			[]infra.Container{
				{
					Name:    "kine",
					Image:   kineImage,
					Command: []string{"kine"},
				},
				{
					Name:    "kube-apiserver",
					Image:   kubeAPIServerImage,
					Command: []string{"kube-apiserver"},
					Args: []string{
						"--etcd-servers", "http://127.0.0.1:2379",
						"--tls-cert-file", secretsPathFile(cluster, "apiserver.crt"),
						"--tls-private-key-file", secretsPathFile(cluster, "apiserver.key"),
						"--client-ca-file", secretsPathFile(cluster, "apiserver-client-ca.crt"),
					},
					Mounts: map[string]string{
						secretsPath(cluster): secretsPath(cluster),
					},
				},
				{
					Name:    "kube-controller-manager",
					Image:   kubeControllerManagerImage,
					Command: []string{"kube-controller-manager"},
					Args: []string{
						"--kubeconfig", secretsPathFile(cluster, "controller-manager.kubeconfig"),
					},
					Mounts: map[string]string{
						secretsPath(cluster): secretsPath(cluster),
					},
				},
				{
					Name:    "kube-scheduler",
					Image:   kubeSchedulerImage,
					Command: []string{"kube-scheduler"},
					Args: []string{
						"--kubeconfig", secretsPathFile(cluster, "scheduler.kubeconfig"),
					},
					Mounts: map[string]string{
						secretsPath(cluster): secretsPath(cluster),
					},
				},
			},
			map[int]int{
				node.HostPort: 6443,
			},
		),
	)
	return err
}
