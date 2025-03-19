package docker

import "github.com/fedstackjs/azukiiro/instancer"

func getProjectNameForTask(task instancer.InstanceTask, config *DockerAdapterConfig) string {
	return "inst-" + task.InstanceId()
}

func getProjectDomainForTask(task instancer.InstanceTask, config *DockerAdapterConfig) string {
	return task.InstanceId() + config.DomainSuffix
}
