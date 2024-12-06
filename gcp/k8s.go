package main

import (
	"context"
	"encoding/json"
	"fmt"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"time"
)

type K8sFlags struct {
	Service    string
	Namespace  string
	KubeConfig string
}

func (k *K8sFlags) k8sConnectIn() (*kubernetes.Clientset, error) {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if k.KubeConfig != "" {
		fmt.Println("Using kubeconfig")
		config, err = clientcmd.BuildConfigFromFlags("", k.KubeConfig)
	} else {
		fmt.Println("Using in-cluster config")
	}

	if err != nil {
		panic(err.Error())
	}
	// Creates the clientset
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	return clientSet, err
}

func (k *K8sFlags) getLatestEtag(clientSet *kubernetes.Clientset) (string, error) {

	svc, err := clientSet.
		CoreV1().
		Services(k.Namespace).
		Get(context.TODO(), k.Service, metav1.GetOptions{
			TypeMeta:        metav1.TypeMeta{},
			ResourceVersion: "",
		})
	return svc.Annotations["api.cloudflare.eTag"], err

}

func (k *K8sFlags) changeSourceIps(clientSet *kubernetes.Clientset, lbRanges []string, eTag string) error {

	payloadSvc := v1.Service{
		//TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				"api.cloudflare.eTag":       eTag,
				"api.cloudflare.ipv4Cidrs":  fmt.Sprintf("%s", lbRanges),
				"api.cloudflare.lastUpdate": fmt.Sprintf("%s", time.Now()),
			},
		},
		Spec: v1.ServiceSpec{
			LoadBalancerSourceRanges: lbRanges,
		},
	}

	payloadBytes, _ := json.Marshal(payloadSvc)
	_, err := clientSet.
		CoreV1().
		Services(k.Namespace).
		Patch(context.TODO(), k.Service, types.MergePatchType, payloadBytes, metav1.PatchOptions{FieldManager: "JsonPatch", DryRun: []string{}})

	return err
}
