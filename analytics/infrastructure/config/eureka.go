package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"
)

// EurekaClient representa un cliente de Eureka
type EurekaClient struct {
	serverURL   string
	serviceName string
	instanceID  string
	ipAddr      string
	port        string
}

// getOutboundIP obtiene la IP real de la máquina (no 0.0.0.0 o 127.0.0.1)
func getOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Printf("Warning: could not detect IP address: %v", err)
		return "127.0.0.1"
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

// NewEurekaClient crea un nuevo cliente de Eureka
// Si instanceIP está vacío, detecta automáticamente la IP
func NewEurekaClient(serverURL, serviceName, instanceIP, port string) *EurekaClient {
	// Si la IP no está configurada o es 0.0.0.0, detectar la IP real
	if instanceIP == "" || instanceIP == "0.0.0.0" {
		instanceIP = getOutboundIP()
		log.Printf("Auto-detected IP address for Eureka registration: %s", instanceIP)
	} else {
		log.Printf("Using configured IP address for Eureka registration: %s", instanceIP)
	}

	instanceID := fmt.Sprintf("%s:%s:%s", serviceName, instanceIP, port)
	return &EurekaClient{
		serverURL:   serverURL,
		serviceName: serviceName,
		instanceID:  instanceID,
		ipAddr:      instanceIP,
		port:        port,
	}
}

// Register registra el servicio en Eureka
func (c *EurekaClient) Register() error {
	registerURL := fmt.Sprintf("%sapps/%s", c.serverURL, c.serviceName)

	payload := map[string]interface{}{
		"instance": map[string]interface{}{
			"instanceId": c.instanceID,
			"hostName":   c.ipAddr,
			"app":        c.serviceName,
			"ipAddr":     c.ipAddr,
			"status":     "UP",
			"port": map[string]interface{}{
				"$":        c.port,
				"@enabled": "true",
			},
			"healthCheckUrl":                fmt.Sprintf("http://%s:%s/health", c.ipAddr, c.port),
			"statusPageUrl":                 fmt.Sprintf("http://%s:%s/info", c.ipAddr, c.port),
			"homePageUrl":                   fmt.Sprintf("http://%s:%s/", c.ipAddr, c.port),
			"vipAddress":                    c.serviceName,
			"secureVipAddress":              c.serviceName,
			"isCoordinatingDiscoveryServer": "false",
			"leaseInfo": map[string]interface{}{
				"renewalIntervalInSecs": 10,
				"durationInSecs":        30,
			},
			"metadata": map[string]interface{}{
				"management.port": c.port,
			},
			"dataCenterInfo": map[string]interface{}{
				"@class": "com.netflix.appinfo.InstanceInfo$DefaultDataCenterInfo",
				"name":   "MyOwn",
			},
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal registration payload: %w", err)
	}

	req, err := http.NewRequest("POST", registerURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create registration request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to register with eureka: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("eureka registration failed with status: %d", resp.StatusCode)
	}

	log.Printf("Successfully registered with Eureka: %s", c.instanceID)
	return nil
}

// SendHeartbeat envía heartbeat a Eureka
func (c *EurekaClient) SendHeartbeat() error {
	heartbeatURL := fmt.Sprintf("%sapps/%s/%s", c.serverURL, c.serviceName, c.instanceID)

	req, err := http.NewRequest("PUT", heartbeatURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create heartbeat request: %w", err)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send heartbeat: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("heartbeat failed with status: %d", resp.StatusCode)
	}

	return nil
}

// StartHeartbeat inicia el envío periódico de heartbeats
func (c *EurekaClient) StartHeartbeat() {
	// Heartbeat cada 10 segundos (matching leaseInfo.renewalIntervalInSecs)
	ticker := time.NewTicker(10 * time.Second)
	go func() {
		for range ticker.C {
			if err := c.SendHeartbeat(); err != nil {
				log.Printf("Heartbeat error: %v", err)
			} else {
				log.Printf("Heartbeat sent successfully for %s", c.instanceID)
			}
		}
	}()
	log.Printf("Started heartbeat for %s (every 10 seconds)", c.instanceID)
}

// Deregister elimina el registro del servicio de Eureka
func (c *EurekaClient) Deregister() error {
	deregisterURL := fmt.Sprintf("%sapps/%s/%s", c.serverURL, c.serviceName, c.instanceID)

	req, err := http.NewRequest("DELETE", deregisterURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create deregister request: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to deregister: %w", err)
	}
	defer resp.Body.Close()

	log.Printf("Deregistered from Eureka: %s", c.instanceID)
	return nil
}
