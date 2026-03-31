package main

import (
	"fmt"
	"testing"
)

// TestParseWebRTCURL 测试 WebRTC URL 解析
func TestParseWebRTCURL(t *testing.T) {
	service := NewRtmpService()
	
	testCases := []struct {
		input    string
		expected string
		hasError bool
	}{
		{
			input:    "webrtc://192.168.1.100/live/test",
			expected: "http://192.168.1.100:1985/rtc/v1/whep/?app=live&stream=test",
			hasError: false,
		},
		{
			input:    "webrtc://192.168.1.100:8080/live/demo",
			expected: "http://192.168.1.100:8080/rtc/v1/whep/?app=live&stream=demo",
			hasError: false,
		},
		{
			input:    "webrtc://localhost/myapp/stream123",
			expected: "http://localhost:1985/rtc/v1/whep/?app=myapp&stream=stream123",
			hasError: false,
		},
		{
			input:    "rtmp://192.168.1.100/live/test",
			expected: "",
			hasError: true,
		},
		{
			input:    "webrtc://192.168.1.100",
			expected: "",
			hasError: true,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result, err := service.ParseWebRTCURL(tc.input)
			
			if tc.hasError {
				if err == nil {
					t.Errorf("Expected error for input %s, but got none", tc.input)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for input %s: %v", tc.input, err)
				}
				if result != tc.expected {
					t.Errorf("For input %s:\nExpected: %s\nGot:      %s", tc.input, tc.expected, result)
				}
			}
		})
	}
}

// TestGetWebRTCPlayURL 测试生成 WebRTC 播放地址
func TestGetWebRTCPlayURL(t *testing.T) {
	service := NewRtmpService()
	
	// 测试 webrtc:// URL 格式
	url := service.GetWebRTCPlayURL("test", false)
	fmt.Printf("WebRTC Play URL: %s\n", url)
	
	// 测试 HTTP 信令 URL 格式
	sdpURL := service.GetWebRTCPlayURL("test", true)
	fmt.Printf("WebRTC SDP URL: %s\n", sdpURL)
}

// TestGetWebRTCPushURL 测试生成 WebRTC 推流地址
func TestGetWebRTCPushURL(t *testing.T) {
	service := NewRtmpService()
	
	url := service.GetWebRTCPushURL("test")
	fmt.Printf("WebRTC Push URL (WHIP): %s\n", url)
}

// ExampleUsage 演示完整的使用流程
func ExampleUsage() {
	service := NewRtmpService()
	
	// 1. 创建流
	streamName := "demo"
	service.AddStreamName(streamName)
	
	// 2. 获取 WebRTC 播放地址（给客户端使用）
	playURL := service.GetWebRTCPlayURL(streamName, false)
	fmt.Printf("播放地址: %s\n", playURL)
	
	// 3. 获取 WebRTC 推流地址（给推流端使用）
	pushURL := service.GetWebRTCPushURL(streamName)
	fmt.Printf("推流地址: %s\n", pushURL)
	
	// 4. 解析 WebRTC URL（从客户端接收）
	clientURL := "webrtc://192.168.1.100/live/demo"
	httpURL, err := service.ParseWebRTCURL(clientURL)
	if err != nil {
		fmt.Printf("解析失败: %v\n", err)
	} else {
		fmt.Printf("HTTP 信令地址: %s\n", httpURL)
	}
	
	// 5. 推送到设备
	deviceIPs := []string{"192.168.1.101", "192.168.1.102"}
	result := service.PushStreamToDevices(streamName, deviceIPs)
	fmt.Printf("推送结果: %v\n", result)
}
