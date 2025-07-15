.package main

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {
	// Build kubeconfig path
	var kubeconfig string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = filepath.Join(home, ".kube", "config")
	}

	// Create the clientset
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatalf("Error building kubeconfig: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating kubernetes client: %v", err)
	}

	ctx := context.Background()

	// Get all deployments across all namespaces
	deployments, err := clientset.AppsV1().Deployments("").List(ctx, metav1.ListOptions{})
	if err != nil {
		log.Fatalf("Error listing deployments: %v", err)
	}

	// Get all statefulsets across all namespaces
	statefulsets, err := clientset.AppsV1().StatefulSets("").List(ctx, metav1.ListOptions{})
	if err != nil {
		log.Fatalf("Error listing statefulsets: %v", err)
	}

	// Get all daemonsets across all namespaces
	daemonsets, err := clientset.AppsV1().DaemonSets("").List(ctx, metav1.ListOptions{})
	if err != nil {
		log.Fatalf("Error listing daemonsets: %v", err)
	}

	restarted := 0

	// Process deployments
	for _, deployment := range deployments.Items {
		if strings.Contains(strings.ToLower(deployment.Name), "database") {
			err := restartDeployment(ctx, clientset, deployment.Namespace, deployment.Name)
			if err != nil {
				log.Printf("Error restarting deployment %s/%s: %v", deployment.Namespace, deployment.Name, err)
			} else {
				fmt.Printf("Successfully restarted deployment: %s/%s\n", deployment.Namespace, deployment.Name)
				restarted++
			}
		}
	}

	// Process statefulsets
	for _, statefulset := range statefulsets.Items {
		if strings.Contains(strings.ToLower(statefulset.Name), "database") {
			err := restartStatefulSet(ctx, clientset, statefulset.Namespace, statefulset.Name)
			if err != nil {
				log.Printf("Error restarting statefulset %s/%s: %v", statefulset.Namespace, statefulset.Name, err)
			} else {
				fmt.Printf("Successfully restarted statefulset: %s/%s\n", statefulset.Namespace, statefulset.Name)
				restarted++
			}
		}
	}

	// Process daemonsets
	for _, daemonset := range daemonsets.Items {
		if strings.Contains(strings.ToLower(daemonset.Name), "database") {
			err := restartDaemonSet(ctx, clientset, daemonset.Namespace, daemonset.Name)
			if err != nil {
				log.Printf("Error restarting daemonset %s/%s: %v", daemonset.Namespace, daemonset.Name, err)
			} else {
				fmt.Printf("Successfully restarted daemonset: %s/%s\n", daemonset.Namespace, daemonset.Name)
				restarted++
			}
		}
	}

	fmt.Printf("\nTotal resources restarted: %d\n", restarted)
}

func restartDeployment(ctx context.Context, clientset *kubernetes.Clientset, namespace, name string) error {
	deployment, err := clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get deployment: %w", err)
	}

	// Add restart annotation to trigger rollout
	if deployment.Spec.Template.Annotations == nil {
		deployment.Spec.Template.Annotations = make(map[string]string)
	}
	deployment.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)

	_, err = clientset.AppsV1().Deployments(namespace).Update(ctx, deployment, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update deployment: %w", err)
	}

	return nil
}

func restartStatefulSet(ctx context.Context, clientset *kubernetes.Clientset, namespace, name string) error {
	statefulset, err := clientset.AppsV1().StatefulSets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get statefulset: %w", err)
	}

	// Add restart annotation to trigger rollout
	if statefulset.Spec.Template.Annotations == nil {
		statefulset.Spec.Template.Annotations = make(map[string]string)
	}
	statefulset.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)

	_, err = clientset.AppsV1().StatefulSets(namespace).Update(ctx, statefulset, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update statefulset: %w", err)
	}

	return nil
}

func restartDaemonSet(ctx context.Context, clientset *kubernetes.Clientset, namespace, name string) error {
	daemonset, err := clientset.AppsV1().DaemonSets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get daemonset: %w", err)
	}

	// Add restart annotation to trigger rollout
	if daemonset.Spec.Template.Annotations == nil {
		daemonset.Spec.Template.Annotations = make(map[string]string)
	}
	daemonset.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)

	_, err = clientset.AppsV1().DaemonSets(namespace).Update(ctx, daemonset, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update daemonset: %w", err)
	}

	return nil
}