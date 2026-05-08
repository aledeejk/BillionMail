package database_initialization

import (
	//"billionmail-core/internal/consts"
	//"billionmail-core/internal/service/public"
	"context"
	"fmt"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"time"
)

var registeredHandlers = make([]func(), 0, 256)

// InitDatabase initializes the database configuration
func InitDatabase() (err error) {

	//dbPass, err := public.DockerEnv("DBPASS")
	//if err != nil {
	//	dbPass = ""
	//}

	 // Временно, для теста — прямое указание параметров
    host := "127.0.0.1"
    port := "25432"
    user := "billionmail"
    name := "billionmail"
    dbPass := "billionmail123"

    // Распечатаем для проверки
    fmt.Println("=== DB CONNECTION DEBUG ===")
    fmt.Println("Host:", host)
    fmt.Println("Port:", port)
    fmt.Println("User:", user)
    fmt.Println("Name:", name)

	// Setting database configuration
	err = gdb.SetConfig(gdb.Config{
		"default": gdb.ConfigGroup{
			gdb.ConfigNode{
				// Debug: true,
				Host:             host,
				Port:             port,
				User:             user,
				Pass:             dbPass,
				Name:             name,
				Type:             "pgsql",
				Role:             "master",
				MaxOpenConnCount: 100,
			},
		},
	})

	if err != nil {
		return fmt.Errorf("Set database configuration failed: %v", err)
	}

	// Testing database connection until successful
	connectionOK := false
	for i := 0; i < 10; i++ {
		_, err = g.DB().Exec(context.Background(), "SELECT 1")

		if err != nil {
			// Wait for 5 seconds before retrying
			g.Log().Debug(context.Background(), "Database connection failed, retrying in 5 seconds...")
			time.Sleep(time.Second * 5)
			continue
		}

		connectionOK = true
		g.Log().Debug(context.Background(), "Database connection successful")
		break
	}

	if !connectionOK {
		return fmt.Errorf("database connection test failed after 10 attempts")
	}

	// Execute registered handlers
	for _, handler := range registeredHandlers {
		if handler != nil {
			handler()
		}
	}

	// Empty the registered handlers
	registeredHandlers = registeredHandlers[:0]

	return nil
}

// registerHandler registers a handler for the database initialization
func registerHandler(handler func()) {
	if handler == nil {
		return
	}

	registeredHandlers = append(registeredHandlers, handler)
}
