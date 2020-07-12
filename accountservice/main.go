package main
import (
        "fmt"
        "github.com/PanYicheng/go-microservice/accountservice/service"  // NEW
        "github.com/PanYicheng/go-microservice/accountservice/dbclient"  // NEW
)
var appName = "accountservice"
func main() {
        fmt.Printf("Starting %v\n", appName)
        initializeBoltClient()
		service.StartWebServer("6767")           // NEW
}
// Creates instance and calls the OpenBoltDb and Seed funcs
func initializeBoltClient() {
        service.DBClient = &dbclient.BoltClient{}
        service.DBClient.OpenBoltDb()
        service.DBClient.Seed()
}
