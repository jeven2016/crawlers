package stream

import (
	"crawlers/pkg/base"
	"crawlers/pkg/model/entity"
	"testing"
)

// exists, true
func TestUpdateTaskStatusOne(t *testing.T) {
	task := &entity.CatalogPageTask{}
	updateTaskStatus(task, true, true)

	if task.Status != base.TaskStatusFinished {
		t.Error("status should be finished")
	}

	if task.CreatedDate != nil {
		t.Error("CreatedDate shouldn't be set")
	}

	if task.LastUpdated == nil {
		t.Error("LastUpdated should be set")
	}
}

// exists, false
func TestUpdateTaskStatusTwo(t *testing.T) {
	task := &entity.CatalogPageTask{}
	updateTaskStatus(task, true, false)

	if task.Status != base.TaskStatusRetryFailed {
		t.Error("status should be TaskStatusRetryFailed")
	}

	if task.LastUpdated == nil {
		t.Error("LastUpdated should be set")
	}
}

// exists, retries=1, false
func TestUpdateTaskStatusRetry(t *testing.T) {
	task := &entity.CatalogPageTask{Status: base.TaskStatusFailed, Retries: 1}
	updateTaskStatus(task, true, false)

	if task.Status != base.TaskStatusRetryFailed {
		t.Error("status should be TaskStatusRetryFailed")
	}

	if task.LastUpdated == nil {
		t.Error("LastUpdated should be set")
	}

	if task.Retries != 2 {
		t.Error("Retries should be 2")
	}
}

// not exists, true
func TestUpdateTaskStatusThree(t *testing.T) {
	task := &entity.CatalogPageTask{}
	updateTaskStatus(task, false, true)

	if task.Status != base.TaskStatusFinished {
		t.Error("status should be TaskStatusFinished")
	}

	if task.LastUpdated != nil {
		t.Error("LastUpdated shouldn't be set")
	}
}

// not exists, false
func TestUpdateTaskStatusFour(t *testing.T) {
	task := &entity.CatalogPageTask{}
	updateTaskStatus(task, false, false)

	if task.Status != base.TaskStatusFailed {
		t.Error("status should be TaskStatusFailed")
	}

	if task.CreatedDate == nil {
		t.Error("CreatedDate should be set")
	}

	if task.LastUpdated != nil {
		t.Error("LastUpdated shouldn't be set")
	}
}
