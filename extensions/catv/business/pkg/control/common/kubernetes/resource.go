package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"

	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	clientset *kubernetes.Clientset
	once      sync.Once
	initErr   error
)

// getClientset returns a singleton Kubernetes clientset configured for in-cluster access.
// The clientset is initialized once and reused for all subsequent calls.
//
// Returns:
//   - *kubernetes.Clientset: configured Kubernetes client
//   - error: if in-cluster config fails or clientset creation fails
func getClientset() (*kubernetes.Clientset, error) {
	once.Do(func() {
		config, err := rest.InClusterConfig()
		if err != nil {
			initErr = fmt.Errorf("failed to get in-cluster config: %w", err)
			return
		}
		clientset, initErr = kubernetes.NewForConfig(config)
		if initErr != nil {
			initErr = fmt.Errorf("failed to create clientset: %w", initErr)
		}
	})
	return clientset, initErr
}

// do executes a Kubernetes REST API request with the specified parameters.
//
// Parameters:
//   - ctx: context for request cancellation and timeout
//   - method: HTTP method (GET, POST, PUT, DELETE)
//   - group: API group (empty for core API)
//   - version: API version (e.g., "v1", "v1beta1")
//   - namespace: Kubernetes namespace (empty for cluster-scoped resources)
//   - resource: resource type (e.g., "pods", "jobs")
//   - name: resource name (empty for list operations)
//   - object: request body object (nil for GET/DELETE)
//
// Returns:
//   - string: response body as string
//   - error: if request fails at any stage
func do(ctx context.Context, method, group, version, namespace, resource, name string, object runtime.Object) (string, error) {
	// Validate required parameters (before expensive operations)
	if version == "" {
		return "", fmt.Errorf("version is required")
	}
	if resource == "" && name == "" {
		return "", fmt.Errorf("either resource or name must be specified")
	}

	// Validate HTTP method early
	validMethods := map[string]bool{
		http.MethodGet:    true,
		http.MethodPost:   true,
		http.MethodPut:    true,
		http.MethodDelete: true,
	}
	if !validMethods[method] {
		return "", fmt.Errorf("unsupported HTTP method: %s", method)
	}

	clientset, err := getClientset()
	if err != nil {
		return "", err
	}

	var request *rest.Request
	switch method {
	case http.MethodGet:
		request = clientset.RESTClient().Get()
	case http.MethodPost:
		request = clientset.RESTClient().Post()
	case http.MethodPut:
		request = clientset.RESTClient().Put()
	case http.MethodDelete:
		request = clientset.RESTClient().Delete()
	}

	// Use Prefix instead of AbsPath to allow Namespace(), Resource(), Name() to work
	if group != "" {
		request.Prefix("apis", group, version)
	} else {
		request.Prefix("api", version)
	}

	if namespace != "" {
		request.Namespace(namespace)
	}

	if resource != "" {
		request.Resource(resource)
	}

	if name != "" {
		request.Name(name)
	}

	if object != nil {
		body, err := json.Marshal(object)
		if err != nil {
			return "", fmt.Errorf("failed to marshal object: %w", err)
		}
		request.Body(body)
	}

	slog.Debug("Kubernetes API request",
		"method", method,
		"resource", resource,
		"namespace", namespace,
		"name", name)

	result := request.Do(ctx)

	body, err := result.Raw()
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resultErr := result.Error(); resultErr != nil {
		slog.Error("Kubernetes API request failed",
			"method", method,
			"resource", resource,
			"error", resultErr)
		return string(body), resultErr
	}

	return string(body), nil
}

// Get retrieves a Kubernetes resource and unmarshals it into the specified type.
//
// Type parameter:
//   - T: target type for unmarshaling (e.g., batchV1.Job, coreV1.Pod)
//
// Parameters:
//   - ctx: context for request cancellation and timeout
//   - group: API group (empty for core API)
//   - version: API version
//   - namespace: resource namespace
//   - resource: resource type
//   - name: resource name
//
// Returns:
//   - T: unmarshaled resource object
//   - error: if request fails or unmarshal fails
//
// Example:
//
//	job, err := Get[batchV1.Job](ctx, "batch", "v1", "default", "jobs", "my-job")
func Get[T any](ctx context.Context, group, version, namespace, resource, name string) (T, error) {
	var zero T

	body, err := do(ctx, http.MethodGet, group, version, namespace, resource, name, nil)
	if err != nil {
		return zero, fmt.Errorf("failed to get resource: %w", err)
	}

	var result T
	if err := json.Unmarshal([]byte(body), &result); err != nil {
		return zero, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return result, nil
}

// Post creates a new Kubernetes resource.
//
// Parameters:
//   - ctx: context for request cancellation and timeout
//   - group: API group (e.g., "batch" for Jobs)
//   - version: API version (e.g., "v1")
//   - namespace: target namespace
//   - resource: resource type (e.g., "jobs")
//   - object: resource object to create
//
// Returns:
//   - error: if creation fails
//
// Example:
//
//	err := Post(ctx, "batch", "v1", "default", "jobs", &jobSpec)
func Post(ctx context.Context, group, version, namespace, resource string, object runtime.Object) error {
	_, err := do(ctx, http.MethodPost, group, version, namespace, resource, "", object)
	if err != nil {
		return fmt.Errorf("failed to create resource: %w", err)
	}
	return nil
}

// Put updates an existing Kubernetes resource.
//
// Parameters:
//   - ctx: context for request cancellation and timeout
//   - group: API group
//   - version: API version
//   - namespace: resource namespace
//   - resource: resource type
//   - name: resource name to update
//   - object: updated resource object
//
// Returns:
//   - error: if update fails
func Put(ctx context.Context, group, version, namespace, resource, name string, object runtime.Object) error {
	_, err := do(ctx, http.MethodPut, group, version, namespace, resource, name, object)
	if err != nil {
		return fmt.Errorf("failed to update resource: %w", err)
	}
	return nil
}

// Delete removes a Kubernetes resource.
//
// Parameters:
//   - ctx: context for request cancellation and timeout
//   - group: API group
//   - version: API version
//   - namespace: resource namespace
//   - resource: resource type
//   - name: resource name to delete
//
// Returns:
//   - error: if deletion fails
func Delete(ctx context.Context, group, version, namespace, resource, name string) error {
	_, err := do(ctx, http.MethodDelete, group, version, namespace, resource, name, nil)
	if err != nil {
		return fmt.Errorf("failed to delete resource: %w", err)
	}
	return nil
}

// DeleteWithDeleteOptions removes a Kubernetes resource with custom delete options.
//
// Parameters:
//   - ctx: context for request cancellation and timeout
//   - group: API group
//   - version: API version
//   - namespace: resource namespace
//   - resource: resource type
//   - name: resource name to delete
//   - deleteOptions: options controlling deletion behavior (e.g., grace period, propagation policy)
//
// Returns:
//   - error: if deletion fails
//
// Example:
//
//	deleteOpts := metaV1.DeleteOptions{
//	    GracePeriodSeconds: &gracePeriod,
//	    PropagationPolicy: &propagationPolicy,
//	}
//	err := DeleteWithDeleteOptions(ctx, "batch", "v1", "default", "jobs", "my-job", deleteOpts)
func DeleteWithDeleteOptions(ctx context.Context, group, version, namespace, resource, name string, deleteOptions metaV1.DeleteOptions) error {
	_, err := do(ctx, http.MethodDelete, group, version, namespace, resource, name, &deleteOptions)
	if err != nil {
		return fmt.Errorf("failed to delete resource with options: %w", err)
	}
	return nil
}
