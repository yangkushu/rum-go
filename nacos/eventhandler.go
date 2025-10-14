package nacos

import (
	"fmt"
	"github.com/nacos-group/nacos-sdk-go/model"
	"github.com/yangkushu/rum-go/log"
)

// 服务监控回调对象定义
type ClusterMonitor interface {
	OnAddInstances(servicename string, list []InstanceInfo)
	OnDelInstances(servicename string, list []InstanceInfo)
}

func createEventHandler(serviceName, groupName string, service *ClusterService) *EventHandler {
	return &EventHandler{
		block:       false,
		serviceName: serviceName,
		groupName:   groupName,
		monitors:    make([]ClusterMonitor, 0),
		instances:   make(map[string]InstanceInfo),
		service:     service,
	}
}

type EventHandler struct {
	block       bool
	serviceName string
	groupName   string
	monitors    []ClusterMonitor
	// 所有实例列表  {instanceID: InstanceInfo}
	instances map[string]InstanceInfo
	service   *ClusterService
}

func (h *EventHandler) addMonitor(monitor ClusterMonitor) {
	h.monitors = append(h.monitors, monitor)
}

func (h *EventHandler) setBlock(block bool) {
	h.block = block
}

func (h *EventHandler) onEvent(services []model.SubscribeService, err error) {
	if h.block {
		return
	}

	defer func() {
		if p := recover(); p != nil {
			log.Error("nacos onEvent catch panics",
				log.String("recover", fmt.Sprint(p)),
				log.String("error", err.Error()),
				log.String("services", fmt.Sprintf("%v", services)),
			)
		}
	}()

	addList := []InstanceInfo{}
	removeList := []InstanceInfo{}

	oldlist := h.instances
	newList := make(map[string]InstanceInfo)

	instanceList := make([]InstanceInfo, 0)

	for _, service := range services {
		info := getInstanceInfo(h.groupName, &service)
		instanceList = append(instanceList, info)
		newList[service.InstanceId] = info
		if _, ok := oldlist[service.InstanceId]; !ok {
			addList = append(addList, info)
		} else {
			delete(oldlist, service.InstanceId)
		}
	}

	h.instances = newList
	if nil != h.service {
		h.service.updateInstanceList(h.serviceName, instanceList)
	}

	if len(oldlist) > 0 {
		for _, service := range oldlist {
			removeList = append(removeList, service)
		}
	}

	if len(addList) > 0 {
		for _, monitor := range h.monitors {
			monitor.OnAddInstances(h.serviceName, addList)
		}
	}

	if len(removeList) > 0 {
		for _, monitor := range h.monitors {
			monitor.OnDelInstances(h.serviceName, removeList)
		}
	}
}
