// Copyright 2018-2023 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package testutil

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"
	"time"

	"github.com/kubeflow/pipelines/backend/src/common/util"
	"github.com/kubeflow/pipelines/backend/test/logger"

	"github.com/onsi/gomega"
	authenticationv1 "k8s.io/api/authentication/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	logReadMaxAttempts = 4
	logReadRetryDelay  = 5 * time.Second
)

func CreateK8sClient() (*kubernetes.Clientset, error) {
	restConfig, configErr := util.GetKubernetesConfig()
	if configErr != nil {
		return nil, configErr
	}
	k8sClient, clientErr := kubernetes.NewForConfig(restConfig)
	if clientErr != nil {
		return nil, clientErr
	}
	return k8sClient, nil
}

// ReadContainerLogs - Read pod logs from a specific container
func ReadContainerLogs(client *kubernetes.Clientset, namespace string, containerName string, follow *bool, sinceTime *time.Time, logLimit *int64) string {
	pod := GetPodContainingContainer(client, namespace, containerName)
	if pod != nil {
		return ReadPodLogs(client, namespace, pod.Name, follow, sinceTime, logLimit)
	} else {
		return fmt.Sprintf("Could not find pod containing container with name '%s'", containerName)
	}
}

// ReadPodLogs - Read pod logs from a specific names, with container name containing a substring and from a certain time period (default being from past 1 min)
func ReadPodLogs(client *kubernetes.Clientset, namespace string, podName string, follow *bool, sinceTime *time.Time, logLimit *int64) string {
	podFromPodName := GetPodContainingName(client, namespace, podName)
	podLogOptions := GetDefaultPodLogOptions()
	if logLimit != nil {
		podLogOptions.LimitBytes = logLimit
	}
	if follow != nil {
		podLogOptions.Follow = *follow
	}
	if sinceTime != nil {
		timeSince := metav1.NewTime(sinceTime.UTC())
		podLogOptions.SinceTime = &timeSince
	}
	buf := new(bytes.Buffer)
	if podFromPodName != nil {
		for _, container := range podFromPodName.Spec.Containers {
			podLogOptions.Container = container.Name
			logs, err := readContainerLogsWithRetry(client, namespace, podFromPodName.Name, podLogOptions)
			if err != nil {
				logger.Log("Failed to stream pod logs for pod %s container %s due to %v", podFromPodName.Name, container.Name, err)
				continue
			}
			buf.WriteString(logs)

			// If the current container restarted and current logs are empty, attempt previous logs once.
			if strings.TrimSpace(logs) == "" {
				previousOptions := *podLogOptions
				previousOptions.Previous = true
				previousLogs, previousErr := readContainerLogsOnce(client, namespace, podFromPodName.Name, &previousOptions)
				if previousErr == nil && strings.TrimSpace(previousLogs) != "" {
					buf.WriteString(previousLogs)
				}
			}
		}
		if buf.Len() == 0 {
			logger.Log("No pod logs available for pod '%s'. %s", podFromPodName.Name, formatPodStatus(podFromPodName))
		}
	} else {
		logger.Log("No pod logs available for pod with name '%s'.", podName)
	}
	return buf.String()
}

// GetDefaultPodLogOptions - Get default pod log options for the pod log reader API request
func GetDefaultPodLogOptions() *v1.PodLogOptions {
	logLimit := int64(50000000)
	sinceTime := metav1.NewTime(time.Now().Add(-1 * time.Minute).UTC())
	return &v1.PodLogOptions{
		Previous:   false,
		SinceTime:  &sinceTime,
		Timestamps: true,
		LimitBytes: &logLimit,
		Follow:     false,
	}
}

// GetPodContainingName - Get the name of the pod with name containing substring
func GetPodContainingName(client *kubernetes.Clientset, namespace, podName string) *v1.Pod {
	pods, err := client.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		logger.Log("Failed to list pods due to: %v", err)
	}
	// Exact name match first.
	for _, pod := range pods.Items {
		if pod.Name == podName {
			return &pod
		}
	}
	// Prefix match keeps semantics for workflow-generated suffixes.
	for _, pod := range pods.Items {
		if strings.HasPrefix(pod.Name, podName+"-") {
			return &pod
		}
	}
	for _, pod := range pods.Items {
		podNameSplit := strings.Split(pod.Name, "-")
		expectedPodNameSplit := strings.Split(podName, "-")
		contains := true
		for _, name := range expectedPodNameSplit {
			if !slices.Contains(podNameSplit, name) {
				contains = false
			}
		}
		if contains {
			return &pod
		}
	}
	return nil
}

func readContainerLogsWithRetry(client *kubernetes.Clientset, namespace, podName string, options *v1.PodLogOptions) (string, error) {
	var lastErr error
	for attempt := 1; attempt <= logReadMaxAttempts; attempt++ {
		logs, err := readContainerLogsOnce(client, namespace, podName, options)
		if err == nil {
			return logs, nil
		}
		lastErr = err
		if !isTransientLogReadError(err) || attempt == logReadMaxAttempts {
			break
		}
		time.Sleep(logReadRetryDelay)
	}
	if lastErr == nil {
		lastErr = errors.New("unknown log stream failure")
	}
	return "", lastErr
}

func readContainerLogsOnce(client *kubernetes.Clientset, namespace, podName string, options *v1.PodLogOptions) (string, error) {
	podLogsRequest := client.CoreV1().Pods(namespace).GetLogs(podName, options)
	podLogs, err := podLogsRequest.Stream(context.Background())
	if err != nil {
		return "", err
	}
	if podLogs == nil {
		return "", errors.New("nil pod log stream")
	}
	defer func() {
		if closeErr := podLogs.Close(); closeErr != nil {
			logger.Log("Failed to close pod log reader due to %v", closeErr)
		}
	}()
	buf := new(bytes.Buffer)
	if _, copyErr := io.Copy(buf, podLogs); copyErr != nil {
		return "", copyErr
	}
	return buf.String(), nil
}

func isTransientLogReadError(err error) bool {
	errMsg := err.Error()
	return strings.Contains(errMsg, "PodInitializing") || strings.Contains(errMsg, "is waiting to start")
}

func formatPodStatus(pod *v1.Pod) string {
	containerStates := make([]string, 0, len(pod.Status.ContainerStatuses))
	for _, status := range pod.Status.ContainerStatuses {
		switch {
		case status.State.Waiting != nil:
			containerStates = append(containerStates, fmt.Sprintf("%s=waiting(%s)", status.Name, status.State.Waiting.Reason))
		case status.State.Running != nil:
			containerStates = append(containerStates, fmt.Sprintf("%s=running", status.Name))
		case status.State.Terminated != nil:
			containerStates = append(containerStates, fmt.Sprintf("%s=terminated(%s)", status.Name, status.State.Terminated.Reason))
		}
	}
	return fmt.Sprintf("pod phase=%s, container states=[%s]", pod.Status.Phase, strings.Join(containerStates, ", "))
}

// GetPodContainingContainer - Get the name of the pod with container name containing substring
func GetPodContainingContainer(client *kubernetes.Clientset, namespace, containerName string) *v1.Pod {
	pods, err := client.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		logger.Log("Failed to list pods due to: %v", err)
	}
	for _, pod := range pods.Items {
		for _, container := range pod.Spec.Containers {
			if strings.Contains(container.Name, containerName) {
				return &pod
			}
		}
	}
	return nil
}

func CreateUserToken(client *kubernetes.Clientset, namespace, serviceAccountName string) string {
	// Define TokenRequest
	tokenReq := &authenticationv1.TokenRequest{
		Spec: authenticationv1.TokenRequestSpec{
			Audiences:         []string{"pipelines.kubeflow.org"},                             // Token for Kubernetes API server
			ExpirationSeconds: func(i int64) *int64 { return &i }(int64(time.Hour.Seconds())), // 1-hour expiration
		},
	}

	// Create the token
	tokenResponse, err := client.CoreV1().ServiceAccounts(namespace).CreateToken(context.TODO(), serviceAccountName, tokenReq, metav1.CreateOptions{})
	gomega.Expect(err).ToNot(gomega.HaveOccurred(), fmt.Sprintf("Failed to create service account token for '%s' service account under '%s' namespace", serviceAccountName, namespace))
	if tokenResponse != nil {
		return tokenResponse.Status.Token
	}
	return ""
}

// CreateSecret - Create K8s secret in the provided namespace
func CreateSecret(client *kubernetes.Clientset, namespace string, secret *v1.Secret) {
	_, createErr := client.CoreV1().Secrets(namespace).Create(context.TODO(), secret, metav1.CreateOptions{})
	if createErr == nil {
		logger.Log("%s created", secret.Name)
	} else {
		logger.Log("Looks like %s already exists, because creation failed due to %s", secret.Name, createErr.Error())
		_, getErr := client.CoreV1().Secrets(namespace).Get(context.TODO(), secret.Name, metav1.GetOptions{})
		gomega.Expect(getErr).ToNot(gomega.HaveOccurred(), "Failed to get secret '%s'")
	}
}
