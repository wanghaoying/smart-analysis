package model

import (
	"time"
)

// User 用户模型
type User struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	Username  string    `json:"username" gorm:"uniqueIndex"`
	Email     string    `json:"email" gorm:"uniqueIndex"`
	Password  string    `json:"-" gorm:"not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// File 文件模型
type File struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	UserID    int       `json:"user_id"`
	Name      string    `json:"name"`
	OrigName  string    `json:"orig_name"`
	Path      string    `json:"path"`
	Size      int64     `json:"size"`
	Type      string    `json:"type"`
	Status    string    `json:"status"` // uploaded, processing, ready, error
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	User      User      `json:"user" gorm:"foreignKey:UserID"`
}

// Session 会话模型
type Session struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	UserID    int       `json:"user_id"`
	Name      string    `json:"name"`
	FileID    *int      `json:"file_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	User      User      `json:"user" gorm:"foreignKey:UserID"`
	File      *File     `json:"file" gorm:"foreignKey:FileID"`
}

// Query 查询记录模型
type Query struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	SessionID int       `json:"session_id"`
	UserID    int       `json:"user_id"`
	Question  string    `json:"question"`
	Answer    string    `json:"answer"`
	QueryType string    `json:"query_type"` // analysis, visualization, report
	Status    string    `json:"status"`     // processing, completed, error
	CreatedAt time.Time `json:"created_at"`
	Session   Session   `json:"session" gorm:"foreignKey:SessionID"`
	User      User      `json:"user" gorm:"foreignKey:UserID"`
}

// LLMConfig LLM配置模型
type LLMConfig struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	UserID    int       `json:"user_id"`
	Provider  string    `json:"provider"` // openai, hunyuan, tongyi
	APIKey    string    `json:"api_key"`
	Model     string    `json:"model"`
	IsDefault bool      `json:"is_default"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	User      User      `json:"user" gorm:"foreignKey:UserID"`
}

// Usage 使用量模型
type Usage struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	UserID    int       `json:"user_id"`
	Provider  string    `json:"provider"`
	Model     string    `json:"model"`
	Tokens    int       `json:"tokens"`
	Cost      float64   `json:"cost"`
	QueryID   int       `json:"query_id"`
	CreatedAt time.Time `json:"created_at"`
	User      User      `json:"user" gorm:"foreignKey:UserID"`
	Query     Query     `json:"query" gorm:"foreignKey:QueryID"`
}
