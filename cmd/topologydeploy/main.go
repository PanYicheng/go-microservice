package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"context"
	"path/filepath"
	"github.com/docker/docker/client"
	"github.com/docker/docker/api/types"
	"github.com/alexflint/go-arg"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

type Service struct {
	Name         string
	Port         string
	Concurrency  bool
	CallList     []string
	ResponseTime float64
	Instances    int
}

type CircuitInfo struct {
	Name                   string
	Timeout                int
	MaxConcurrentRequests  int
	SleepWindow            int
	ErrorPercentThreshold  int
	RequestVolumeThreshold int
}

var (
	deployFd *os.File
)

var cfg struct {
	ConfigFile string `arg:"env:CONF_FILE" default:"../../deployments/customtopology/config.json" help:"service topology file name"`
	ScriptDir string `arg:"env:SCRIPT_DIR" default:"../../scripts"`
	PublishedService string `arg:"env" default:"servicea"`
}

func main() {
	logrus.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: "2006-01-02T15:04:05.000",
		FullTimestamp:   true,
	})
	// Read program configs from command line or environment
	arg.MustParse(&cfg)

	// Open deploy scripts file for writing. 
	var err error
	deployFd, err = os.OpenFile(filepath.Join(cfg.ScriptDir, "deploy_customservices.sh"),
		os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0744)
	if err != nil {
		logrus.Fatal(err)
	}
	defer deployFd.Close()

	// Read service configs from config file.
	newServices, err := parseServices(cfg.ConfigFile)
	if err != nil {
		logrus.Fatal(err)
	}
	var	serviceNames []string
	for _, s := range newServices {
		serviceNames = append(serviceNames, s.Name)
	}
	logrus.WithField("ServiceNames", serviceNames).Info("services from config file")

	// Try locate reusable services from Docker and backup config file.
	backupServices, err := parseServices(filepath.Join("../../deployments/customtopology", "backup.json"))
	if err != nil {
		logrus.Error(err)
	}

	reuseServices := reuse(backupServices, newServices)
	reuseServiceMap := make(map[string]*Service)
	serviceNames = nil
	for _, s := range reuseServices {
		serviceNames = append(serviceNames, s.Name)
		reuseServiceMap[s.Name] = &s
	}
	logrus.WithField("ServiceNames", serviceNames).Infoln("reusable services from docker")

	var newCreateServices []Service
	serviceNames = nil
	for _, newservice := range newServices {
		if _, ok := reuseServiceMap[newservice.Name]; !ok {
			newCreateServices = append(newCreateServices, newservice)
			serviceNames = append(serviceNames, newservice.Name)
		}
	}
	logrus.WithField("ServiceNames", serviceNames).Infoln("new services required deploying")

	generateCode("../../deployments/customtopology", newCreateServices)

	copyFile(cfg.ConfigFile, filepath.Join("../../deployments/customtopology", "backup.json"))
}

// copyFile copies file in src to dst path.
func copyFile(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

// reuse helps reuse existing services. It returns services that are running
// in the docker swarm.
func reuse(backupServices, newServices []Service) []Service {
	// currently disable it.
	return nil

	// 从 backup.json 中读取的 service 的信息
	// backupServices, err := parseServices("backup.json")
	// if err != nil {
	// 	return nil
	// }

	endpoint := "unix:///var/run/docker.sock"
	dockerClient, err := client.NewClientWithOpts(client.WithHost(endpoint), client.WithAPIVersionNegotiation())
	if err != nil {
		logrus.Errorln(err.Error())
		return nil
	}

	services, err := dockerClient.ServiceList(context.Background(), types.ServiceListOptions{})
	if err != nil {
		logrus.Errorln(err.Error())
		return nil
	}

	// docker 中运行的各个 service 及从 backup.json 中读取的信息
	backupServiceMap := make(map[string]*Service)
	for _, s := range backupServices {
		backupServiceMap[s.Name] = &s
	}
	oldServices := make(map[string]Service)
	for _, s := range services {
		if backupS, ok := backupServiceMap[s.Spec.Annotations.Name]; ok {
			oldServices[s.Spec.Annotations.Name] = *backupS
		}
	}

	// 在 oldSevices 中找到可以重复使用的 service
	var reuseServices []Service
	for name, oldservice := range oldServices {
		for _, newservice := range newServices {
			if newservice.Name == oldservice.Name && isSliceEqual(newservice.CallList, oldservice.CallList) && newservice.Port == oldservice.Port {
				reuseServices = append(reuseServices, newservice)
				delete(oldServices, name)
				break
			}
		}
	}

	// 将无法重复使用的 service 删除
	for _, s := range oldServices {
		str := "docker service rm " + s.Name + "\n"
		deployFd.Write([]byte(str))
	}

	// 重新设置可重复使用 service 的信息
	for _, s := range reuseServices {
		str1 := fmt.Sprintf("docker service scale "+s.Name+"=%v\n", s.Instances)
		str2 := fmt.Sprintf("curl -d '{\"ResponseTime\":%v}' http://127.0.0.1:%v/responsetime\n", s.ResponseTime, s.Port)
		str3 := fmt.Sprintf("curl -d '{\"Concurrency\":%v}' http://127.0.0.1:%v/concurrency\n", s.Concurrency, s.Port)
		str := str1 + str2 + str3
		deployFd.Write([]byte(str))
	}

	return reuseServices
}

// isSliceEqual judges whether two slices of strings are orderly equal.
func isSliceEqual(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, sa := range a {
		if sa != b[i] {
			return false
		}
	}
	return true
}

// generateCode generates both the configuration files and deploy script for given services. 
func generateCode(deployDir string, services []Service) {

	for i, service := range services {
		logrus.Infof("Handling the %d th service: %s\n", i, service.Name)

		// 为新服务创建目录
		err := os.MkdirAll(filepath.Join(deployDir, "services/"+service.Name), 0755)
		if err != nil {
			logrus.Fatal(err)
		}

		// 为新服务创建配置文件 conf.json
		confFd, err := os.OpenFile(filepath.Join(deployDir, "services/"+service.Name+"/conf.json"), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
		if err != nil {
			logrus.Fatal(err)
		}
		defer confFd.Close()

		// 写 conf.json
		bytes, err := json.MarshalIndent(service, "", "    ")
		if err != nil {
			logrus.Fatal(err)
		}
		_, err = confFd.Write(bytes)
		if err != nil {
			logrus.Fatal(err)
		}

		// 熔断信息文件
		circuitFd, err := os.OpenFile(filepath.Join(deployDir, "services/"+service.Name+"/circuitinfo.json"), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
		if err != nil {
			logrus.Fatal(err)
		}
		defer circuitFd.Close()
		circuitinfo := CircuitInfo{
			Name:                   "AtoB",
			Timeout:                1000,
			MaxConcurrentRequests:  1000,
			SleepWindow:            5000,
			ErrorPercentThreshold:  5,
			RequestVolumeThreshold: 5,
		}
		bytes, err = json.MarshalIndent(circuitinfo, "", "    ")
		if err != nil {
			logrus.Fatal(err)
		}
		_, err = circuitFd.Write(bytes)
		if err != nil {
			logrus.Fatal(err)
		}

		// 从 source 目录中复制源文件到新的服务目录中
		// files := []string{"main.go", "router.go", "routes.go", "handler.go", "model.go", "wrapper.go", "util.go", "zipkin.go", "prometheus.go"}
		// for _, f := range files {
		// 	err = copyFile("source/"+f, "services/"+service.Name+"/"+f)

		// }

		// 写入 Dockerfile
		// dockerFd, err := os.OpenFile("services/"+service.Name+"/Dockerfile", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
		// if err != nil {
		// 	log.Fatal(err)
		// }
		// defer confFd.Close()

		// str1 := "FROM ubuntu\n"
		// str2 := "EXPOSE " + service.Port + "\n"
		// str3 := "ADD " + service.Name + " / \n"
		// str4 := "CMD [\"./" + service.Name + "\"]\n"
		// str := str1 + str2 + str3 + str4
		// _, err = dockerFd.Write([]byte(str))
		// if err != nil {
		// 	log.Fatal(err)
		// }

		// 向启动脚本写入命令
		str3 := "docker service rm " + service.Name + "\n"
		instances := fmt.Sprintf("%d", int(service.Instances))
		wd, _ := os.Getwd()
		str4 := "docker service create --name=" + service.Name + " --replicas=" + instances
		str4 = str4 + " --network=my_network"
		// Publish port only for this service.
		if service.Name == cfg.PublishedService {
			str4 = str4 + " -p=" + service.Port + ":" + service.Port
		}
		str4 = str4 + " --mount type=bind,source=" + filepath.Join(wd, deployDir, "/services/" + service.Name) + ",target=/data/" + " unusedprefix/customservice\n\n"
		// str = str1 + str2 + str3 + str4
		str := str3 + str4
		_, err = deployFd.Write([]byte(str))
		if err != nil {
			logrus.Fatal(err)
		}
	}
}

// parseServices read service configs in the specified file and returns a slice.
func parseServices(conf string) ([]Service, error) {

	// 从文件中读取 services
	jsonData, err := ioutil.ReadFile(conf)
	if err != nil {
		return nil, err
	}

	var services []Service
	err = json.Unmarshal(jsonData, &services)
	if err != nil {
		return nil, err
	}

	// 将每个 service 的名字， 调用列表中的名字改为小写
	// 处理每个 service 的 instance
	for i, _ := range services {
		services[i].Name = strings.ToLower(services[i].Name)
		for j, _ := range services[i].CallList {
			services[i].CallList[j] = strings.ToLower(services[i].CallList[j])
		}

		instance := int(services[i].Instances)
		if instance <= 0 {
			services[i].Instances = 1
		} else {
			services[i].Instances = instance
		}
	}

	// 拓扑排序， 去掉潜在的调用环路
	services = topologicalSort(services)

	if services == nil {
		return nil, errors.New("services is nil")
	}

	// 给 callList 加上端口号
	// 同时检查出是否有 service 没有给出信息
	port := make(map[string]string)
	for _, service := range services {
		port[service.Name] = service.Port
		for i, name := range service.CallList {
			if _, ok := port[name]; ok {
				service.CallList[i] = service.CallList[i] + ":" + port[name]
			} else {
				return nil, errors.New("can't find " + name + "'s information")
			}
		}
	}

	return services, nil
}

// 以 service 名为节点， callList 是节点之间的关系
// 去除掉潜在的环
func topologicalSort(services []Service) []Service {

	tmp := make(map[string]Service)
	for _, service := range services {
		tmp[service.Name] = service
	}

	graph := make(map[string][]string)
	indegree := make(map[string]int)

	for _, service := range services {
		indegree[service.Name] = 0
		for _, v := range service.CallList {
			// service 依赖 v
			graph[v] = append(graph[v], service.Name)
			indegree[service.Name]++
		}
	}

	var q []string
	for k, v := range indegree {
		if v == 0 {
			q = append(q, k)
		}
	}

	var res []Service

	for len(q) != 0 {
		x := q[0]
		res = append(res, tmp[x])
		q = q[1:]
		for _, v := range graph[x] {
			indegree[v]--
			if indegree[v] == 0 {
				q = append(q, v)
			}
		}
	}

	return res
}
