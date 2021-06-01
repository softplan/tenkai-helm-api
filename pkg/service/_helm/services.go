package helmapi

import (
	"fmt"
	model2 "github.com/softplan/tenkai-helm-api/pkg/dbms/model"
	"strings"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

const loadBalancerWidth = 16

//GetServices - Get Service information
func (svc HelmServiceImpl) GetServices(kubeconfig string, namespace string) ([]model2.Service, error) {
	services := make([]model2.Service, 0)
	_, client, err := svc.GetHelmConnection().GetKubeClient("", kubeconfig)
	if err != nil {
		return services, err
	}
	opts := &metav1.ListOptions{}
	list, err := client.CoreV1().Services(namespace).List(*opts)
	if err != nil {
		return services, err
	}
	var service *model2.Service
	for _, element := range list.Items {
		service = fillService(element)
		services = append(services, *service)
	}
	return services, nil
}

func makePortString(ports []v1.ServicePort) string {
	pieces := make([]string, len(ports))
	for ix := range ports {
		port := &ports[ix]
		pieces[ix] = fmt.Sprintf("%d/%s", port.Port, port.Protocol)
		if port.NodePort > 0 {
			pieces[ix] = fmt.Sprintf("%d:%d/%s", port.Port, port.NodePort, port.Protocol)
		}
	}
	return strings.Join(pieces, ",")
}

func fillService(service v1.Service) *model2.Service {
	result := model2.Service{Name: service.Name}
	result.ClusterIP = service.Spec.ClusterIP
	result.ExternalIP = getServiceExternalIP(&service, false)
	result.Ports = makePortString(service.Spec.Ports)
	result.Type = string(service.Spec.Type)
	result.Age = translateTimestampSince(service.CreationTimestamp)
	return &result
}

func checkIps(svc *v1.Service) string {
	if len(svc.Spec.ExternalIPs) > 0 {
		return strings.Join(svc.Spec.ExternalIPs, ",")
	}
	return "<none>"
}

func getServiceExternalIP(svc *v1.Service, wide bool) string {
	switch svc.Spec.Type {
	case v1.ServiceTypeClusterIP, v1.ServiceTypeNodePort:
		return checkIps(svc)
	case v1.ServiceTypeLoadBalancer:
		lbIps := loadBalancerStatusStringer(svc.Status.LoadBalancer, wide)
		if len(svc.Spec.ExternalIPs) > 0 {
			results := []string{}
			if len(lbIps) > 0 {
				results = append(results, strings.Split(lbIps, ",")...)
			}
			results = append(results, svc.Spec.ExternalIPs...)
			return strings.Join(results, ",")
		}
		if len(lbIps) > 0 {
			return lbIps
		}
		return "<pending>"
	case v1.ServiceTypeExternalName:
		return svc.Spec.ExternalName
	}
	return "<unknown>"
}

func loadBalancerStatusStringer(s v1.LoadBalancerStatus, wide bool) string {
	ingress := s.Ingress
	result := sets.NewString()
	for i := range ingress {
		if ingress[i].IP != "" {
			result.Insert(ingress[i].IP)
		} else if ingress[i].Hostname != "" {
			result.Insert(ingress[i].Hostname)
		}
	}

	r := strings.Join(result.List(), ",")
	if !wide && len(r) > loadBalancerWidth {
		r = r[0:(loadBalancerWidth-3)] + "..."
	}
	return r
}
