// Copyright 2018 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package util

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/cenkalti/backoff"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func WaitForAPIAvailable(initializeTimeout time.Duration, basePath string, apiAddress string, scheme string) error {
	operation := func() error {
		response, err := http.Get(fmt.Sprintf("%s://%s%s/healthz", scheme, apiAddress, basePath))
		if err != nil {
			return err
		}

		// If we get a 503 service unavailable, it's a non-retriable error.
		if response.StatusCode == 503 {
			return backoff.Permanent(errors.Wrapf(
				err, "Waiting for ml pipeline API server failed with non retriable error."))
		}

		return nil
	}

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = initializeTimeout
	err := backoff.Retry(operation, b)
	return errors.Wrapf(err, "Waiting for ml pipeline API server failed after all attempts.")
}

// GetKubernetesConfig will first try an in-cluster configuration but fallback to using a kubeconfig.
func GetKubernetesConfig() (*rest.Config, error) {
	restConfig, errInCluster := rest.InClusterConfig()
	if errInCluster == nil {
		return restConfig, nil
	}

	// Fallback to using a kubeconfig
	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(), &clientcmd.ConfigOverrides{},
	)

	restConfig, errKubeconfig := clientConfig.ClientConfig()
	if errKubeconfig != nil {
		return nil, fmt.Errorf("%w; %w", errInCluster, errKubeconfig)
	}

	return restConfig, nil
}

func GetKubernetesClientFromClientConfig(clientConfig clientcmd.ClientConfig) (
	*kubernetes.Clientset, *rest.Config, string, error,
) {
	// Get the clientConfig
	config, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, nil, "", errors.Wrapf(err,
			"Failed to get cluster config during K8s client initialization")
	}
	// Get namespace
	namespace, _, err := clientConfig.Namespace()
	if err != nil {
		return nil, nil, "", errors.Wrapf(err,
			"Failed to get the namespace during K8s client initialization")
	}
	// create the clientset
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, "", errors.Wrapf(err,
			"Failed to create client set during K8s client initialization")
	}
	return clientSet, config, namespace, nil
}

func GetRpcConnectionWithTimeout(address string, tlsEnabled bool, caCertPath string, timeout time.Time) (*grpc.ClientConn, error) {
	creds := insecure.NewCredentials()
	if tlsEnabled {
		if caCertPath == "" {
			return nil, errors.New("CA cert path is empty")
		}

		caCert, err := os.ReadFile(caCertPath)
		if err != nil {
			return nil, errors.Wrap(err, "Encountered error when reading CA cert path for creating a metadata client.")
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		config := &tls.Config{
			RootCAs: caCertPool,
		}
		creds = credentials.NewTLS(config)
	}

	ctx, _ := context.WithDeadline(context.Background(), timeout)

	conn, err := grpc.DialContext(ctx, address, grpc.WithTransportCredentials(creds), grpc.WithBlock())
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to create gRPC connection")
	}
	return conn, nil
}

func GetRpcConnection(address string, tlsEnabled bool, caCertPath string) (*grpc.ClientConn, error) {
	creds := insecure.NewCredentials()
	if tlsEnabled {
		if caCertPath == "" {
			return nil, errors.New("CA cert path is empty")
		}

		caCert, err := os.ReadFile(caCertPath)
		if err != nil {
			return nil, errors.Wrap(err, "Encountered error when reading CA cert path for creating a metadata client.")
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		config := &tls.Config{
			RootCAs: caCertPool,
		}
		creds = credentials.NewTLS(config)
	}

	conn, err := grpc.Dial(
		address,
		grpc.WithTransportCredentials(creds),
	)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to create gRPC connection")
	}
	return conn, nil
}

func ExtractMasterIPAndPort(config *rest.Config) string {
	host := config.Host
	host = strings.TrimPrefix(host, "http://")
	host = strings.TrimPrefix(host, "https://")
	return host
}
