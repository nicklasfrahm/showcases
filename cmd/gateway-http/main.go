package main

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/nicklasfrahm/showcases/pkg/broker"
	"github.com/nicklasfrahm/showcases/pkg/gateway"
	"github.com/nicklasfrahm/showcases/pkg/service"
)

var (
	name      = "unknown"
	version   = "dev"
	mapMethod = map[string]string{
		http.MethodGet:    "read",
		http.MethodPut:    "update",
		http.MethodDelete: "delete",
	}
	mapListMethod = map[string]string{
		http.MethodGet:  "find",
		http.MethodPost: "create",
	}
)

func main() {
	// Load authorized users.
	users := make(map[string]string)
	usersCreds := strings.Split(os.Getenv("AUTHORIZED_CREDENTIALS"), ",")
	for _, userCred := range usersCreds {
		userPass := strings.Split(userCred, ":")
		if len(userPass) == 2 {
			users[userPass[0]] = userPass[1]
		}
	}

	// Create new service instance.
	svc := service.New(service.Config{
		Name:    name,
		Version: version,
	})

	// Configure broker connection.
	svc.UseBroker(broker.NewNATS(&broker.NATSOptions{
		URI:            os.Getenv("BROKER_URI"),
		RequestTimeout: 1 * time.Second,
	}))

	// Configure gateway.
	svc.UseGateway(gateway.NewHTTP(gateway.Port(os.Getenv("PORT"))))

	svc.GatewayMiddleware(WebsocketStreams())
	svc.GatewayMiddleware(NormalizeProtoToChannel())
	svc.GatewayMiddleware(AuthN(users))
	svc.GatewayMiddleware(DispatchToChannel())

	// Wait until error occurs or signal is received.
	svc.Start()
}
