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

package cluster

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
	"oneinfra.ereslibre.es/m/internal/pkg/cluster/endpoint"
	"oneinfra.ereslibre.es/m/internal/pkg/manifests"
)

// IngressNodeName prints the ingress node name for this cluster
func IngressNodeName(clusterName string) error {
	stdin, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return err
	}
	clusters := manifests.RetrieveClusters(string(stdin))
	nodes := manifests.RetrieveNodes(string(stdin))

	cluster, ok := clusters[clusterName]
	if !ok {
		return errors.Errorf("cluster %q not found", clusterName)
	}

	ingressNode, err := endpoint.IngressNode(nodes, cluster)
	if err != nil {
		return err
	}

	fmt.Println(ingressNode.HypervisorName)

	return nil
}