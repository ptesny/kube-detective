package detective

import (
	"fmt"
	"strconv"

	"k8s.io/kubernetes/pkg/api"
)

func (d *Detective) hitServices(sourceHostNetwork, targetHostNetwork bool) {
	for _, service := range d.services {
		for _, pod := range d.pods {
			if sourceHostNetwork == pod.Spec.SecurityContext.HostNetwork {
				if s, err := strconv.ParseBool(service.Labels["hostNetwork"]); err == nil && targetHostNetwork == s {
					d.dialClusterIP(pod, service)
				}
			}
		}
	}
}

func (d *Detective) hitExternalIP(sourceHostNetwork, targetHostNetwork bool) {
	for _, service := range d.services {
		for _, pod := range d.pods {
			if sourceHostNetwork == pod.Spec.SecurityContext.HostNetwork {
				if s, err := strconv.ParseBool(service.Labels["hostNetwork"]); err == nil && targetHostNetwork == s {
					d.dialExternalIP(pod, service)
				}
			}
		}
	}
}

func (d *Detective) hitPods(sourceHostNetwork, targetHostNetwork bool) {
	for _, source := range d.pods {
		for _, target := range d.pods {
			if sourceHostNetwork == source.Spec.SecurityContext.HostNetwork && targetHostNetwork == target.Spec.SecurityContext.HostNetwork {
				d.dialPodIP(source, target)
			}
		}
	}
}

func (d *Detective) dialPodIP(source *api.Pod, target *api.Pod) {
	_, err := d.dial(source, target.Status.PodIP, PodHttpPort)

	result := "success"
	if err != nil {
		result = "failure"
	}

	fmt.Printf("[%v] %30v --> %-30v   %-15v --> %-15v\n",
		result,
		source.Spec.NodeName,
		target.Spec.NodeName,
		source.Status.PodIP,
		target.Status.PodIP,
	)
}

func (d *Detective) dialClusterIP(pod *api.Pod, service *api.Service) {
	_, err := d.dial(pod, service.Spec.ClusterIP, service.Spec.Ports[0].Port)

	result := "success"
	if err != nil {
		result = "failure"
	}

	fmt.Printf("[%v] %30v --> ClusterIP --> %-30v   %-15v --> %-15v --> %-15v\n",
		result,
		pod.Spec.NodeName,
		service.Labels["nodeName"],
		pod.Status.PodIP,
		service.Spec.ClusterIP,
		service.Labels["podIP"],
	)
}

func (d *Detective) dialExternalIP(pod *api.Pod, service *api.Service) {
	_, err := d.dial(pod, service.Spec.ExternalIPs[0], service.Spec.Ports[0].Port)

	result := "success"
	if err != nil {
		result = "failure"
	}

	fmt.Printf("[%v] %30v --> ExternalIP --> %-30v   %-15v --> %-15v --> %-15v\n",
		result,
		pod.Spec.NodeName,
		service.Labels["nodeName"],
		pod.Status.PodIP,
		service.Spec.ExternalIPs[0],
		service.Labels["podIP"],
	)
}

func (d *Detective) dial(pod *api.Pod, host string, port int32) (string, error) {
	cmd := fmt.Sprintf("wget --timeout=1 -O - http://%v:%v", host, port)
	return RunHostCmd(d.namespace.Name, pod.Name, cmd)
}
