package helpers

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// K8sClient provides methods to interact with Kubernetes
type K8sClient struct {
	clientset     *kubernetes.Clientset
	dynamicClient dynamic.Interface
}

// CRD Group/Version/Resource definitions for Lissto
var (
	BlueprintGVR = schema.GroupVersionResource{
		Group:    "env.lissto.dev",
		Version:  "v1alpha1",
		Resource: "blueprints",
	}
	StackGVR = schema.GroupVersionResource{
		Group:    "env.lissto.dev",
		Version:  "v1alpha1",
		Resource: "stacks",
	}
	EnvGVR = schema.GroupVersionResource{
		Group:    "env.lissto.dev",
		Version:  "v1alpha1",
		Resource: "envs",
	}
)

// NewK8sClient creates a new Kubernetes client
func NewK8sClient() (*K8sClient, error) {
	// Build config from kubeconfig
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		kubeconfig = filepath.Join(os.Getenv("HOME"), ".kube", "config")
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to build kubeconfig: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %w", err)
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}

	return &K8sClient{
		clientset:     clientset,
		dynamicClient: dynamicClient,
	}, nil
}

// GetBlueprint retrieves a Blueprint CRD
func (k *K8sClient) GetBlueprint(ctx context.Context, namespace, name string) (*unstructured.Unstructured, error) {
	return k.dynamicClient.Resource(BlueprintGVR).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
}

// ListBlueprints lists all Blueprints in a namespace
func (k *K8sClient) ListBlueprints(ctx context.Context, namespace string) (*unstructured.UnstructuredList, error) {
	return k.dynamicClient.Resource(BlueprintGVR).Namespace(namespace).List(ctx, metav1.ListOptions{})
}

// BlueprintExists checks if a Blueprint exists
func (k *K8sClient) BlueprintExists(ctx context.Context, namespace, name string) bool {
	_, err := k.GetBlueprint(ctx, namespace, name)
	return err == nil
}

// GetStack retrieves a Stack CRD
func (k *K8sClient) GetStack(ctx context.Context, namespace, name string) (*unstructured.Unstructured, error) {
	return k.dynamicClient.Resource(StackGVR).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
}

// ListStacks lists all Stacks in a namespace
func (k *K8sClient) ListStacks(ctx context.Context, namespace string) (*unstructured.UnstructuredList, error) {
	return k.dynamicClient.Resource(StackGVR).Namespace(namespace).List(ctx, metav1.ListOptions{})
}

// StackExists checks if a Stack exists
func (k *K8sClient) StackExists(ctx context.Context, namespace, name string) bool {
	_, err := k.GetStack(ctx, namespace, name)
	return err == nil
}

// GetDeployment retrieves a Deployment
func (k *K8sClient) GetDeployment(ctx context.Context, namespace, name string) (*appsv1.Deployment, error) {
	return k.clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
}

// DeploymentExists checks if a Deployment exists
func (k *K8sClient) DeploymentExists(ctx context.Context, namespace, name string) bool {
	_, err := k.GetDeployment(ctx, namespace, name)
	return err == nil
}

// DeploymentReady checks if a Deployment is ready
func (k *K8sClient) DeploymentReady(ctx context.Context, namespace, name string) bool {
	deploy, err := k.GetDeployment(ctx, namespace, name)
	if err != nil {
		return false
	}
	return deploy.Status.ReadyReplicas == *deploy.Spec.Replicas
}

// GetDeploymentImage gets the image of the first container in a Deployment
func (k *K8sClient) GetDeploymentImage(ctx context.Context, namespace, name string) (string, error) {
	deploy, err := k.GetDeployment(ctx, namespace, name)
	if err != nil {
		return "", err
	}
	if len(deploy.Spec.Template.Spec.Containers) == 0 {
		return "", fmt.Errorf("deployment has no containers")
	}
	return deploy.Spec.Template.Spec.Containers[0].Image, nil
}

// GetService retrieves a Service
func (k *K8sClient) GetService(ctx context.Context, namespace, name string) (*corev1.Service, error) {
	return k.clientset.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
}

// ServiceExists checks if a Service exists
func (k *K8sClient) ServiceExists(ctx context.Context, namespace, name string) bool {
	_, err := k.GetService(ctx, namespace, name)
	return err == nil
}

// GetConfigMap retrieves a ConfigMap
func (k *K8sClient) GetConfigMap(ctx context.Context, namespace, name string) (*corev1.ConfigMap, error) {
	return k.clientset.CoreV1().ConfigMaps(namespace).Get(ctx, name, metav1.GetOptions{})
}

// ConfigMapExists checks if a ConfigMap exists
func (k *K8sClient) ConfigMapExists(ctx context.Context, namespace, name string) bool {
	_, err := k.GetConfigMap(ctx, namespace, name)
	return err == nil
}

// ListPods lists pods with optional label selector
func (k *K8sClient) ListPods(ctx context.Context, namespace, labelSelector string) (*corev1.PodList, error) {
	return k.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
}

// NamespaceExists checks if a namespace exists
func (k *K8sClient) NamespaceExists(ctx context.Context, name string) bool {
	_, err := k.clientset.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
	return err == nil
}

// WaitForResource waits for a resource to exist
func (k *K8sClient) WaitForResource(ctx context.Context, gvr schema.GroupVersionResource, namespace, name string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		_, err := k.dynamicClient.Resource(gvr).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
		if err == nil {
			return nil
		}
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("timeout waiting for %s/%s in namespace %s", gvr.Resource, name, namespace)
}

// WaitForResourceDeletion waits for a resource to be deleted
func (k *K8sClient) WaitForResourceDeletion(ctx context.Context, gvr schema.GroupVersionResource, namespace, name string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		_, err := k.dynamicClient.Resource(gvr).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return nil // Resource not found, deletion complete
		}
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("timeout waiting for deletion of %s/%s in namespace %s", gvr.Resource, name, namespace)
}

// GetBlueprintAnnotation gets an annotation from a Blueprint
func (k *K8sClient) GetBlueprintAnnotation(ctx context.Context, namespace, name, key string) (string, error) {
	bp, err := k.GetBlueprint(ctx, namespace, name)
	if err != nil {
		return "", err
	}
	annotations := bp.GetAnnotations()
	if annotations == nil {
		return "", nil
	}
	return annotations[key], nil
}

// GetStackPhase gets the phase from a Stack status
func (k *K8sClient) GetStackPhase(ctx context.Context, namespace, name string) (string, error) {
	stack, err := k.GetStack(ctx, namespace, name)
	if err != nil {
		return "", err
	}
	status, found, err := unstructured.NestedMap(stack.Object, "status")
	if err != nil || !found {
		return "", nil
	}
	phase, _, _ := unstructured.NestedString(status, "phase")
	return phase, nil
}
