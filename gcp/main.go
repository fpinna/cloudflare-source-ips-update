package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {

	// Flags
	var flagSVC, flagNS, flagKubeConfig string
	flag.StringVar(&flagSVC, "svc", "", "Service")
	flag.StringVar(&flagNS, "ns", "", "Namespace")
	flag.StringVar(&flagKubeConfig, "kubeconfig", "", "KubeConfig")
	flag.Parse()

	if flagSVC == "" || flagNS == "" {
		fmt.Println("-svc and -ns must be defined")
		os.Exit(1)
	}

	// Fill struct flags
	run := K8sFlags{
		Service:    flagSVC,
		Namespace:  flagNS,
		KubeConfig: flagKubeConfig,
	}

	// Get cloudflare ips
	now, err := getJsonIps()
	if err != nil {
		fmt.Println(err)
	}

	//Kubernetes connect
	clientSet, _ := run.k8sConnectIn()

	// Get latest eTag from k8s Service annotation
	last, err := run.getLatestEtag(clientSet)
	if err != nil {
		fmt.Println(err)
	}

	// Compare/Decision
	if now.Result.Etag != last {

		fmt.Sprintf("New eTag: %s\n Old eTag: %s", now.Result.Etag, last)
		// First element in append is vpn source following by gcp health check origins
		err = run.changeSourceIps(clientSet, append(now.Result.Ipv4Cidrs, "x.x.x.x/x"), now.Result.Etag)
		if err != nil {
			panic(err.Error())
		}
		fmt.Println("Service updated")
		fmt.Sprintf("New eTag: %s\n- - -\nIps: %s", now.Result.Etag, now.Result.Ipv4Cidrs)
	} else {
		fmt.Println("Nothing to do")
	}
}
