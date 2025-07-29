package service

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"smart-analysis/internal/model"
	"smart-analysis/internal/utils"
	"time"
)

type FileService struct {
	files    map[int]*model.File
	nextID   int
	basePath string
}

func NewFileService() *FileService {
	return &FileService{
		files:    make(map[int]*model.File),
		nextID:   1,
		basePath: "./uploads",
	}
}

// Upload 上传文件
func (s *FileService) Upload(userID int, fileHeader *multipart.FileHeader) (*model.File, error) {
	// 检查文件大小 500MB
	if fileHeader.Size > 500*1024*1024 { // 500MB
		return nil, errors.New("file size exceeds limit")
	}

	// 检查文件类型
	ext := filepath.Ext(fileHeader.Filename)
	if ext != ".csv" && ext != ".xlsx" && ext != ".xls" && ext != ".json" {
		return nil, errors.New("unsupported file type")
	}

	// 生成文件名
	filename := utils.GenerateFileName(fileHeader.Filename)
	filePath := filepath.Join(s.basePath, filename)

	// 确保上传目录存在
	if err := os.MkdirAll(s.basePath, 0755); err != nil {
		return nil, err
	}

	// 保存文件
	src, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	dst, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		return nil, err
	}

	// 创建文件记录
	file := &model.File{
		ID:        s.nextID,
		UserID:    userID,
		Name:      filename,
		OrigName:  fileHeader.Filename,
		Path:      filePath,
		Size:      fileHeader.Size,
		Type:      ext,
		Status:    "uploaded",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	s.files[s.nextID] = file
	s.nextID++

	// 异步处理文件（解析数据结构等）
	go s.processFile(file)

	return file, nil
}

// processFile 处理文件（解析数据结构）
func (s *FileService) processFile(file *model.File) {
	file.Status = "processing"

	// 根据文件类型解析
	switch utils.GetFileType(file.Name) {
	case utils.CSV:
		_, err := utils.ParseCSV(file.Path)
		if err != nil {
			file.Status = "error"
			return
		}
	case utils.Excel:
		_, err := utils.ParseExcel(file.Path)
		if err != nil {
			file.Status = "error"
			return
		}
	case utils.JSON:
		_, err := utils.ParseJSON(file.Path)
		if err != nil {
			file.Status = "error"
			return
		}
	}

	file.Status = "ready"
	file.UpdatedAt = time.Now()
}

// GetFilesByUserID 获取用户的文件列表
func (s *FileService) GetFilesByUserID(userID int) []*model.File {
	var userFiles []*model.File
	for _, file := range s.files {
		if file.UserID == userID {
			userFiles = append(userFiles, file)
		}
	}
	return userFiles
}

// GetFileByID 根据ID获取文件
func (s *FileService) GetFileByID(fileID int) (*model.File, error) {
	file, exists := s.files[fileID]
	if !exists {
		return nil, errors.New("file not found")
	}
	return file, nil
}

// DeleteFile 删除文件
func (s *FileService) DeleteFile(userID, fileID int) error {
	file, exists := s.files[fileID]
	if !exists {
		return errors.New("file not found")
	}

	if file.UserID != userID {
		return errors.New("permission denied")
	}

	// 删除物理文件
	if err := os.Remove(file.Path); err != nil {
		return err
	}

	// 删除记录
	delete(s.files, fileID)
	return nil
}

// PreviewFile 预览文件数据
func (s *FileService) PreviewFile(userID, fileID int, limit int) (interface{}, error) {
	file, exists := s.files[fileID]
	if !exists {
		return nil, errors.New("file not found")
	}

	if file.UserID != userID {
		return nil, errors.New("permission denied")
	}

	if file.Status != "ready" {
		return nil, fmt.Errorf("file is not ready, current status: %s", file.Status)
	}

	// 根据文件类型返回预览数据
	switch utils.GetFileType(file.Name) {
	case utils.CSV:
		data, err := utils.ParseCSV(file.Path)
		if err != nil {
			return nil, err
		}
		// 限制返回行数
		if limit > 0 && len(data.Rows) > limit {
			data.Rows = data.Rows[:limit]
		}
		return data, nil
	case utils.Excel:
		data, err := utils.ParseExcel(file.Path)
		if err != nil {
			return nil, err
		}
		// 限制返回行数
		if limit > 0 && len(data.Rows) > limit {
			data.Rows = data.Rows[:limit]
		}
		return data, nil
	case utils.JSON:
		return utils.ParseJSON(file.Path)
	}

	return nil, errors.New("unsupported file type")
}
