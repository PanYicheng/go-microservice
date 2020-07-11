package main
import (
        "fmt"
        "github.com/PanYicheng/go-microservice/service-demo/service"  // NEW
)
var appName = "service-demo"
func main() {
        fmt.Printf("Starting %v\n", appName)
        service.StartWebServer("6767")           // NEW
}

