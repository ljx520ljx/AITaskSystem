package model

import (
	"time"

	"gorm.io/gorm"
)

// Task 表示一个工作任务
// 使用 GORM 的 Model 结构体包含默认的 ID, CreatedAt, UpdatedAt, DeletedAt 字段
type Task struct {
	gorm.Model
	Title          string    `json:"title"`           // 任务标题
	Description    string    `json:"description"`     // 任务详情
	Status         string    `json:"status"`          // 状态: Pending(待办), Completed(已完成)
	DueDate        time.Time `json:"due_date"`        // 截止时间
	Priority       string    `json:"priority"`        // 优先级: High(高), Normal(普通), Low(低)
	EstimatedHours float64   `json:"estimated_hours"` // 预计耗时(小时)
}

// TableName 指定数据库表名，虽然 GORM 会自动复数化，但显式指定更规范
func (Task) TableName() string {
	return "tasks"
}
