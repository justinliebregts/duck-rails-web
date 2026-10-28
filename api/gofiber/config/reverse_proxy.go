package config

import "os"

func Retrievek8sServiceName() string {
	return os.Getenv("K8S_CONFIG_SERVICE_NAME")
}

func Retrievek8sNameSpace() string {
	return os.Getenv("K8S_NAMESPACE")
}

func RetrieveK8sServiceURL() string {
	return os.Getenv("K8S_SERVICE_URL")
}

// func FormatK8sServiceURL(k8sServiceName, k8sNameSpace string) string {
// 	if k8sServiceName == "" {
// 		k8sServiceName = Retrievek8sServiceName()
// 	}
// 	if k8sNameSpace == "" {
// 		k8sNameSpace = Retrievek8sNameSpace()
// 	}

// 	return "http://" + k8sServiceName + "." + k8sNameSpace + ".svc.cluster.local:3000"
// }
