package kube

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
	"path/filepath"
	"sync"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var clientset *kubernetes.Clientset

var once sync.Once

func ClusterNodenames() []string {
	once.Do(func() {
		home := homeDir()
		kubeconfig := filepath.Join(home, ".kube", "config")
		config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			log.Fatalln("'~/.kube/config' not exists!!!")
		}

		clientset, err = kubernetes.NewForConfig(config)
		if err != nil{
			log.Fatalln("kubernetes client create fail!!!")
		}
	})

	nodes, err := clientset.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil{
		log.Fatalln("get kubernetes nodes fail!!!")
	}

	nodeItems := nodes.Items

	res := make([]string, 0, len(nodeItems))
	for i := range nodeItems{
		res = append(res, nodeItems[i].Name)
	}

	return res
}




func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
