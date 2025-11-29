package service

import (
	"AITaskSystem/model"
	"AITaskSystem/repository"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// DeepSeekConfig 配置结构
type DeepSeekConfig struct {
	APIKey   string
	Endpoint string
	Model    string
}

// AIService 负责智能解析
type AIService struct {
	Config DeepSeekConfig
	Repo   *repository.TaskRepository
}

// NewAIService 初始化
func NewAIService(repo *repository.TaskRepository) *AIService {
	//从环境变量获取 Key
	apiKey := os.Getenv("DEEPSEEK_API_KEY")

	if apiKey == "" {
		// 如果本地没有设置环境变量，代码将无法连接 AI。
		fmt.Println("Warning: DEEPSEEK_API_KEY environment variable is not set. AI features will fail.")
		apiKey = ""
	}

	return &AIService{
		Config: DeepSeekConfig{
			APIKey:   apiKey,
			Endpoint: "https://ark.cn-beijing.volces.com/api/v3/chat/completions",
			Model:    "deepseek-r1-250528",
		},
		Repo: repo,
	}
}

type AIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

type TaskAIInfo struct {
	DueDateStr     string  `json:"due_date"`
	EstimatedHours float64 `json:"estimated_hours"`
}

// ParseTaskFromInput 核心逻辑：解析单个任务
func (s *AIService) ParseTaskFromInput(input string) model.Task {
	task := model.Task{
		Title:       input,
		Status:      "Pending",
		Description: "AI 自动解析的任务",
	}
	task.CreatedAt = time.Now()

	// 1. 基础信息提取
	aiResult := s.extractTaskInfoWithAI(input)
	task.EstimatedHours = aiResult.EstimatedHours
	if aiResult.DueDateStr != "" {
		if parsedTime, err := time.Parse("2006-01-02 15:04:05", aiResult.DueDateStr); err == nil {
			task.DueDate = parsedTime
		}
	}

	// 2. 初始优先级计算
	task.Priority = s.calculatePriority(&task)

	return task
}

// SyncPrioritiesForDependencies [强化版] 依赖反向同步
func (s *AIService) SyncPrioritiesForDependencies(newTask *model.Task) {
	// 只有新任务比较重要时，才去提升别人
	if newTask.Priority == "Low" {
		return
	}

	// 1. 获取候选任务
	allTasks, err := s.Repo.FindAll()
	if err != nil {
		return
	}

	var candidates []model.Task
	var candidateTitles []string
	// 筛选逻辑：状态为 Pending，ID 不同，且优先级低于新任务
	for _, t := range allTasks {
		if t.Status == "Pending" && t.ID != newTask.ID && isLowerPriority(t.Priority, newTask.Priority) {
			candidates = append(candidates, t)
			candidateTitles = append(candidateTitles, fmt.Sprintf(`{"id": %d, "title": "%s"}`, t.ID, t.Title))
		}
	}

	if len(candidates) == 0 {
		return
	}

	// 限制长度
	if len(candidateTitles) > 15 {
		candidateTitles = candidateTitles[:15]
	}

	// 2. 强逻辑 Prompt
	prompt := fmt.Sprintf(`
		新任务: "%s" (优先级: %s)
		
		现有低优先级任务列表:
		[%s]
		
		请分析：现有列表中，是否有任务是完成"新任务"的【前置条件】或【隐含依赖】？
		例如：如果不先"购买服务器"，就无法"上线官网"。那么"购买服务器"就是前置依赖。
		
		如果有，请返回这些任务的 ID 列表。如果没有，返回空列表。
		必须仅返回 JSON 数组格式，例如：[2, 5]
	`, newTask.Title, newTask.Priority, strings.Join(candidateTitles, ","))

	responseStr, err := s.askDeepSeek(prompt)
	if err != nil {
		fmt.Println("AI Sync Error:", err)
		return
	}

	// 3. 提取数字并更新
	re := regexp.MustCompile(`\d+`)
	matches := re.FindAllString(responseStr, -1)

	for _, match := range matches {
		idVal, _ := strconv.Atoi(match)
		targetID := uint(idVal)

		for _, t := range candidates {
			if t.ID == targetID {
				fmt.Printf(">>> 自动升级任务 ID %d (%s) 优先级到 %s\n", t.ID, t.Title, newTask.Priority)
				t.Priority = newTask.Priority
				t.Description = fmt.Sprintf("⚠️ 因任务 [%s] 需要，AI已自动提升此任务优先级。", newTask.Title)
				s.Repo.Update(&t)
			}
		}
	}
}

func isLowerPriority(p1, p2 string) bool {
	weight := map[string]int{"High": 3, "Normal": 2, "Low": 1, "": 1}
	return weight[p1] < weight[p2]
}

// GenerateReportWithAI 生成周报
func (s *AIService) GenerateReportWithAI(tasks []model.Task) string {
	if len(tasks) == 0 {
		return "本周暂无已完成的任务。"
	}
	var taskListStr strings.Builder
	taskListStr.WriteString("已完成任务列表：\n")
	for i, t := range tasks {
		taskListStr.WriteString(fmt.Sprintf("%d. %s (耗时: %.1f小时)\n", i+1, t.Title, t.EstimatedHours))
	}
	// 提示 AI 使用 Markdown 格式
	prompt := fmt.Sprintf(`
		请根据以下任务列表生成周报。
		要求：
		1. 使用 Markdown 格式 (使用 **加粗**, - 列表)。
		2. 总结工作亮点。
		3. 计算总耗时。
		
		%s
	`, taskListStr.String())

	report, err := s.askDeepSeek(prompt)
	if err != nil {
		return "AI 生成报告失败"
	}
	return report
}

// ----------------- 内部通用方法 -----------------

func (s *AIService) extractTaskInfoWithAI(input string) TaskAIInfo {
	nowStr := time.Now().Format("2006-01-02 15:04:05")
	prompt := fmt.Sprintf(`当前时间:%s。任务:"%s"。提取截止时间(due_date, YYYY-MM-DD HH:mm:ss)和耗时(estimated_hours)。返回JSON。`, nowStr, input)
	responseStr, _ := s.askDeepSeek(prompt)
	responseStr = cleanJSONString(responseStr)
	var info TaskAIInfo
	json.Unmarshal([]byte(responseStr), &info)
	return info
}

func (s *AIService) calculatePriority(newTask *model.Task) string {
	score := 0.0
	// 1. 时间分
	if !newTask.DueDate.IsZero() {
		hoursLeft := time.Until(newTask.DueDate).Hours()
		if hoursLeft < 0 {
			score += 100
		} else if hoursLeft <= 24 {
			score += 50
		} else if hoursLeft <= 72 {
			score += 30
		}
	}
	// 2. 关键词分
	lowerTitle := strings.ToLower(newTask.Title)
	if strings.Contains(lowerTitle, "紧急") || strings.Contains(lowerTitle, "必须") || strings.Contains(lowerTitle, "上线") {
		score += 30
	}

	// 3. 依赖分
	if s.Repo != nil {
		if pendingTasks, err := s.Repo.FindAll(); err == nil && len(pendingTasks) > 0 {
			var existingTitles []string
			for _, t := range pendingTasks {
				if t.Status == "Pending" && t.ID != newTask.ID {
					existingTitles = append(existingTitles, t.Title)
				}
			}
			if len(existingTitles) > 0 {
				score += s.analyzeDependencyWithAI(newTask.Title, existingTitles)
			}
		}
	}
	if score >= 40 {
		return "High"
	} else if score >= 20 {
		return "Normal"
	}
	return "Low"
}

func (s *AIService) analyzeDependencyWithAI(newTitle string, existingTitles []string) float64 {
	if len(existingTitles) > 20 {
		existingTitles = existingTitles[:20]
	}
	prompt := fmt.Sprintf(`任务列表: %v。新任务: "%s"。新任务是否是列表中任务的前置依赖？是则返回Yes。`, existingTitles, newTitle)
	result, _ := s.askDeepSeek(prompt)
	if strings.Contains(strings.ToLower(result), "yes") {
		return 40.0
	}
	return 0.0
}

func (s *AIService) askDeepSeek(prompt string) (string, error) {
	// 如果 Key 为空，直接返回错误，避免无效请求
	if s.Config.APIKey == "" {
		return "", fmt.Errorf("API Key not set")
	}

	requestBody, _ := json.Marshal(map[string]interface{}{
		"model":    s.Config.Model,
		"messages": []map[string]string{{"role": "system", "content": "你是一个项目管理助手。请直接按要求输出结果，不要废话。"}, {"role": "user", "content": prompt}},
	})
	req, _ := http.NewRequest("POST", s.Config.Endpoint, bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.Config.APIKey)
	client := &http.Client{Timeout: 45 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	bodyBytes, _ := io.ReadAll(resp.Body)
	var result AIResponse
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return "", fmt.Errorf("err")
	}
	if len(result.Choices) == 0 {
		return "", fmt.Errorf("empty")
	}
	return result.Choices[0].Message.Content, nil
}

func cleanJSONString(str string) string {
	str = strings.TrimSpace(str)
	str = strings.TrimPrefix(str, "```json")
	str = strings.TrimPrefix(str, "```")
	str = strings.TrimSuffix(str, "```")
	return strings.TrimSpace(str)
}
