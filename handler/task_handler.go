package handler

import (
	"AITaskSystem/repository"
	"AITaskSystem/service"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type TaskHandler struct {
	Repo      *repository.TaskRepository
	AIService *service.AIService
}

func NewTaskHandler(repo *repository.TaskRepository) *TaskHandler {
	aiService := service.NewAIService(repo)
	return &TaskHandler{
		Repo:      repo,
		AIService: aiService,
	}
}

// CreateTask [统一入口]
func (h *TaskHandler) CreateTask(c *gin.Context) {
	var req struct {
		Title  string `json:"title"`
		Prompt string `json:"prompt"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input := req.Prompt
	if input == "" {
		input = req.Title
	}
	if input == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "内容不能为空"})
		return
	}

	// 1. 分析新任务
	task := h.AIService.ParseTaskFromInput(input)

	// 2. 保存新任务
	if err := h.Repo.Create(&task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task"})
		return
	}

	// 3.触发依赖反向同步
	// AI 会去检查有没有旧任务需要跟着变更为高优先级
	go h.AIService.SyncPrioritiesForDependencies(&task)

	c.JSON(http.StatusCreated, gin.H{
		"message": "Task created",
		"task":    task,
	})
}

// GetAllTasks
func (h *TaskHandler) GetAllTasks(c *gin.Context) {
	tasks, err := h.Repo.FindAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error"})
		return
	}
	c.JSON(http.StatusOK, tasks)
}

// GetTaskByID
func (h *TaskHandler) GetTaskByID(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)
	task, err := h.Repo.FindByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		return
	}
	c.JSON(http.StatusOK, task)
}

// UpdateTask (支持通用更新，包括状态反选)
func (h *TaskHandler) UpdateTask(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)
	existingTask, err := h.Repo.FindByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		return
	}
	// 绑定部分更新
	if err := c.ShouldBindJSON(existingTask); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.Repo.Update(existingTask); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed"})
		return
	}
	c.JSON(http.StatusOK, existingTask)
}

// MarkComplete (保留兼容，但前端建议用 UpdateTask)
func (h *TaskHandler) MarkComplete(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)
	task, err := h.Repo.FindByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		return
	}
	task.Status = "Completed"
	h.Repo.Update(task)
	c.JSON(http.StatusOK, task)
}

// DeleteTask
func (h *TaskHandler) DeleteTask(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)
	if err := h.Repo.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Deleted"})
}

// GetWeeklyReport
func (h *TaskHandler) GetWeeklyReport(c *gin.Context) {
	completed, err := h.Repo.FindCompleted()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error"})
		return
	}
	report := h.AIService.GenerateReportWithAI(completed)
	c.JSON(http.StatusOK, gin.H{"report": report})
}
