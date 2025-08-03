package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

// Task 结构体表示任务
type Task struct {
	ID        int       `json:"id"`
	Text      string    `json:"text"`
	Completed bool      `json:"completed"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// History 结构体表示历史记录
type History struct {
	ID        int       `json:"id"`
	TaskID    int       `json:"task_id"`
	TaskText  string    `json:"task_text"`
	Action    string    `json:"action"` // "add", "complete", "delete"
	CreatedAt time.Time `json:"created_at"`
}

var db *sql.DB

// 初始化数据库
func initDB() {
	var err error
	db, err = sql.Open("sqlite3", "./data/tasks.db")
	if err != nil {
		log.Fatal(err)
	}

	// 创建任务表
	sqlStmt := `
    CREATE TABLE IF NOT EXISTS tasks (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        text TEXT NOT NULL,
        completed BOOLEAN NOT NULL DEFAULT 0,
        created_at DATETIME NOT NULL,
        updated_at DATETIME NOT NULL
    );`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Fatal(err)
	}

	// 创建历史记录表
	sqlStmt = `
    CREATE TABLE IF NOT EXISTS history (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        task_id INTEGER NOT NULL,
        task_text TEXT NOT NULL,
        action TEXT NOT NULL,
        created_at DATETIME NOT NULL
    );`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Fatal(err)
	}
}

// 获取所有任务
func getTasks(c *gin.Context) {
	rows, err := db.Query("SELECT id, text, completed, created_at, updated_at FROM tasks ORDER BY created_at DESC")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var task Task
		err := rows.Scan(&task.ID, &task.Text, &task.Completed, &task.CreatedAt, &task.UpdatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		tasks = append(tasks, task)
	}

	c.JSON(http.StatusOK, tasks)
}

// 添加新任务
func addTask(c *gin.Context) {
	var req struct {
		Text string `json:"text" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	now := time.Now()
	result, err := db.Exec(
		"INSERT INTO tasks(text, completed, created_at, updated_at) VALUES(?, ?, ?, ?)",
		req.Text, false, now, now,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	taskID, _ := result.LastInsertId()
	task := Task{
		ID:        int(taskID),
		Text:      req.Text,
		Completed: false,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// 添加历史记录
	historyID, _ := addHistory(int(taskID), req.Text, "add", now)
	history := History{
		ID:        int(historyID),
		TaskID:    int(taskID),
		TaskText:  req.Text,
		Action:    "add",
		CreatedAt: now,
	}

	c.JSON(http.StatusCreated, gin.H{
		"task":    task,
		"history": history,
	})
}

// 切换任务状态
func toggleTask(c *gin.Context) {
	id := c.Param("id")

	// 获取当前任务状态
	var task Task
	err := db.QueryRow("SELECT id, text, completed, created_at, updated_at FROM tasks WHERE id = ?", id).
		Scan(&task.ID, &task.Text, &task.Completed, &task.CreatedAt, &task.UpdatedAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	// 切换状态
	newStatus := !task.Completed
	now := time.Now()
	_, err = db.Exec(
		"UPDATE tasks SET completed = ?, updated_at = ? WHERE id = ?",
		newStatus, now, id,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 更新任务对象
	task.Completed = newStatus
	task.UpdatedAt = now

	// 添加历史记录
	action := "complete"
	if !newStatus {
		action = "uncomplete"
	}
	historyID, _ := addHistory(task.ID, task.Text, action, now)
	history := History{
		ID:        int(historyID),
		TaskID:    task.ID,
		TaskText:  task.Text,
		Action:    action,
		CreatedAt: now,
	}

	c.JSON(http.StatusOK, gin.H{
		"task":    task,
		"history": history,
	})
}

// 删除任务
func deleteTask(c *gin.Context) {
	id := c.Param("id")

	// 获取任务信息
	var task Task
	err := db.QueryRow("SELECT id, text, completed, created_at, updated_at FROM tasks WHERE id = ?", id).
		Scan(&task.ID, &task.Text, &task.Completed, &task.CreatedAt, &task.UpdatedAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	// 删除任务
	_, err = db.Exec("DELETE FROM tasks WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 添加历史记录
	now := time.Now()
	historyID, _ := addHistory(task.ID, task.Text, "delete", now)
	history := History{
		ID:        int(historyID),
		TaskID:    task.ID,
		TaskText:  task.Text,
		Action:    "delete",
		CreatedAt: now,
	}

	c.JSON(http.StatusOK, gin.H{
		"history": history,
	})
}

// 清除已完成任务
func clearCompleted(c *gin.Context) {
	var req struct {
		IDs []int `json:"ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(req.IDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No task IDs provided"})
		return
	}

	// 获取要删除的任务信息
	var tasks []Task
	query := "SELECT id, text, completed, created_at, updated_at FROM tasks WHERE id IN ("
	args := make([]interface{}, len(req.IDs))
	for i, id := range req.IDs {
		if i > 0 {
			query += ","
		}
		query += "?"
		args[i] = id
	}
	query += ")"

	rows, err := db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var task Task
		err := rows.Scan(&task.ID, &task.Text, &task.Completed, &task.CreatedAt, &task.UpdatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		tasks = append(tasks, task)
	}

	// 删除任务
	query = "DELETE FROM tasks WHERE id IN ("
	for i, id := range req.IDs {
		if i > 0 {
			query += ","
		}
		query += "?"
		args[i] = id
	}
	query += ")"

	_, err = db.Exec(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 添加历史记录
	now := time.Now()
	var historyEntries []History
	for _, task := range tasks {
		historyID, _ := addHistory(task.ID, task.Text, "delete", now)
		historyEntries = append(historyEntries, History{
			ID:        int(historyID),
			TaskID:    task.ID,
			TaskText:  task.Text,
			Action:    "delete",
			CreatedAt: now,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"history": historyEntries,
	})
}

// 获取历史记录
func getHistory(c *gin.Context) {
	rows, err := db.Query("SELECT id, task_id, task_text, action, created_at FROM history ORDER BY created_at DESC LIMIT 50")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var history []History
	for rows.Next() {
		var entry History
		err := rows.Scan(&entry.ID, &entry.TaskID, &entry.TaskText, &entry.Action, &entry.CreatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		history = append(history, entry)
	}

	c.JSON(http.StatusOK, history)
}

// 添加历史记录
func addHistory(taskID int, taskText string, action string, createdAt time.Time) (int64, error) {
	result, err := db.Exec(
		"INSERT INTO history(task_id, task_text, action, created_at) VALUES(?, ?, ?, ?)",
		taskID, taskText, action, createdAt,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func main() {
	// 初始化数据库
	initDB()
	defer db.Close()

	// 创建Gin路由
	r := gin.Default()

	// 允许跨域
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// 静态文件服务 - 推荐使用/static前缀，避免与/api冲突
	r.Static("/static", "./static")

	api := r.Group("/api")
	{
		api.GET("/tasks", getTasks)
		api.POST("/tasks", addTask)
		api.PUT("/tasks/:id/toggle", toggleTask)
		api.DELETE("/tasks/:id", deleteTask)
		api.DELETE("/tasks/completed", clearCompleted)
		api.GET("/history", getHistory)
	}

	// SPA前端路由兜底
	r.NoRoute(func(c *gin.Context) {
		c.File("./static/index.html")
	})

	// 启动服务器
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server running on port %s", port)
	r.Run(":" + port)
}
