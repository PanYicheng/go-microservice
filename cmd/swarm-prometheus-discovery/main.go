package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/alexflint/go-arg"
	"github.com/docker/docker/api/types"
	// "github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var cfg struct {
	NetworkName       string `default:"my_network" arg:"env:NETWORK_NAME" help:"Specify the name of the network you want to scrape"`
	IgnoredServiceStr string `default:"prometheus,grafana,swarm-prometheus-discovery,zipkin,rabbitmq" arg:"env:IGNORED_SERVICE_STR" help:"Comma-separated list of service names we do not want to scrape"`
	Interval          int `default:"1" arg:"env" help:"Update interval of service discovery"`
}

var networkID = ""
var ignoredServiceIDs = make([]string, 0)

func main() {
	arg.MustParse(&cfg)

	logrus.SetLevel(logrus.TraceLevel)
	logrus.Infof("Network name: %v\n", cfg.NetworkName)
	logrus.Println("Starting Swarm-scraper!")

	// Connect to the Docker API
	cli, err := client.NewClientWithOpts(client.WithHost("unix:///var/run/docker.sock"), client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	// Find the networkID we want to address tasks on.
	findNetworkId(cli, cfg.NetworkName)
	// Parse any services we don't want to scrape (such as ourselves and prometheus server)
	// into a slice. Also translate to service IDs.
	parseIgnoredServices(cli, cfg.IgnoredServiceStr)

	// Start the task poller
	go func(cli *client.Client) {
		for {
			time.Sleep(time.Second * time.Duration(cfg.Interval))
			pollTasks(cli)
		}
	}(cli)

	// Block...
	log.Println("Waiting at block...")

	wg := sync.WaitGroup{} // Use a WaitGroup to block main() exit
	wg.Add(1)
	wg.Wait()
}

func findNetworkId(cli *client.Client, networkName string) {
	networkList, err := cli.NetworkList(context.Background(), types.NetworkListOptions{})
	if err != nil {
		fmt.Println(err.Error())
		panic(err)
	}
	for _, n := range networkList {
		entry := logrus.StandardLogger().WithFields(logrus.Fields{
			"name": n.Name,
			"ID":   n.ID})
		if n.Name == networkName {
			networkID = n.ID
			entry.Debug("Found network ID")
			return
		}
	}
	logrus.Errorf("Could not find NetworkID of %v, will assume 'ingress'\n", networkName)
	for _, n := range networkList {
		if n.Name == "ingress" {
			networkID = n.ID
			return
		}
	}
	panic("Could neither resolve network " + networkName + " nor ingress network, panic!")
}

func parseIgnoredServices(cli *client.Client, ignoredServicesStr string) {
	var ignoredServices []string
	if strings.Contains(ignoredServicesStr, ",") {
		ignoredServices = strings.Split(ignoredServicesStr, ",")
	} else {
		ignoredServices = append(ignoredServices, ignoredServicesStr)
	}
	srvs, err := cli.ServiceList(context.Background(), types.ServiceListOptions{})
	if err != nil {
		panic(err)
	}
	for _, s := range srvs {
		if isInList(ignoredServices, s.Spec.Annotations.Name) {
			ignoredServiceIDs = append(ignoredServiceIDs, s.ID)
			logrus.Debugf("Ignored service name: %s, ID:%s\n", s.Spec.Annotations.Name, s.ID[:8])
		}
	}
}

// pollTasks search tasks deployed in swarm to find those on our network.
func pollTasks(cli *client.Client) {

	ctx := context.Background()
	//filter := filters.NewArgs(filters.Arg("desired-state", "running"))
	// tasks, err := cli.TaskList(context.Background(), types.TaskListOptions{})
	// if err != nil {
	// 	panic(err)
	// }
	taskMap := make(map[string]*ScrapedTask)
	// for _, t := range tasks {
	// 	entry := logrus.WithFields(logrus.Fields{
	// 		"TaskID":      t.ID[:8],
	// 		"Name":        t.Name,
	// 		"ServiceID":   t.ServiceID[:8],
	// 		"ContainerID": t.Status.ContainerStatus.ContainerID[:8]})

	// 	// Skip if the service holding the task is in ignoredList, e.g. don't scrape prometheus...
	// 	if isInList(ignoredServiceIDs, t.ServiceID) {
	// 		entry.Debug("task is ignored.")
	// 		continue
	// 	}
		// logrus.Debug("Networks\n")
		// for _, net := range t.Spec.Networks {
		// 	logrus.Debugf("Target: %s\n", net.Target)
		// }

		// Find HTTP port?
		// portNumber := -1
		// for _, pc := range t.Status.PortStatus.Ports {
		// 	logrus.Debugf("Name:%s, Prot:%s, Target:%s, Published:%s\n", pc.Name, pc.Protocol, pc.TargetPort, pc.PublishedPort)
		//     if pc.Protocol == "tcp" {
		//         portNumber = fmt.Sprint(pc.PublishedPort)
		// 		entry.WithField("Port", portNumber)
		// 		break
		//     }
		// }
	//	containerJSON, err := cli.ContainerInspect(ctx, t.Status.ContainerStatus.ContainerID)
	//	if err != nil {
	//		panic(err)
	//	}
	//	ports := make([]int, 0)
	//	for port, _ := range containerJSON.NetworkSettings.Ports {
	//		ports = append(ports, port.Int())
	//	}
	//	entry = entry.WithField("Ports", ports)
	//	nets := make([]string, 0)
	//	for net, _ := range containerJSON.NetworkSettings.Networks {
	//		nets = append(nets, net)
	//	}
	//	entry = entry.WithField("Networks", nets)

	//	// Skip if no exposed tcp port
	//	var portNumber string
	//	if len(ports) == 0 {
	//		entry.Debug("task has no tcp port.")
	//		continue
	//	} else {
	//		portNumber = strconv.Itoa(ports[0])
	//		entry.Debug("processing task.")
	//	}

	//	// Iterate network attachments on task
	//	for _, neta := range t.NetworksAttachments {
	//		// Only extract IP if on expected network.
	//		if neta.Network.ID == networkID {
	//			if taskEntry, ok := taskMap[t.ID]; ok {
	//				processExistingTask(taskEntry, neta, portNumber, &t)
	//			} else {
	//				processNewTask(neta, portNumber, &t, taskMap)
	//			}
	//		}
	//	}
	//}

	svcList, err := cli.ServiceList(ctx, types.ServiceListOptions{})
	if err != nil {
		logrus.Error(err)
	}
	var svcAddrMap = make(map[string]string)
	for _, svc := range svcList {
		fmt.Printf("Service ID:%s Name:%s\n", svc.ID[:8], svc.Spec.Name)
		// Ignore some services
		if isInList(ignoredServiceIDs, svc.ID) {
			fmt.Println("Ignored.")
			continue
		}
		// Resolve service IP within provided network.
		for _, vip := range svc.Endpoint.VirtualIPs {
			// fmt.Printf("Port:%d PubPort:%d\n", port.TargetPort, port.PublishedPort)
			fmt.Printf("NetworkID:%s, Addr:%s\n", vip.NetworkID, vip.Addr)
			if vip.NetworkID == networkID {
				svcAddrMap[svc.Spec.Name] = vip.Addr
				break
			}
		}
		taskMap[svc.Spec.Name] = &ScrapedTask{
			Targets: []string{formatIp(svcAddrMap[svc.Spec.Name], strconv.Itoa(6767))},
			Labels: make(map[string]string)}
	}
	// Transform values of map into slice.
	taskList := make([]ScrapedTask, 0)
	for _, value := range taskMap {
		taskList = append(taskList, *value)
	}

	// Write config file
	bytes, err := json.Marshal(taskList)
	if err != nil {
		panic(err)
	}
	logrus.Info(string(bytes))

	file, err := os.Create("/etc/swarm-endpoints/swarm-endpoints.json")
	if err != nil {
		logrus.Errorf("Error writing file: %v\n", err.Error())
	} else {
		file.Write(bytes)
		file.Close()
	}
}

func processNewTask(neta swarm.NetworkAttachment, portNumber string, t *swarm.Task, taskMap map[string]*ScrapedTask) {
	// New service
	taskEntry := ScrapedTask{Targets: make([]string, 0), Labels: make(map[string]string)}
	for _, adr := range neta.Addresses {
		taskEntry.Targets = append(taskEntry.Targets, formatIp(adr, portNumber))
	}
	taskEntry.Labels["Name"] = t.Annotations.Name
	taskMap[t.ID] = &taskEntry
}

func processExistingTask(taskEntry *ScrapedTask, neta swarm.NetworkAttachment, portNumber string, t *swarm.Task) {
	// Existing task
	localTargets := make([]string, len(taskEntry.Targets))
	copy(localTargets, taskEntry.Targets)
	for _, adr := range neta.Addresses {
		localTargets = append(localTargets, formatIp(adr, portNumber))
	}
	taskEntry.Targets = localTargets
	taskEntry.Labels["Name"] = t.Annotations.Name
}

func isInList(list []string, s string) bool {
	for _, ignored := range list {
		if ignored == s {
			return true
		}
	}
	return false
}

func formatIp(ip string, port string) string {
	// Remove /NN part of ip
	index := strings.Index(ip, "/")
	ip = ip[:index] + ":" + port
	return ip
}

type ScrapedTask struct {
	Targets []string          `json:"targets"`
	Labels  map[string]string `json:"labels"`
}
