package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
)

// ============================================================
// rpa_prompts.go — RPA Agent 工具定义、Prompt 模板、层1调度
// ============================================================

// ---------- 工具定义 ----------

// rpaToolDefs 所有工具的完整定义（内部使用）
// 精简为 10 个核心工具：层2 RPA 界面操作 + exec_cmd + ask_user + task_done
// description 尽量精简，每个工具控制在 30 tokens 以内
var rpaToolDefs = []map[string]interface{}{
	// ---- 层2：RPA 界面操作工具 ----
	{
		"type": "function",
		"function": map[string]interface{}{
			"name":        "get_screen_context",
			"description": "获取屏幕可点击节点列表（含坐标），操作前后必须调用确认界面状态",
			"parameters":  map[string]interface{}{"type": "object", "properties": map[string]interface{}{}, "required": []string{}},
		},
	},
	{
		"type": "function",
		"function": map[string]interface{}{
			"name":        "open_app",
			"description": "启动App。抖音=com.ss.android.ugc.aweme 微信=com.tencent.mm 快手=com.smile.gifmaker",
			"parameters": map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{"pkg": map[string]interface{}{"type": "string", "description": "App包名"}},
				"required":   []string{"pkg"},
			},
		},
	},
	{
		"type": "function",
		"function": map[string]interface{}{
			"name":        "click",
			"description": "点击屏幕坐标",
			"parameters": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"x": map[string]interface{}{"type": "integer", "description": "X像素"},
					"y": map[string]interface{}{"type": "integer", "description": "Y像素"},
				},
				"required": []string{"x", "y"},
			},
		},
	},
	{
		"type": "function",
		"function": map[string]interface{}{
			"name":        "swipe",
			"description": "滑动屏幕。所有坐标均为屏幕绝对像素坐标。上滑刷内容示例：x0=540,y0=1400,x1=540,y1=400（y1必须是绝对坐标，不是偏移量）",
			"parameters": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"x0":          map[string]interface{}{"type": "integer", "description": "起始X（绝对像素）"},
					"y0":          map[string]interface{}{"type": "integer", "description": "起始Y（绝对像素），上滑时填1400"},
					"x1":          map[string]interface{}{"type": "integer", "description": "结束X（绝对像素）"},
					"y1":          map[string]interface{}{"type": "integer", "description": "结束Y（绝对像素），上滑刷内容时填400"},
					"duration_ms": map[string]interface{}{"type": "integer", "description": "滑动时长ms，默认1000"},
				},
				"required": []string{"x0", "y0", "x1", "y1"},
			},
		},
	},
	{
		"type": "function",
		"function": map[string]interface{}{
			"name":        "send_text",
			"description": "向输入框输入文字",
			"parameters": map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{"text": map[string]interface{}{"type": "string", "description": "文字内容"}},
				"required":   []string{"text"},
			},
		},
	},
	{
		"type": "function",
		"function": map[string]interface{}{
			"name":        "key_press",
			"description": "按键。3=HOME 4=BACK 66=ENTER 187=多任务",
			"parameters": map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{"key_code": map[string]interface{}{"type": "integer", "description": "键码"}},
				"required":   []string{"key_code"},
			},
		},
	},
	{
		"type": "function",
		"function": map[string]interface{}{
			"name":        "stop_app",
			"description": "强制关闭App",
			"parameters": map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{"pkg": map[string]interface{}{"type": "string", "description": "App包名"}},
				"required":   []string{"pkg"},
			},
		},
	},
	// ---- exec_cmd：通过 HTTP 执行 adb shell 命令（不依赖 RPA SDK 连接）----
	{
		"type": "function",
		"function": map[string]interface{}{
			"name":        "exec_cmd",
			"description": "通过HTTP执行adb shell命令，不依赖RPA连接",
			"parameters": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"cmd":           map[string]interface{}{"type": "string", "description": "shell命令"},
					"wait_for_exit": map[string]interface{}{"type": "boolean", "description": "等待完成"},
				},
				"required": []string{"cmd"},
			},
		},
	},
	// ---- 控制工具 ----
	{
		"type": "function",
		"function": map[string]interface{}{
			"name":        "ask_user",
			"description": "向用户提问并等待回复",
			"parameters": map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{"question": map[string]interface{}{"type": "string", "description": "问题"}},
				"required":   []string{"question"},
			},
		},
	},
	{
		"type": "function",
		"function": map[string]interface{}{
			"name":        "task_done",
			"description": "任务完成时必须调用",
			"parameters": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"success": map[string]interface{}{"type": "boolean", "description": "是否成功"},
					"summary": map[string]interface{}{"type": "string", "description": "结果摘要"},
				},
				"required": []string{"success", "summary"},
			},
		},
	},
}

// rpaToolNamesByTask 各任务类型工具白名单
// 精简后共 10 个工具：get_screen_context / click / swipe / send_text / key_press /
//   open_app / stop_app / exec_cmd / ask_user / task_done
var rpaToolNamesByTask = map[string][]string{
	"browse_video": {"get_screen_context", "open_app", "swipe", "click", "key_press", "exec_cmd", "task_done"},
	"send_message": {"get_screen_context", "open_app", "click", "swipe", "send_text", "key_press", "task_done"},
	"like_content": {"get_screen_context", "open_app", "click", "swipe", "key_press", "exec_cmd", "task_done"},
	"search":       {"get_screen_context", "open_app", "click", "swipe", "send_text", "key_press", "exec_cmd", "task_done"},
	"install_app":  {"get_screen_context", "open_app", "click", "send_text", "exec_cmd", "task_done"},
	// custom: nil → 返回全部（调用方自行承担 token 超限风险）
	"custom": nil,
}

// RpaGetToolsJSON 返回指定任务类型所需工具的 OpenAI tools 格式定义
// taskType 为空或 "custom" 时返回全部工具
func (a *App) RpaGetToolsJSON(taskType string) interface{} {
	allowedNames, ok := rpaToolNamesByTask[taskType]
	// nil 或未知 taskType → 返回全部
	if !ok || allowedNames == nil {
		return rpaToolDefs
	}

	allowed := make(map[string]bool, len(allowedNames))
	for _, n := range allowedNames {
		allowed[n] = true
	}

	result := make([]map[string]interface{}, 0, len(allowedNames))
	for _, tool := range rpaToolDefs {
		fn, _ := tool["function"].(map[string]interface{})
		name, _ := fn["name"].(string)
		if allowed[name] {
			result = append(result, tool)
		}
	}
	return result
}

// ---------- Prompt 模板 ----------

// RpaBuildSystemPrompt 根据任务类型构建 system prompt
// 1024 token 窗口下必须极简，核心规则内嵌到工具 description 中
func (a *App) RpaBuildSystemPrompt(taskType string, deviceResolution string, experience string) string {
	// 极简 base：只保留最关键的执行约束
	base := "你是Android自动化助手。规则：每次只调一个工具；必须用tool_call格式调用工具，禁止用文字描述代替工具调用；完成或失败必须调task_done。open_app返回值的nodes字段就是当前界面节点，直接从中读取坐标操作，无需再调get_screen_context；nodes显示'界面节点暂未就绪'时才调get_screen_context重试。禁止输出None或空内容。所有坐标均为屏幕绝对像素值。"

	if deviceResolution != "" {
		base += "屏幕分辨率：" + deviceResolution + "。"
	}

	// 任务专属提示（极简版）
	taskHints := map[string]string{
		"browse_video": "步骤：open_app(返回nodes即当前界面)→[循环: 调用swipe工具(x0=540,y0=1400,x1=540,y1=400,duration_ms=400，注意y1=400是绝对坐标)→看swipe结果里的计数提示，达到目标条数后调task_done]。严格按计数提示决定是否继续。",
		"send_message": "步骤：open_app(返回nodes即当前界面)→从nodes找输入框坐标→click输入框→get_screen_context(确认聚焦)→send_text→key_press(66)→get_screen_context(确认已发送)→task_done。",
		"like_content": "步骤：open_app(返回nodes即当前界面)→从nodes找点赞按钮坐标→click→get_screen_context(确认已点赞)→task_done。",
		"search":       "步骤：open_app(返回nodes即当前界面)→从nodes找搜索框坐标→click搜索框→send_text→key_press(66)→get_screen_context(确认搜索结果)→task_done。",
		"install_app":  "优先用install_apk直接安装。",
		"custom": "当前界面节点已在对话历史中提供。规则：1.直接从历史中的节点列表找目标元素的center坐标；2.若目标App未打开则先调open_app；3.找到目标坐标后调click；4.操作后调get_screen_context确认结果；5.完成调task_done。",
	}

	if hint, ok := taskHints[taskType]; ok && hint != "" {
		base += hint
	}

	// 注入向量库经验（已经很精简了，保留）
	if experience != "" {
		base += "参考经验：" + experience
	}

	return base
}

// ---------- 层1工具调度 ----------

// RpaCallFixedTool 层1固定规则工具的统一调度入口
// 精简后只保留 exec_cmd（通过 HTTP exec 通道）
// toolName 对应工具名，args 为 JSON 解析后的参数 map
// 额外参数通过 rpaContext 传入：deviceIP、containerName
func (a *App) RpaCallFixedTool(toolName string, args map[string]interface{}) map[string]interface{} {
	log.Printf("[RpaCallFixedTool] tool=%s args=%v", toolName, args)

	getString := func(key string) string {
		v, _ := args[key].(string)
		return v
	}

	switch toolName {

	// ---- exec_cmd：通过 HTTP exec 通道执行 adb shell 命令 ----
	// 前端 agentLoop 应将 device_ip 和 container_name 注入 args
	case "exec_cmd":
		deviceIP := getString("device_ip")
		containerName := getString("container_name")
		cmd := getString("cmd")
		if deviceIP == "" || cmd == "" {
			return map[string]interface{}{"success": false, "message": "缺少 device_ip 或 cmd 参数"}
		}
		result := a.execAndroidCmdForRpa(deviceIP, containerName, cmd)
		log.Printf("[RpaCallFixedTool] exec_cmd device=%s container=%s cmd=%q result=%v", deviceIP, containerName, cmd, result)
		return result

	default:
		return map[string]interface{}{"success": false, "message": fmt.Sprintf("未知工具: %s", toolName)}
	}
}

// execAndroidCmdForRpa 通过设备 HTTP API 在指定容器中执行 adb shell 命令
// 接口：POST http://{deviceIP}:8000/android/exec
// Body: {"name": "容器名", "command": ["sh", "-c", "命令"]}
func (a *App) execAndroidCmdForRpa(deviceIP, containerName, cmd string) map[string]interface{} {
	reqBody := map[string]interface{}{
		"name":    containerName,
		"command": []string{"sh", "-c", cmd},
	}

	url := fmt.Sprintf("http://%s/android/exec", deviceAddr(deviceIP))
	log.Printf("[execAndroidCmdForRpa] POST %s name=%s cmd=%q", url, containerName, cmd)

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return map[string]interface{}{"success": false, "message": "序列化失败: " + err.Error()}
	}

	httpReq, err := a.buildDeviceRequest("POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return map[string]interface{}{"success": false, "message": "构造请求失败: " + err.Error()}
	}

	resp, err := a.httpClient.Do(httpReq)
	if err != nil {
		log.Printf("[execAndroidCmdForRpa] HTTP 失败: %v", err)
		return map[string]interface{}{"success": false, "message": "HTTP 请求失败: " + err.Error()}
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return map[string]interface{}{"success": false, "message": "读取响应失败"}
	}
	log.Printf("[execAndroidCmdForRpa] status=%d body=%s", resp.StatusCode, string(body))

	if decodeErr := json.Unmarshal(body, &result); decodeErr != nil {
		// 响应不是 JSON，直接返回原始文本
		return map[string]interface{}{"success": true, "output": string(body)}
	}

	// V3 API 响应格式: {code: 0, data: "输出"} 或 {code: 0, message: "ok"}
	if code, ok := result["code"]; ok {
		codeStr := fmt.Sprintf("%v", code)
		if codeStr == "0" || codeStr == "200" {
			output := ""
			if d, ok := result["data"].(string); ok {
				output = d
			} else if d, ok := result["data"]; ok && d != nil {
				output = fmt.Sprintf("%v", d)
			}
			return map[string]interface{}{"success": true, "output": output}
		}
		msg, _ := result["message"].(string)
		return map[string]interface{}{"success": false, "message": msg}
	}

	// 无 code 字段：HTTP 200 视为成功
	if resp.StatusCode == 200 {
		output, _ := result["output"].(string)
		if output == "" {
			output, _ = result["data"].(string)
		}
		return map[string]interface{}{"success": true, "output": output}
	}
	msg, _ := result["message"].(string)
	return map[string]interface{}{"success": false, "message": msg}
}
