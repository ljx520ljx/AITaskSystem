package repository

import (
	"AITaskSystem/model"
	"gorm.io/gorm"
)

// TaskRepository 定义了任务操作的接口
// 即使以后换成 MySQL，只要实现了这个接口，业务层代码都不用动
type TaskRepository struct {
	DB *gorm.DB
}

// NewTaskRepository 构造函数
func NewTaskRepository(db *gorm.DB) *TaskRepository {
	return &TaskRepository{DB: db}
}

// Create 创建任务
func (r *TaskRepository) Create(task *model.Task) error {
	return r.DB.Create(task).Error
}

// FindAll 获取所有任务
func (r *TaskRepository) FindAll() ([]model.Task, error) {
	var tasks []model.Task
	// 这里的 Order 按照创建时间倒序排列
	err := r.DB.Order("created_at desc").Find(&tasks).Error
	return tasks, err
}

// FindPending 获取所有待办任务 (用于周报)
func (r *TaskRepository) FindCompleted() ([]model.Task, error) {
	var tasks []model.Task
	err := r.DB.Where("status = ?", "Completed").Find(&tasks).Error
	return tasks, err
}

// Update 更新任务
func (r *TaskRepository) Update(task *model.Task) error {
	return r.DB.Save(task).Error
}

// FindByID 根据ID查找
func (r *TaskRepository) FindByID(id uint) (*model.Task, error) {
	var task model.Task
	err := r.DB.First(&task, id).Error
	return &task, err
}

// Delete 删除任务
func (r *TaskRepository) Delete(id uint) error {
	return r.DB.Delete(&model.Task{}, id).Error
}
