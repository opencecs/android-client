package main

import (
	"fmt"
)

func (a *App) StartP2P(deviceIP string, port int, streamName string, listenPort int) map[string]interface{} {
	if a.P2PManager == nil {
		return map[string]interface{}{"success": false, "message": "p2p manager not initialized"}
	}
	if listenPort <= 0 {
		return map[string]interface{}{"success": false, "message": "listen port required"}
	}
	if streamName == "" {
		streamName = fmt.Sprintf("p2p-%d", listenPort)
	}
	streamURL := fmt.Sprintf("srt://%s:%d", GetLocalIP(), listenPort)
	result := a.P2PManager.Start(deviceIP, port, streamName, streamURL, listenPort)
	if result["success"] == true {
		fmt.Printf("[SRT] Listening for SRT connection on %s\n", streamURL)
		result["listenUrl"] = streamURL
	}
	return result
}

func (a *App) StopP2P(deviceIP string, port int, streamName string) map[string]interface{} {
	if a.P2PManager == nil {
		return map[string]interface{}{"success": false, "message": "p2p manager not initialized"}
	}
	return a.P2PManager.Stop(deviceIP, port, streamName)
}

func (a *App) GetP2PStatus(deviceIP string, port int, streamName string) map[string]interface{} {
	if a.P2PManager == nil {
		return map[string]interface{}{"running": false, "message": "p2p manager not initialized"}
	}
	return a.P2PManager.Status(deviceIP, port, streamName)
}

func (a *App) StopAllP2P() {
	if a.P2PManager == nil {
		return
	}
	a.P2PManager.StopAll()
}
