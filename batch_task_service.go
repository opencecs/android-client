package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

// Target 批量任务目标（设备+容器）
type Target struct {
	DeviceIP      string `json:"device_ip"`
	ContainerID   string `json:"container_id"`
	ContainerName string `json:"container_name"`
	DeviceVersion string `json:"device_version"` // "v0", "v1", "v2", "v3"
	Password      string `json:"password,omitempty"`
}

// BatchTask 批量任务
type BatchTask struct {
	ID             string                 `json:"id"`
	Name           string                 `json:"name"`
	Command        string                 `json:"command"`
	Targets        []Target               `json:"targets"`
	ScheduleType   string                 `json:"schedule_type"` // "once", "periodic", "cron", "immediate"
	ScheduleConfig map[string]interface{} `json:"schedule_config"`
	Enabled        bool                   `json:"enabled"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
	LastRun        *time.Time             `json:"last_run,omitempty"`
	NextRun        *time.Time             `json:"next_run,omitempty"`
	CronEntryID    cron.EntryID           `json:"-"` // 不序列化
}

// CommandTemplate 命令模板
type CommandTemplate struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Command     string    `json:"command"`
	Category    string    `json:"category"`
	Variables   []string  `json:"variables"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ExecutionHistory 执行历史
type ExecutionHistory struct {
	ID        string            `json:"id"`
	TaskID    string            `json:"task_id,omitempty"`
	TaskName  string            `json:"task_name"`
	Command   string            `json:"command"`
	Targets   []Target          `json:"targets"`
	Results   []ExecutionResult `json:"results"`
	Status    string            `json:"status"` // "success", "partial", "failed"
	StartTime time.Time         `json:"start_time"`
	EndTime   time.Time         `json:"end_time"`
	Duration  int64             `json:"duration"` // 毫秒
}

// ExecutionResult 单个目标的执行结果
type ExecutionResult struct {
	DeviceIP    string `json:"device_ip"`
	ContainerID string `json:"container_id"`
	Success     bool   `json:"success"`
	Output      string `json:"output"`
	Error       string `json:"error,omitempty"`
	Duration    int64  `json:"duration"` // 毫秒
	ExitCode    int    `json:"exit_code"`
}

// BatchTaskService 批量任务服务
type BatchTaskService struct {
	app          *App
	scheduler    *cron.Cron
	tasks        map[string]*BatchTask
	templates    map[string]*CommandTemplate
	history      []ExecutionHistory
	dataDir      string
	tasksMutex   sync.RWMutex
	templatesMutex sync.RWMutex
	historyMutex sync.RWMutex
}

// InitBatchTaskService 初始化批量任务服务
func (s *BatchTaskService) InitBatchTaskService(app *App, dataDir string) error {
	s.app = app
	s.dataDir = dataDir
	s.tasks = make(map[string]*BatchTask)
	s.templates = make(map[string]*CommandTemplate)
	s.history = []ExecutionHistory{}

	// 确保数据目录存在
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("创建数据目录失败: %w", err)
	}

	// 初始化cron调度器
	s.scheduler = cron.New()
	s.scheduler.Start()

	// 加载持久化数据
	if err := s.loadTasksFromFile(); err != nil {
		app.log.Warn("加载批量任务失败: %v", err)
	}
	if err := s.loadTemplatesFromFile(); err != nil {
		app.log.Warn("加载命令模板失败: %v", err)
	}
	if err := s.loadHistoryFromFile(); err != nil {
		app.log.Warn("加载执行历史失败: %v", err)
	}

	// 恢复定时任务到cron调度器
	s.tasksMutex.RLock()
	for _, task := range s.tasks {
		if task.Enabled && task.ScheduleType != "immediate" {
			if err := s.scheduleTask(task); err != nil {
				app.log.Warn("恢复定时任务失败 [%s]: %v", task.ID, err)
			}
		}
	}
	s.tasksMutex.RUnlock()

	app.log.Info("批量任务服务初始化完成")
	return nil
}

// ExecuteBatchCommand 立即批量执行命令
func (s *BatchTaskService) ExecuteBatchCommand(targets []Target, command string, taskName string) (*ExecutionHistory, error) {
	// 创建执行历史记录
	history := &ExecutionHistory{
		ID:        fmt.Sprintf("exec_%d", time.Now().UnixNano()),
		TaskName:  taskName,
		Command:   command,
		Targets:   targets,
		Results:   make([]ExecutionResult, 0),
		StartTime: time.Now(),
	}

	s.app.log.Info("开始批量执行命令: %s, 目标数量: %d", taskName, len(targets))

	// 并发执行（限制并发数）
	maxConcurrent := 10
	semaphore := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup
	var resultsMutex sync.Mutex

	for _, target := range targets {
		wg.Add(1)
		semaphore <- struct{}{} // 获取信号量

		go func(t Target) {
			defer wg.Done()
			defer func() { <-semaphore }() // 释放信号量

			startTime := time.Now()

			// 变量替换
			finalCommand := s.replaceVariables(command, t)

			// 确定Docker端口
			dockerPort := 2375
			if t.DeviceVersion == "v3" {
				dockerPort = 8000
			}

			// 执行命令
			result, err := dockerExec(
				t.DeviceIP,
				dockerPort,
				t.ContainerID,
				[]string{"sh", "-c", fmt.Sprintf("sd -c \"%s\"", finalCommand)},
				t.Password,
				t.DeviceVersion,
			)

			// 构建执行结果
			execResult := ExecutionResult{
				DeviceIP:    t.DeviceIP,
				ContainerID: t.ContainerID,
				Success:     err == nil && result.ExitCode == 0,
				Output:      result.Stdout,
				ExitCode:    result.ExitCode,
				Duration:    time.Since(startTime).Milliseconds(),
			}

			if err != nil {
				execResult.Error = err.Error()
			} else if result.ExitCode != 0 {
				execResult.Error = result.Stderr
			}

			// 保存执行结果
			resultsMutex.Lock()
			history.Results = append(history.Results, execResult)
			resultsMutex.Unlock()

			// 推送进度事件
			s.app.emitEvent("batch-task:progress", map[string]interface{}{
				"historyID":   history.ID,
				"completed":   len(history.Results),
				"total":       len(targets),
				"deviceIP":    t.DeviceIP,
				"containerID": t.ContainerID,
				"success":     execResult.Success,
				"output":      execResult.Output,
				"error":       execResult.Error,
			})
		}(target)
	}

	// 等待所有任务完成
	wg.Wait()

	history.EndTime = time.Now()
	history.Duration = history.EndTime.Sub(history.StartTime).Milliseconds()

	// 判断整体状态
	successCount := 0
	for _, r := range history.Results {
		if r.Success {
			successCount++
		}
	}

	if successCount == len(targets) {
		history.Status = "success"
	} else if successCount == 0 {
		history.Status = "failed"
	} else {
		history.Status = "partial"
	}

	// 保存到历史记录
	s.addExecutionHistory(history)

	s.app.log.Info("批量命令执行完成: %s, 成功: %d/%d", taskName, successCount, len(targets))

	// 推送完成事件
	s.app.emitEvent("batch-task:complete", map[string]interface{}{
		"historyID": history.ID,
		"status":    history.Status,
		"success":   successCount,
		"failed":    len(targets) - successCount,
		"total":     len(targets),
	})

	return history, nil
}

// CreateScheduledTask 创建定时任务
func (s *BatchTaskService) CreateScheduledTask(task *BatchTask) error {
	s.tasksMutex.Lock()
	defer s.tasksMutex.Unlock()

	// 生成任务ID
	if task.ID == "" {
		task.ID = fmt.Sprintf("task_%d", time.Now().UnixNano())
	}

	// 设置时间戳
	now := time.Now()
	task.CreatedAt = now
	task.UpdatedAt = now

	// 添加到调度器
	if task.Enabled && task.ScheduleType != "immediate" {
		if err := s.scheduleTask(task); err != nil {
			return fmt.Errorf("添加到调度器失败: %w", err)
		}
	}

	// 保存到内存
	s.tasks[task.ID] = task

	// 持久化
	if err := s.saveTasksToFile(); err != nil {
		s.app.log.Warn("保存任务到文件失败: %v", err)
	}

	s.app.log.Info("创建定时任务成功: %s (%s)", task.Name, task.ID)
	return nil
}

// UpdateScheduledTask 更新定时任务
func (s *BatchTaskService) UpdateScheduledTask(task *BatchTask) error {
	s.tasksMutex.Lock()
	defer s.tasksMutex.Unlock()

	oldTask, exists := s.tasks[task.ID]
	if !exists {
		return fmt.Errorf("任务不存在: %s", task.ID)
	}

	// 移除旧的cron任务
	if oldTask.CronEntryID > 0 {
		s.scheduler.Remove(oldTask.CronEntryID)
	}

	// 更新时间戳
	task.UpdatedAt = time.Now()
	task.CreatedAt = oldTask.CreatedAt // 保持创建时间不变

	// 添加新的cron任务
	if task.Enabled && task.ScheduleType != "immediate" {
		if err := s.scheduleTask(task); err != nil {
			return fmt.Errorf("更新调度失败: %w", err)
		}
	}

	// 更新内存
	s.tasks[task.ID] = task

	// 持久化
	if err := s.saveTasksToFile(); err != nil {
		s.app.log.Warn("保存任务到文件失败: %v", err)
	}

	s.app.log.Info("更新定时任务成功: %s", task.ID)
	return nil
}

// DeleteScheduledTask 删除定时任务
func (s *BatchTaskService) DeleteScheduledTask(taskID string) error {
	s.tasksMutex.Lock()
	defer s.tasksMutex.Unlock()

	task, exists := s.tasks[taskID]
	if !exists {
		return fmt.Errorf("任务不存在: %s", taskID)
	}

	// 从调度器移除
	if task.CronEntryID > 0 {
		s.scheduler.Remove(task.CronEntryID)
	}

	// 从内存删除
	delete(s.tasks, taskID)

	// 持久化
	if err := s.saveTasksToFile(); err != nil {
		s.app.log.Warn("保存任务到文件失败: %v", err)
	}

	s.app.log.Info("删除定时任务成功: %s", taskID)
	return nil
}

// GetScheduledTasks 获取所有定时任务
func (s *BatchTaskService) GetScheduledTasks() []*BatchTask {
	s.tasksMutex.RLock()
	defer s.tasksMutex.RUnlock()

	tasks := make([]*BatchTask, 0, len(s.tasks))
	for _, task := range s.tasks {
		tasks = append(tasks, task)
	}

	return tasks
}

// GetScheduledTask 获取单个定时任务
func (s *BatchTaskService) GetScheduledTask(taskID string) (*BatchTask, error) {
	s.tasksMutex.RLock()
	defer s.tasksMutex.RUnlock()

	task, exists := s.tasks[taskID]
	if !exists {
		return nil, fmt.Errorf("任务不存在: %s", taskID)
	}

	return task, nil
}

// ExecuteScheduledTask 执行定时任务（由cron调度器调用）
func (s *BatchTaskService) ExecuteScheduledTask(taskID string) {
	s.tasksMutex.RLock()
	task, exists := s.tasks[taskID]
	s.tasksMutex.RUnlock()

	if !exists || !task.Enabled {
		return
	}

	s.app.log.Info("定时任务触发: %s (%s)", task.Name, taskID)

	// 更新最后运行时间
	now := time.Now()
	s.tasksMutex.Lock()
	task.LastRun = &now
	s.tasksMutex.Unlock()

	// 执行批量命令
	_, err := s.ExecuteBatchCommand(task.Targets, task.Command, task.Name)
	if err != nil {
		s.app.log.Error("定时任务执行失败 [%s]: %v", taskID, err)
	}

	// 如果是一次性任务，执行后禁用
	if task.ScheduleType == "once" {
		s.tasksMutex.Lock()
		task.Enabled = false
		s.scheduler.Remove(task.CronEntryID)
		task.CronEntryID = 0
		s.tasksMutex.Unlock()

		if err := s.saveTasksToFile(); err != nil {
			s.app.log.Warn("保存任务状态失败: %v", err)
		}
	}
}

// SaveCommandTemplate 保存命令模板
func (s *BatchTaskService) SaveCommandTemplate(template *CommandTemplate) error {
	s.templatesMutex.Lock()
	defer s.templatesMutex.Unlock()

	// 生成模板ID
	if template.ID == "" {
		template.ID = fmt.Sprintf("tpl_%d", time.Now().UnixNano())
		template.CreatedAt = time.Now()
	}
	template.UpdatedAt = time.Now()

	// 保存到内存
	s.templates[template.ID] = template

	// 持久化
	if err := s.saveTemplatesToFile(); err != nil {
		return fmt.Errorf("保存模板到文件失败: %w", err)
	}

	s.app.log.Info("保存命令模板成功: %s (%s)", template.Name, template.ID)
	return nil
}

// GetCommandTemplates 获取所有命令模板
func (s *BatchTaskService) GetCommandTemplates() []*CommandTemplate {
	s.templatesMutex.RLock()
	defer s.templatesMutex.RUnlock()

	templates := make([]*CommandTemplate, 0, len(s.templates))
	for _, tpl := range s.templates {
		templates = append(templates, tpl)
	}

	return templates
}

// DeleteCommandTemplate 删除命令模板
func (s *BatchTaskService) DeleteCommandTemplate(templateID string) error {
	s.templatesMutex.Lock()
	defer s.templatesMutex.Unlock()

	if _, exists := s.templates[templateID]; !exists {
		return fmt.Errorf("模板不存在: %s", templateID)
	}

	delete(s.templates, templateID)

	if err := s.saveTemplatesToFile(); err != nil {
		return fmt.Errorf("保存模板到文件失败: %w", err)
	}

	s.app.log.Info("删除命令模板成功: %s", templateID)
	return nil
}

// GetExecutionHistory 获取执行历史（支持分页）
func (s *BatchTaskService) GetExecutionHistory(limit, offset int) ([]ExecutionHistory, int, error) {
	s.historyMutex.RLock()
	defer s.historyMutex.RUnlock()

	total := len(s.history)

	// 边界检查
	if offset >= total {
		return []ExecutionHistory{}, total, nil
	}

	end := offset + limit
	if end > total {
		end = total
	}

	// 返回逆序（最新的在前面）
	result := make([]ExecutionHistory, 0, end-offset)
	for i := total - 1 - offset; i >= total-end && i >= 0; i-- {
		result = append(result, s.history[i])
	}

	return result, total, nil
}

// ExportExecutionHistory 导出执行历史
func (s *BatchTaskService) ExportExecutionHistory(format string) (string, error) {
	s.historyMutex.RLock()
	defer s.historyMutex.RUnlock()

	switch strings.ToLower(format) {
	case "json":
		data, err := json.MarshalIndent(s.history, "", "  ")
		if err != nil {
			return "", fmt.Errorf("JSON序列化失败: %w", err)
		}

		// 保存到临时文件
		filename := fmt.Sprintf("batch_history_export_%d.json", time.Now().Unix())
		filepath := filepath.Join(s.dataDir, filename)
		if err := os.WriteFile(filepath, data, 0644); err != nil {
			return "", fmt.Errorf("写入文件失败: %w", err)
		}

		return filepath, nil

	case "csv":
		// CSV格式导出
		var csv strings.Builder
		csv.WriteString("ID,Task Name,Command,Status,Start Time,Duration(ms),Success,Failed,Total\n")

		for _, h := range s.history {
			successCount := 0
			failedCount := 0
			for _, r := range h.Results {
				if r.Success {
					successCount++
				} else {
					failedCount++
				}
			}

			csv.WriteString(fmt.Sprintf("%s,%s,%s,%s,%s,%d,%d,%d,%d\n",
				h.ID,
				h.TaskName,
				strings.ReplaceAll(h.Command, ",", ";"),
				h.Status,
				h.StartTime.Format("2006-01-02 15:04:05"),
				h.Duration,
				successCount,
				failedCount,
				len(h.Results),
			))
		}

		// 保存到临时文件
		filename := fmt.Sprintf("batch_history_export_%d.csv", time.Now().Unix())
		filepath := filepath.Join(s.dataDir, filename)
		if err := os.WriteFile(filepath, []byte(csv.String()), 0644); err != nil {
			return "", fmt.Errorf("写入文件失败: %w", err)
		}

		return filepath, nil

	default:
		return "", fmt.Errorf("不支持的导出格式: %s", format)
	}
}

// replaceVariables 变量替换
func (s *BatchTaskService) replaceVariables(command string, target Target) string {
	replacements := map[string]string{
		"{device_ip}":    target.DeviceIP,
		"{container_id}": target.ContainerID,
		"{container_short_id}": func() string {
			if len(target.ContainerID) > 12 {
				return target.ContainerID[:12]
			}
			return target.ContainerID
		}(),
		"{container_name}": target.ContainerName,
		"{timestamp}":      fmt.Sprintf("%d", time.Now().Unix()),
		"{date}":           time.Now().Format("2006-01-02"),
		"{time}":           time.Now().Format("15:04:05"),
	}

	result := command
	for key, value := range replacements {
		result = strings.ReplaceAll(result, key, value)
	}

	return result
}

// scheduleTask 将任务添加到cron调度器
func (s *BatchTaskService) scheduleTask(task *BatchTask) error {
	var cronExpr string

	switch task.ScheduleType {
	case "once":
		// 一次性任务：解析时间戳
		timestamp, ok := task.ScheduleConfig["timestamp"].(float64)
		if !ok {
			return fmt.Errorf("无效的时间戳配置")
		}
		executeTime := time.Unix(int64(timestamp), 0)

		// 如果已经过期，不添加
		if executeTime.Before(time.Now()) {
			task.Enabled = false
			return fmt.Errorf("任务执行时间已过期")
		}

		// 计算延迟执行
		delay := time.Until(executeTime)
		task.NextRun = &executeTime

		// 使用定时器而非cron
		go func() {
			time.Sleep(delay)
			s.ExecuteScheduledTask(task.ID)
		}()

		return nil

	case "periodic":
		// 周期性任务：转换为cron表达式
		interval, ok := task.ScheduleConfig["interval"].(string)
		if !ok {
			return fmt.Errorf("无效的周期配置")
		}

		switch interval {
		case "daily":
			hour := int(task.ScheduleConfig["hour"].(float64))
			minute := int(task.ScheduleConfig["minute"].(float64))
			cronExpr = fmt.Sprintf("%d %d * * *", minute, hour)
		case "weekly":
			hour := int(task.ScheduleConfig["hour"].(float64))
			minute := int(task.ScheduleConfig["minute"].(float64))
			weekday := int(task.ScheduleConfig["weekday"].(float64))
			cronExpr = fmt.Sprintf("%d %d * * %d", minute, hour, weekday)
		case "custom":
			// 自定义间隔（分钟）
			minutes := int(task.ScheduleConfig["minutes"].(float64))
			cronExpr = fmt.Sprintf("*/%d * * * *", minutes)
		default:
			return fmt.Errorf("不支持的周期类型: %s", interval)
		}

	case "cron":
		// Cron表达式
		expr, ok := task.ScheduleConfig["expression"].(string)
		if !ok {
			return fmt.Errorf("无效的cron表达式")
		}
		cronExpr = expr

	default:
		return fmt.Errorf("不支持的调度类型: %s", task.ScheduleType)
	}

	// 添加到cron调度器
	entryID, err := s.scheduler.AddFunc(cronExpr, func() {
		s.ExecuteScheduledTask(task.ID)
	})

	if err != nil {
		return fmt.Errorf("添加cron任务失败: %w", err)
	}

	task.CronEntryID = entryID

	// 计算下次运行时间
	entry := s.scheduler.Entry(entryID)
	nextRun := entry.Next
	task.NextRun = &nextRun

	return nil
}

// addExecutionHistory 添加执行历史
func (s *BatchTaskService) addExecutionHistory(history *ExecutionHistory) {
	s.historyMutex.Lock()
	defer s.historyMutex.Unlock()

	s.history = append(s.history, *history)

	// 限制历史记录数量（最多1000条）
	if len(s.history) > 1000 {
		s.history = s.history[len(s.history)-1000:]
	}

	// 持久化
	if err := s.saveHistoryToFile(); err != nil {
		s.app.log.Warn("保存执行历史失败: %v", err)
	}
}

// 文件持久化方法

func (s *BatchTaskService) saveTasksToFile() error {
	filepath := filepath.Join(s.dataDir, "batch_tasks.json")

	taskList := make([]*BatchTask, 0, len(s.tasks))
	for _, task := range s.tasks {
		taskList = append(taskList, task)
	}

	data, err := json.MarshalIndent(taskList, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath, data, 0644)
}

func (s *BatchTaskService) loadTasksFromFile() error {
	filepath := filepath.Join(s.dataDir, "batch_tasks.json")

	data, err := os.ReadFile(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // 文件不存在不算错误
		}
		return err
	}

	var taskList []*BatchTask
	if err := json.Unmarshal(data, &taskList); err != nil {
		return err
	}

	for _, task := range taskList {
		s.tasks[task.ID] = task
	}

	return nil
}

func (s *BatchTaskService) saveTemplatesToFile() error {
	filepath := filepath.Join(s.dataDir, "command_templates.json")

	templateList := make([]*CommandTemplate, 0, len(s.templates))
	for _, tpl := range s.templates {
		templateList = append(templateList, tpl)
	}

	data, err := json.MarshalIndent(templateList, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath, data, 0644)
}

func (s *BatchTaskService) loadTemplatesFromFile() error {
	filepath := filepath.Join(s.dataDir, "command_templates.json")

	data, err := os.ReadFile(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var templateList []*CommandTemplate
	if err := json.Unmarshal(data, &templateList); err != nil {
		return err
	}

	for _, tpl := range templateList {
		s.templates[tpl.ID] = tpl
	}

	return nil
}

func (s *BatchTaskService) saveHistoryToFile() error {
	filepath := filepath.Join(s.dataDir, "execution_history.json")

	data, err := json.MarshalIndent(s.history, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath, data, 0644)
}

func (s *BatchTaskService) loadHistoryFromFile() error {
	filepath := filepath.Join(s.dataDir, "execution_history.json")

	data, err := os.ReadFile(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	if err := json.Unmarshal(data, &s.history); err != nil {
		return err
	}

	return nil
}

// Shutdown 关闭服务
func (s *BatchTaskService) Shutdown() {
	if s.scheduler != nil {
		s.scheduler.Stop()
	}
	s.app.log.Info("批量任务服务已关闭")
}

// ==================== 批量导入云机备份功能 ====================

// BackupFile 备份文件信息
type BackupFile struct {
	Name string `json:"name"` // 文件名
	Size int64  `json:"size"` // 文件大小(字节)
	Path string `json:"path"` // 完整路径
}

// BatchImportTask 批量导入任务
type BatchImportTask struct {
	ID             string                 `json:"id"`
	BackupFileName string                 `json:"backup_file_name"`
	Devices        []DeviceSlotConfig     `json:"devices"`
	Status         string                 `json:"status"` // "pending", "running", "completed", "failed"
	Progress       BatchImportProgress    `json:"progress"`
	StartTime      time.Time              `json:"start_time"`
	EndTime        *time.Time             `json:"end_time,omitempty"`
}

// DeviceSlotConfig 设备坑位配置
type DeviceSlotConfig struct {
	DeviceIP    string      `json:"device_ip"`
	DeviceType  string      `json:"device_type"` // "12slots" or "24slots"
	SlotConfigs []SlotConfig `json:"slot_configs"`
}

// SlotConfig 坑位配置
type SlotConfig struct {
	SlotNumber int    `json:"slot_number"` // 坑位号 (1-12 或 1-24)
	CopyCount  int    `json:"copy_count"`  // 复制份数
	MachineName string `json:"machine_name"` // 云机名称
}

// BatchImportProgress 批量导入进度
type BatchImportProgress struct {
	TotalTasks     int                      `json:"total_tasks"`      // 总任务数
	CompletedTasks int                      `json:"completed_tasks"`  // 已完成任务数
	FailedTasks    int                      `json:"failed_tasks"`     // 失败任务数
	CurrentDevice  string                   `json:"current_device"`   // 当前处理的设备
	CurrentSlot    int                      `json:"current_slot"`     // 当前处理的坑位
	Details        []BatchImportTaskDetail  `json:"details"`          // 每个任务的详细结果
}

// BatchImportTaskDetail 批量导入任务详情
type BatchImportTaskDetail struct {
	DeviceIP    string `json:"device_ip"`
	SlotNumber  int    `json:"slot_number"`
	MachineName string `json:"machine_name"`
	Success     bool   `json:"success"`
	Message     string `json:"message"`
	Duration    int64  `json:"duration"` // 毫秒
}

// ListBackupFiles 列出所有备份文件
func (s *BatchTaskService) ListBackupFiles() ([]BackupFile, error) {
	backupDir := filepath.Join(getStorageBaseDir(), "cloudMachineBackup")
	s.app.log.Info("扫描备份目录: %s", backupDir)
	
	// 确保目录存在
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return nil, fmt.Errorf("创建备份目录失败: %w", err)
	}

	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return nil, fmt.Errorf("读取备份目录失败: %w", err)
	}

	s.app.log.Info("目录中共有 %d 个文件/文件夹", len(entries))

	var backupFiles []BackupFile
	for _, entry := range entries {
		if entry.IsDir() {
			s.app.log.Info("跳过目录: %s", entry.Name())
			continue
		}
		
		// 只显示 .tar.gz 文件
		fileName := entry.Name()
		if !strings.HasSuffix(strings.ToLower(fileName), ".tar.gz") {
			s.app.log.Info("跳过非 .tar.gz 文件: %s", fileName)
			continue
		}

		info, err := entry.Info()
		if err != nil {
			s.app.log.Warn("获取文件信息失败 [%s]: %v", fileName, err)
			continue
		}

		backupFiles = append(backupFiles, BackupFile{
			Name: fileName,
			Size: info.Size(),
			Path: filepath.Join(backupDir, fileName),
		})
		
		s.app.log.Info("找到备份文件: %s (大小: %d 字节)", fileName, info.Size())
	}

	s.app.log.Info("共找到 %d 个备份文件", len(backupFiles))
	return backupFiles, nil
}

// DeleteBackupFile 删除备份文件
func (s *BatchTaskService) DeleteBackupFile(fileName string) error {
	backupDir := filepath.Join(getStorageBaseDir(), "cloudMachineBackup")
	filePath := filepath.Join(backupDir, fileName)

	// 安全检查：确保文件路径在备份目录内
	if !strings.HasPrefix(filePath, backupDir) {
		return fmt.Errorf("非法文件路径")
	}

	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("删除文件失败: %w", err)
	}

	s.app.log.Info("删除备份文件成功: %s", fileName)
	return nil
}

// GenerateMachineName 生成云机名称
// 规则: {时间戳毫秒}_{坑位号}_TXXXX
// 例如: 1770736461614_1_T0001
func GenerateMachineName(deviceIP string, slotNumber int) string {
	// 生成唯一时间戳（毫秒级）
	timestamp := time.Now().UnixMilli()
	
	// 生成随机4位数字（0001-9999）
	randomNum := rand.Intn(9999) + 1
	
	return fmt.Sprintf("%d_%d_T%04d", timestamp, slotNumber, randomNum)
}

// StartBatchImport 开始批量导入
func (s *BatchTaskService) StartBatchImport(backupFileName string, devicesConfig []DeviceSlotConfig) (*BatchImportTask, error) {
	task := &BatchImportTask{
		ID:             fmt.Sprintf("import_%d", time.Now().UnixNano()),
		BackupFileName: backupFileName,
		Devices:        devicesConfig,
		Status:         "running",
		StartTime:      time.Now(),
	}

	// 计算总任务数
	totalTasks := 0
	for _, device := range devicesConfig {
		for _, slot := range device.SlotConfigs {
			totalTasks += slot.CopyCount
		}
	}

	task.Progress = BatchImportProgress{
		TotalTasks: totalTasks,
		Details:    make([]BatchImportTaskDetail, 0),
	}

	s.app.log.Info("开始批量导入任务: %s, 备份文件: %s, 总任务数: %d", task.ID, backupFileName, totalTasks)

	// 异步执行导入
	go s.executeBatchImport(task)

	return task, nil
}

// executeBatchImport 执行批量导入
func (s *BatchTaskService) executeBatchImport(task *BatchImportTask) {
	defer func() {
		endTime := time.Now()
		task.EndTime = &endTime
		
		if task.Progress.FailedTasks == 0 {
			task.Status = "completed"
		} else if task.Progress.CompletedTasks == 0 {
			task.Status = "failed"
		} else {
			task.Status = "completed" // 部分成功也算完成
		}

		s.app.log.Info("批量导入任务完成: %s, 成功: %d, 失败: %d", 
			task.ID, task.Progress.CompletedTasks, task.Progress.FailedTasks)
		
		s.app.log.Info("发送完成事件: status=%s, 总任务=%d, 成功=%d, 失败=%d",
			task.Status, task.Progress.TotalTasks, task.Progress.CompletedTasks, task.Progress.FailedTasks)
			
		// 序列化 details 数组
		detailsArray := make([]map[string]interface{}, len(task.Progress.Details))
		for i, d := range task.Progress.Details {
			detailsArray[i] = map[string]interface{}{
				"device_ip":    d.DeviceIP,
				"slot_number":  d.SlotNumber,
				"machine_name": d.MachineName,
				"success":      d.Success,
				"message":      d.Message,
				"duration":     d.Duration,
			}
		}
		
		// 发送完成事件（手动序列化确保字段名正确）
		s.app.emitEvent("batch-import:complete", map[string]interface{}{
			"taskID": task.ID,
			"status": task.Status,
			"progress": map[string]interface{}{
				"total_tasks":     task.Progress.TotalTasks,
				"completed_tasks": task.Progress.CompletedTasks,
				"failed_tasks":    task.Progress.FailedTasks,
				"current_device":  task.Progress.CurrentDevice,
				"current_slot":    task.Progress.CurrentSlot,
				"details":         detailsArray,
			},
		})
	}()

	// 遍历所有设备和坑位
	for _, deviceConfig := range task.Devices {
		task.Progress.CurrentDevice = deviceConfig.DeviceIP

		for _, slotConfig := range deviceConfig.SlotConfigs {
			task.Progress.CurrentSlot = slotConfig.SlotNumber

			// 执行复制操作
			for i := 0; i < slotConfig.CopyCount; i++ {
				startTime := time.Now()
				
				// 生成云机名称
				// 如果前端指定了名称，使用指定的名称（多份时添加后缀）
				// 否则每次都生成新的唯一名称（时间戳+随机数确保唯一）
				machineName := ""
				if slotConfig.MachineName != "" {
					// 前端指定了名称
					machineName = slotConfig.MachineName
					if slotConfig.CopyCount > 1 {
						machineName = fmt.Sprintf("%s_%d", machineName, i+1)
					}
				} else {
					// 自动生成名称（每次生成新的时间戳和随机数）
					machineName = GenerateMachineName(deviceConfig.DeviceIP, slotConfig.SlotNumber)
					// 确保毫秒级时间戳不同（避免太快导致时间戳相同）
					if i > 0 {
						time.Sleep(2 * time.Millisecond)
					}
				}

				s.app.log.Info("导入云机: 设备=%s, 坑位=%d, 名称=%s (%d/%d)", 
					deviceConfig.DeviceIP, slotConfig.SlotNumber, machineName, i+1, slotConfig.CopyCount)

				// 调用导入函数
				result := s.app.ImportBackupMachine(
					deviceConfig.DeviceIP,
					task.BackupFileName,
					machineName,
					slotConfig.SlotNumber,
				)

				duration := time.Since(startTime).Milliseconds()
				success := false
				message := "未知错误"

				if result != nil {
					if successVal, ok := result["success"].(bool); ok {
						success = successVal
					}
					if msgVal, ok := result["message"].(string); ok {
						message = msgVal
					}
				}

				// 记录详情
				detail := BatchImportTaskDetail{
					DeviceIP:    deviceConfig.DeviceIP,
					SlotNumber:  slotConfig.SlotNumber,
					MachineName: machineName,
					Success:     success,
					Message:     message,
					Duration:    duration,
				}
				task.Progress.Details = append(task.Progress.Details, detail)

				if success {
					task.Progress.CompletedTasks++
				} else {
					task.Progress.FailedTasks++
				}

			// 发送进度事件（手动序列化确保字段名正确）
			s.app.log.Info("发送进度事件: 设备=%s, 坑位=%d, 成功=%v, 已完成=%d/%d", 
				deviceConfig.DeviceIP, slotConfig.SlotNumber, success, 
				task.Progress.CompletedTasks, task.Progress.TotalTasks)
			
			// 序列化 details 数组
			detailsArray := make([]map[string]interface{}, len(task.Progress.Details))
			for i, d := range task.Progress.Details {
				detailsArray[i] = map[string]interface{}{
					"device_ip":    d.DeviceIP,
					"slot_number":  d.SlotNumber,
					"machine_name": d.MachineName,
					"success":      d.Success,
					"message":      d.Message,
					"duration":     d.Duration,
				}
			}
			
			s.app.log.Info("[Event] 准备发送 batch-import:progress 事件")
			s.app.log.Info("[Event] - totalTasks: %d", task.Progress.TotalTasks)
			s.app.log.Info("[Event] - completedTasks: %d", task.Progress.CompletedTasks)
			s.app.log.Info("[Event] - failedTasks: %d", task.Progress.FailedTasks)
			s.app.log.Info("[Event] - details数量: %d", len(detailsArray))
			
			s.app.emitEvent("batch-import:progress", map[string]interface{}{
				"taskID": task.ID,
				"progress": map[string]interface{}{
					"total_tasks":     task.Progress.TotalTasks,
					"completed_tasks": task.Progress.CompletedTasks,
					"failed_tasks":    task.Progress.FailedTasks,
					"current_device":  task.Progress.CurrentDevice,
					"current_slot":    task.Progress.CurrentSlot,
					"details":         detailsArray,
				},
				"detail": map[string]interface{}{
					"device_ip":    detail.DeviceIP,
					"slot_number":  detail.SlotNumber,
					"machine_name": detail.MachineName,
					"success":      detail.Success,
					"message":      detail.Message,
					"duration":     detail.Duration,
				},
			})

				// 短暂延迟，避免并发过高
				time.Sleep(500 * time.Millisecond)
			}
		}
	}
}
