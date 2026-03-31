package main

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

// ============================================================
// AI 服务 — 所有与设备 AI 相关的 HTTP 通信全部在 Go 后端完成
// 前端只负责展示，通过 Wails 事件接收流式 chunk
// ============================================================

// ---------- 数据结构 ----------

// AIChatMessage 单条对话消息（与 OpenAI 格式兼容）
// ToolCalls/ToolCallID 用于 RPA Agent function calling
// Content 使用 *string 支持 null（assistant + tool_calls 时 content 为空字符串""）
type AIChatMessage struct {
	Role       string      `json:"role"`
	Content    *string     `json:"content"`                // 指针，允许为 null（omitempty 对指针 nil 生效）
	ToolCalls  interface{} `json:"tool_calls,omitempty"`   // assistant 消息的工具调用列表
	ToolCallID string      `json:"tool_call_id,omitempty"` // tool 消息的调用 ID
	Name       string      `json:"name,omitempty"`         // tool 消息的工具名（可选）
}

// AIChatRequest 前端发起对话请求
type AIChatRequest struct {
	DeviceIP    string                   `json:"deviceIp"`
	Model       string                   `json:"model"`
	Messages    []AIChatMessage          `json:"messages"`
	Temperature float64                  `json:"temperature"`
	MaxTokens   int                      `json:"maxTokens"`
	SessionID   string                   `json:"sessionId"`   // 用于多会话并发，前端生成唯一 ID
	Tools       []map[string]interface{} `json:"tools"`       // 工具定义（RPA Agent 模式）
	ToolChoice  string                   `json:"toolChoice"`  // "auto" | "none" | "required"
}

// AIModelStartRequest 启动模型请求（前端传来整个 modelConfig）
type AIModelStartRequest struct {
	DeviceIP    string                 `json:"deviceIp"`
	ModelConfig map[string]interface{} `json:"modelConfig"`
}

// ---------- 会话管理（支持取消） ----------

type aiSession struct {
	cancel chan struct{}
}

var (
	aiSessionsMu sync.Mutex
	aiSessions   = make(map[string]*aiSession)
)

func (a *App) registerAISession(sessionID string) *aiSession {
	aiSessionsMu.Lock()
	defer aiSessionsMu.Unlock()
	// 若旧会话存在，先取消
	if old, ok := aiSessions[sessionID]; ok {
		close(old.cancel)
	}
	sess := &aiSession{cancel: make(chan struct{})}
	aiSessions[sessionID] = sess
	return sess
}

func (a *App) cancelAISession(sessionID string) {
	aiSessionsMu.Lock()
	defer aiSessionsMu.Unlock()
	if sess, ok := aiSessions[sessionID]; ok {
		close(sess.cancel)
		delete(aiSessions, sessionID)
	}
}

func (a *App) removeAISession(sessionID string) {
	aiSessionsMu.Lock()
	defer aiSessionsMu.Unlock()
	delete(aiSessions, sessionID)
}

// ---------- 认证头 ----------

func (a *App) getDeviceAuthHeader(deviceIP string) string {
	a.devicePasswordsMutex.RLock()
	defer a.devicePasswordsMutex.RUnlock()
	if a.devicePasswords == nil {
		return ""
	}
	pw, ok := a.devicePasswords[deviceIP]
	if !ok || pw == "" {
		return ""
	}
	encoded := base64.StdEncoding.EncodeToString([]byte("admin:" + pw))
	return "Basic " + encoded
}

// buildDeviceRequest 构造带认证头的 HTTP 请求
func (a *App) buildDeviceRequest(method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	// 从 URL 中提取 IP
	ip := ""
	s := strings.TrimPrefix(url, "http://")
	if idx := strings.Index(s, ":"); idx > 0 {
		ip = s[:idx]
	}
	if auth := a.getDeviceAuthHeader(ip); auth != "" {
		req.Header.Set("Authorization", auth)
	}
	return req, nil
}

// ---------- 核心方法 ----------

// AISendMessage 前端调用：发送对话消息，流式 chunk 通过 Wails 事件推送
// 事件名：
//   ai:chunk:{sessionId}   → { content, reasoning }
//   ai:done:{sessionId}    → { content, reasoning, usage }
//   ai:error:{sessionId}   → { message }
func (a *App) AISendMessage(req AIChatRequest) map[string]interface{} {
	if req.SessionID == "" {
		req.SessionID = fmt.Sprintf("%d", time.Now().UnixMilli())
	}

	sess := a.registerAISession(req.SessionID)
	go func() {
		defer a.removeAISession(req.SessionID)
		log.Printf("[AIService] 开始流式请求 session=%s model=%s device=%s msgs=%d tools=%d",
			req.SessionID, req.Model, req.DeviceIP, len(req.Messages), len(req.Tools))
		// 打印每条消息的 role 和内容摘要，便于排查注入是否生效
		for i, m := range req.Messages {
			content := ""
			if m.Content != nil {
				content = *m.Content
			}
			log.Printf("[AIService]   msg[%d] role=%s content=%q", i, m.Role, truncate(content, 100))
		}
		if len(req.Tools) > 0 {
			log.Printf("[AIService]   tools[0] name=%v", req.Tools[0]["function"])
		}
		if err := a.doAIStream(req, sess); err != nil {
			log.Printf("[AIService] 流式对话失败 session=%s: %v", req.SessionID, err)
			a.emitEvent("ai:error:"+req.SessionID, map[string]interface{}{
				"message": err.Error(),
			})
		} else {
			log.Printf("[AIService] 流式对话完成 session=%s", req.SessionID)
		}
	}()

	return map[string]interface{}{"success": true, "sessionId": req.SessionID}
}

// truncate 截断字符串，用于日志输出
func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "..."
}

// AICancelMessage 前端调用：取消正在进行的对话
func (a *App) AICancelMessage(sessionID string) map[string]interface{} {
	a.cancelAISession(sessionID)
	return map[string]interface{}{"success": true}
}

// doAIStream 实际执行流式请求（在 goroutine 中运行）
func (a *App) doAIStream(req AIChatRequest, sess *aiSession) error {
	url := fmt.Sprintf("http://%s/v1/chat/completions", deviceAddr(req.DeviceIP))

	temperature := req.Temperature
	if temperature == 0 {
		temperature = 0.7
	}
	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = 4096
	}

	payload := map[string]interface{}{
		"model":       req.Model,
		"messages":    req.Messages,
		"temperature": temperature,
		"max_tokens":  maxTokens,
		"stream":      true,
	}
	if len(req.Tools) > 0 {
		payload["tools"] = req.Tools
		if req.ToolChoice != "" {
			payload["tool_choice"] = req.ToolChoice
		} else {
			payload["tool_choice"] = "auto"
		}
		log.Printf("[AIService] doAIStream tools=%d toolChoice=%s session=%s", len(req.Tools), payload["tool_choice"], req.SessionID)
	} else {
		log.Printf("[AIService] doAIStream tools=0 (no function calling) session=%s", req.SessionID)
	}
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("序列化请求失败: %w", err)
	}

	httpReq, err := a.buildDeviceRequest("POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("构造请求失败: %w", err)
	}

	resp, err := a.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("HTTP 请求失败: %w", err)
	}
	defer resp.Body.Close()

	log.Printf("[AIService] HTTP响应 status=%d session=%s", resp.StatusCode, req.SessionID)
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("设备返回 %d: %s", resp.StatusCode, string(body))
	}

	// 流式解析 SSE
	var (
		fullContent   strings.Builder
		fullReasoning strings.Builder
		usage         map[string]interface{}
		isInThink     bool
		thinkBuf      strings.Builder
		chunkBuf      strings.Builder
	)
	// tool_calls 增量累积：index → { id, name, argsStr }
	toolCallMap := make(map[int]map[string]interface{})

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 64*1024), 64*1024)

	// 批量 flush：每 40ms 推送一次 chunk，减少事件频率
	flushTicker := time.NewTicker(40 * time.Millisecond)
	defer flushTicker.Stop()

	flushPending := func() {
		content := chunkBuf.String()
		reasoning := ""
		// 把 thinkBuf 里已确定在 think 标签内的内容也一并推送
		if isInThink {
			reasoning = thinkBuf.String()
			fullReasoning.WriteString(reasoning)
			thinkBuf.Reset()
		}
		if content != "" || reasoning != "" {
			fullContent.WriteString(content)
			chunkBuf.Reset()
			a.emitEvent("ai:chunk:"+req.SessionID, map[string]interface{}{
				"content":   content,
				"reasoning": reasoning,
			})
		}
	}

	doneCh := make(chan struct{})
	go func() {
		defer close(doneCh)
		lineCount := 0
		for scanner.Scan() {
			// 检查取消信号
			select {
			case <-sess.cancel:
				return
			default:
			}

			line := scanner.Text()
			lineCount++
			if lineCount <= 5 {
				log.Printf("[AIService] SSE line[%d]: %s", lineCount, truncate(line, 120))
			}
			if !strings.HasPrefix(line, "data: ") {
				continue
			}
			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				continue
			}


	var chunk struct {
		Choices []struct {
			Delta struct {
				Content   string `json:"content"`
				Reasoning string `json:"reasoning"`
				ToolCalls []struct {
					Index    int    `json:"index"`
					ID       string `json:"id"`
					Function struct {
						Name      string `json:"name"`
						Arguments string `json:"arguments"`
					} `json:"function"`
				} `json:"tool_calls"`
			} `json:"delta"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
		Usage map[string]interface{} `json:"usage"`
	}
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}
		if chunk.Usage != nil {
			usage = chunk.Usage
		}
		if len(chunk.Choices) == 0 {
			continue
		}

		// 检测 token 溢出或模型异常停止：finish_reason=stop 但没有任何 tool_calls 且 content 为空或 None
		if chunk.Choices[0].FinishReason == "stop" {
			promptTokens := 0
			completionTokens := 0
			if usage != nil {
				if v, ok := usage["prompt_tokens"].(float64); ok {
					promptTokens = int(v)
				}
				if v, ok := usage["completion_tokens"].(float64); ok {
					completionTokens = int(v)
				}
			}
			if completionTokens <= 2 && len(toolCallMap) == 0 {
				if promptTokens > 1800 {
					log.Printf("[AIService] ⚠️  token溢出：prompt_tokens=%d 超出窗口，finish_reason=stop completion_tokens=%d session=%s",
						promptTokens, completionTokens, req.SessionID)
				} else {
					log.Printf("[AIService] ⚠️  模型异常停止(非溢出)：prompt_tokens=%d completion_tokens=%d content=%q session=%s",
						promptTokens, completionTokens, fullContent.String(), req.SessionID)
				}
			}
		}

		delta := chunk.Choices[0].Delta

		// 处理 tool_calls delta（流式累积）
		if len(delta.ToolCalls) > 0 {
			for _, tc := range delta.ToolCalls {
				idx := tc.Index
				if toolCallMap[idx] == nil {
					toolCallMap[idx] = map[string]interface{}{
						"id": tc.ID, "name": "", "argsStr": "",
					}
				}
				if tc.ID != "" {
					toolCallMap[idx]["id"] = tc.ID
				}
				if tc.Function.Name != "" {
					toolCallMap[idx]["name"] = toolCallMap[idx]["name"].(string) + tc.Function.Name
				}
				if tc.Function.Arguments != "" {
					toolCallMap[idx]["argsStr"] = toolCallMap[idx]["argsStr"].(string) + tc.Function.Arguments
				}
			}
		}

		// 处理 reasoning 字段（兼容旧版）
		if delta.Reasoning != "" {
			fullReasoning.WriteString(delta.Reasoning)
			a.emitEvent("ai:chunk:"+req.SessionID, map[string]interface{}{
				"content":   "",
				"reasoning": delta.Reasoning,
			})
		}

		// 处理 content，同时解析 <think> 标签
		if delta.Content != "" {
			thinkBuf.WriteString(delta.Content)
			for {
				cur := thinkBuf.String()
				if !isInThink {
					idx := strings.Index(cur, "<think>")
					if idx < 0 {
						chunkBuf.WriteString(cur)
						thinkBuf.Reset()
						break
					}
					chunkBuf.WriteString(cur[:idx])
					thinkBuf.Reset()
					thinkBuf.WriteString(cur[idx+7:])
					isInThink = true
				} else {
					idx := strings.Index(cur, "</think>")
					if idx < 0 {
						break
					}
					reasoning := cur[:idx]
					fullReasoning.WriteString(reasoning)
					a.emitEvent("ai:chunk:"+req.SessionID, map[string]interface{}{
						"content":   "",
						"reasoning": reasoning,
					})
					thinkBuf.Reset()
					thinkBuf.WriteString(cur[idx+8:])
					isInThink = false
				}
			}
		}

		// 检查是否到了 flush 时机
		select {
		case <-flushTicker.C:
			flushPending()
		default:
		}
	}
		// 扫描结束，把剩余内容全部 flush
		flushPending()
		log.Printf("[AIService] SSE扫描结束 totalLines=%d session=%s scanErr=%v", lineCount, req.SessionID, scanner.Err())
	}()

// 等待流读完或被取消
select {
case <-doneCh:
case <-sess.cancel:
	log.Printf("[AIService] session=%s 被取消", req.SessionID)
	return nil
}

// 整理 tool_calls 放入 done 事件
toolCallsList := make([]map[string]interface{}, 0, len(toolCallMap))
for i := 0; i < len(toolCallMap)+1; i++ {
	tc, ok := toolCallMap[i]
	if !ok {
		continue
	}
	argsStr, _ := tc["argsStr"].(string)
	var argsObj interface{}
	err2 := json.Unmarshal([]byte(argsStr), &argsObj)
	log.Printf("[AIService] tool_call[%d] name=%v argsStr=%q parseErr=%v session=%s",
		i, tc["name"], argsStr, err2, req.SessionID)
	if err2 != nil {
		// argsStr 截断/损坏，跳过该 tool_call 避免前端用残缺参数执行
		log.Printf("[AIService] tool_call[%d] 跳过（argsStr 解析失败） session=%s", i, req.SessionID)
		continue
	}
	toolCallsList = append(toolCallsList, map[string]interface{}{
		"id":   tc["id"],
		"name": tc["name"],
		"args": argsObj,
	})
}

// ── Fallback：模型未走标准 tool_calls 格式，尝试从纯文本 content 中解析工具调用 ──
// 此 RKNN 服务有时会把工具调用以纯文本输出，需要 fallback 解析
// 支持两种格式：
//   1. 自然语言格式：我执行了 {tool}，参数：{JSON}
//   2. XML 格式：    <function name="{tool}" arguments='{JSON}'/>
if len(toolCallsList) == 0 {
	if parsed := parseTextToolCalls(fullContent.String()); len(parsed) > 0 {
		log.Printf("[AIService] fallback 解析到 %d 个文本工具调用 session=%s", len(parsed), req.SessionID)
		toolCallsList = parsed
	}
}

// 打印 usage 统计
if usage != nil {
	log.Printf("[AIService] usage: prompt_tokens=%v completion_tokens=%v total_tokens=%v session=%s",
		usage["prompt_tokens"], usage["completion_tokens"], usage["total_tokens"], req.SessionID)
}

// 发送完成事件
a.emitEvent("ai:done:"+req.SessionID, map[string]interface{}{
	"content":    fullContent.String(),
	"reasoning":  fullReasoning.String(),
	"usage":      usage,
	"tool_calls": toolCallsList,
})
	return nil
}

// parseTextToolCalls 从模型的纯文本输出中解析工具调用（fallback）
// 支持格式：
//  1. 自然语言：我执行了 {tool}，参数：{JSON}
//  2. XML：     <function name="{tool}" arguments='{JSON}'/>
//  3. 代码块：  ```xml\n<function name="{tool}" arguments='{JSON}'/>
func parseTextToolCalls(content string) []map[string]interface{} {
	if content == "" || content == "None" {
		return nil
	}

	results := []map[string]interface{}{}
	idBase := fmt.Sprintf("txt_%d", time.Now().UnixMilli())

	// 已知工具名集合（用于验证解析结果）
	knownTools := map[string]bool{
		"get_screen_context": true, "open_app": true, "click": true,
		"swipe": true, "send_text": true, "key_press": true,
		"stop_app": true, "exec_cmd": true, "ask_user": true, "task_done": true,
	}

	// 格式1：XML <function name="..." arguments='...'/>
	// 也处理代码块包裹的情况
	xmlRe := strings.NewReplacer("```xml", "", "```", "")
	cleaned := xmlRe.Replace(content)

	// 用字符串查找解析 XML 格式
	remaining := cleaned
	for {
		start := strings.Index(remaining, "<function")
		if start < 0 {
			break
		}
		end := strings.Index(remaining[start:], "/>")
		if end < 0 {
			break
		}
		tag := remaining[start : start+end+2]
		remaining = remaining[start+end+2:]

		// 提取 name
		name := ""
		if ni := strings.Index(tag, `name="`); ni >= 0 {
			rest := tag[ni+6:]
			if ei := strings.Index(rest, `"`); ei >= 0 {
				name = rest[:ei]
			}
		}
		if !knownTools[name] {
			continue
		}

		// 提取 arguments（单引号或双引号）
		argsStr := ""
		for _, q := range []string{`arguments='`, `arguments="`} {
			if ai := strings.Index(tag, q); ai >= 0 {
				rest := tag[ai+len(q):]
				closeQ := string(q[len(q)-1])
				if ei := strings.Index(rest, closeQ); ei >= 0 {
					argsStr = rest[:ei]
				}
				break
			}
		}

		var args interface{}
		if err := json.Unmarshal([]byte(argsStr), &args); err != nil {
			args = map[string]interface{}{}
		}
		log.Printf("[AIService] fallback XML tool: name=%s args=%v", name, args)
		results = append(results, map[string]interface{}{
			"id":   fmt.Sprintf("%s_%d", idBase, len(results)),
			"name": name,
			"args": args,
		})
	}
	if len(results) > 0 {
		return results
	}

	// 格式2：自然语言 "我执行了 {tool}，参数：{JSON}" 或 "执行了 {tool} 参数 {JSON}"
	// 用简单行扫描
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// 找工具名
		toolName := ""
		for t := range knownTools {
			if strings.Contains(line, t) {
				toolName = t
				break
			}
		}
		if toolName == "" {
			continue
		}
		// 找 JSON 对象
		jsonStart := strings.Index(line, "{")
		if jsonStart < 0 {
			// 无参数工具（如 get_screen_context）
			results = append(results, map[string]interface{}{
				"id":   fmt.Sprintf("%s_%d", idBase, len(results)),
				"name": toolName,
				"args": map[string]interface{}{},
			})
			log.Printf("[AIService] fallback text tool (no args): name=%s", toolName)
			continue
		}
		// 找匹配的右括号
		depth, end2 := 0, -1
		for i, c := range line[jsonStart:] {
			if c == '{' {
				depth++
			} else if c == '}' {
				depth--
				if depth == 0 {
					end2 = jsonStart + i + 1
					break
				}
			}
		}
		if end2 < 0 {
			end2 = len(line)
		}
		argsStr2 := line[jsonStart:end2]
		var args2 interface{}
		if err := json.Unmarshal([]byte(argsStr2), &args2); err != nil {
			args2 = map[string]interface{}{}
		}
		log.Printf("[AIService] fallback text tool: name=%s args=%v", toolName, args2)
		results = append(results, map[string]interface{}{
			"id":   fmt.Sprintf("%s_%d", idBase, len(results)),
			"name": toolName,
			"args": args2,
		})
	}

	return results
}

// AIStartModel 启动模型服务
func (a *App) AIStartModel(req AIModelStartRequest) map[string]interface{} {
	url := fmt.Sprintf("http://%s/lm/server/start", deviceAddr(req.DeviceIP))
	bodyBytes, err := json.Marshal(req.ModelConfig)
	if err != nil {
		log.Printf("[AIStartModel] JSON序列化失败: %v", err)
		return map[string]interface{}{"success": false, "message": err.Error()}
	}
	log.Printf("[AIStartModel] POST %s", url)
	log.Printf("[AIStartModel] 请求体: %s", string(bodyBytes))
	httpReq, err := a.buildDeviceRequest("POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		log.Printf("[AIStartModel] 构造请求失败: %v", err)
		return map[string]interface{}{"success": false, "message": err.Error()}
	}
	resp, err := a.httpClient.Do(httpReq)
	if err != nil {
		log.Printf("[AIStartModel] HTTP请求失败: %v", err)
		return map[string]interface{}{"success": false, "message": err.Error()}
	}
	defer resp.Body.Close()
	log.Printf("[AIStartModel] HTTP状态码: %d", resp.StatusCode)
	respBody, _ := io.ReadAll(resp.Body)
	log.Printf("[AIStartModel] 响应体: %s", string(respBody))
	var result map[string]interface{}
	json.Unmarshal(respBody, &result)
	if code, ok := result["code"]; ok && fmt.Sprintf("%v", code) == "0" {
		log.Printf("[AIStartModel] 启动成功")
		return map[string]interface{}{"success": true, "data": result}
	}
	msg, _ := result["message"].(string)
	log.Printf("[AIStartModel] 启动失败: code=%v msg=%s", result["code"], msg)
	return map[string]interface{}{"success": false, "message": msg, "data": result}
}

// AIStopModel 停止模型服务
func (a *App) AIStopModel(deviceIP string) map[string]interface{} {
	url := fmt.Sprintf("http://%s/lm/server/stop", deviceAddr(deviceIP))
	httpReq, err := a.buildDeviceRequest("POST", url, strings.NewReader("{}"))
	if err != nil {
		return map[string]interface{}{"success": false, "message": err.Error()}
	}
	resp, err := a.httpClient.Do(httpReq)
	if err != nil {
		return map[string]interface{}{"success": false, "message": err.Error()}
	}
	defer resp.Body.Close()
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	if code, ok := result["code"]; ok && fmt.Sprintf("%v", code) == "0" {
		return map[string]interface{}{"success": true}
	}
	msg, _ := result["message"].(string)
	return map[string]interface{}{"success": false, "message": msg}
}

// AIGetModels 查询设备当前运行的模型列表
func (a *App) AIGetModels(deviceIP string) map[string]interface{} {
	url := fmt.Sprintf("http://%s/lm/models", deviceAddr(deviceIP))
	httpReq, err := a.buildDeviceRequest("GET", url, nil)
	if err != nil {
		log.Printf("[AIGetModels] 构造请求失败: %v", err)
		return map[string]interface{}{"success": false, "message": err.Error()}
	}
	resp, err := a.httpClient.Do(httpReq)
	if err != nil {
		log.Printf("[AIGetModels] HTTP请求失败: %v", err)
		return map[string]interface{}{"success": false, "message": err.Error()}
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	log.Printf("[AIGetModels] 响应: %s", string(respBody))
	var result map[string]interface{}
	json.Unmarshal(respBody, &result)
	return map[string]interface{}{"success": true, "data": result}
}

// AIGetModelFiles 获取设备上的模型文件列表
func (a *App) AIGetModelFiles(deviceIP string) map[string]interface{} {
	url := fmt.Sprintf("http://%s/lm/modelfiles", deviceAddr(deviceIP))
	httpReq, err := a.buildDeviceRequest("GET", url, nil)
	if err != nil {
		return map[string]interface{}{"success": false, "message": err.Error()}
	}
	resp, err := a.httpClient.Do(httpReq)
	if err != nil {
		return map[string]interface{}{"success": false, "message": err.Error()}
	}
	defer resp.Body.Close()
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	return map[string]interface{}{"success": true, "data": result}
}

// AIResetDevice 重置设备 AI 状态
// 步骤1：GET /lm/info 获取 device_id
// 步骤2：POST /lm/reset { "type":"hw", "device_id": <id> }
func (a *App) AIResetDevice(deviceIP string) map[string]interface{} {
	// device_id 默认传 0
	resetURL := fmt.Sprintf("http://%s/lm/reset", deviceAddr(deviceIP))
	resetBody := `{"type":"hw","device_id":0}`
	log.Printf("[AIResetDevice] POST %s body: %s", resetURL, resetBody)
	resetReq, err := a.buildDeviceRequest("POST", resetURL, strings.NewReader(resetBody))
	if err != nil {
		return map[string]interface{}{"success": false, "message": err.Error()}
	}
	resetResp, err := a.httpClient.Do(resetReq)
	if err != nil {
		return map[string]interface{}{"success": false, "message": "重置请求失败: " + err.Error()}
	}
	defer resetResp.Body.Close()
	respBody, _ := io.ReadAll(resetResp.Body)
	log.Printf("[AIResetDevice] /lm/reset HTTP %d 响应: %s", resetResp.StatusCode, string(respBody))

	var result map[string]interface{}
	json.Unmarshal(respBody, &result)
	if code, ok := result["code"]; ok && fmt.Sprintf("%v", code) == "0" {
		return map[string]interface{}{"success": true}
	}
	msg, _ := result["msg"].(string)
	if msg == "" {
		msg, _ = result["message"].(string)
	}
	return map[string]interface{}{"success": false, "message": msg}
}

// AIGetDeviceInfo 查询算力棒状态（/lm/info），返回在线状态和内存信息
// online: true/false，chips: 芯片列表（含内存信息）
func (a *App) AIGetDeviceInfo(deviceIP string) map[string]interface{} {
	infoURL := fmt.Sprintf("http://%s/lm/info", deviceAddr(deviceIP))
	httpReq, err := a.buildDeviceRequest("GET", infoURL, nil)
	if err != nil {
		return map[string]interface{}{"success": false, "online": false, "message": err.Error()}
	}
	resp, err := a.httpClient.Do(httpReq)
	if err != nil {
		return map[string]interface{}{"success": false, "online": false, "message": err.Error()}
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)

	var raw struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			FirmwareVersion   string `json:"firmware_version"`
			NsmiVersion       string `json:"nsmi_version"`
			PcieDriverVersion string `json:"pcie_driver_version"`
			Devices           []struct {
				DeviceID int    `json:"device_id"`
				Mode     string `json:"mode"`
				Chips    []struct {
					ChipID   int    `json:"chip_id"`
					ChipName string `json:"chip_name"`
					Temp     int    `json:"temp"`
					Health   int    `json:"health"`
					MemoryInfo struct {
						FreeSize  int64   `json:"free_size"`
						TotalSize int64   `json:"total_size"`
						UtilRate  float64 `json:"util_rate"`
					} `json:"memory_info"`
				} `json:"chips"`
			} `json:"devices"`
		} `json:"data"`
	}

	if err := json.Unmarshal(respBody, &raw); err != nil {
		return map[string]interface{}{"success": false, "online": false, "message": "解析响应失败"}
	}

	// devices 为空 → 算力棒离线
	if raw.Code != 0 || len(raw.Data.Devices) == 0 {
		return map[string]interface{}{"success": true, "online": false}
	}

	// 整理芯片内存信息
	type ChipInfo struct {
		ChipID    int     `json:"chip_id"`
		ChipName  string  `json:"chip_name"`
		Temp      int     `json:"temp"`
		FreeSize  int64   `json:"free_size"`
		TotalSize int64   `json:"total_size"`
		UsedSize  int64   `json:"used_size"`
		UtilRate  float64 `json:"util_rate"`
	}
	var chips []ChipInfo
	for _, dev := range raw.Data.Devices {
		for _, c := range dev.Chips {
			chips = append(chips, ChipInfo{
				ChipID:    c.ChipID,
				ChipName:  c.ChipName,
				Temp:      c.Temp,
				FreeSize:  c.MemoryInfo.FreeSize,
				TotalSize: c.MemoryInfo.TotalSize,
				UsedSize:  c.MemoryInfo.TotalSize - c.MemoryInfo.FreeSize,
				UtilRate:  c.MemoryInfo.UtilRate,
			})
		}
	}

	return map[string]interface{}{
		"success":          true,
		"online":           true,
		"firmware_version": raw.Data.FirmwareVersion,
		"nsmi_version":     raw.Data.NsmiVersion,
		"chips":            chips,
	}
}

