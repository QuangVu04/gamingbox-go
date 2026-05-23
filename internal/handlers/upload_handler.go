package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
)

type UploadHandler struct{}

func NewUploadHandler() *UploadHandler {
	return &UploadHandler{}
}

func (h *UploadHandler) UploadImage(c *gin.Context) {
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Không tìm thấy file ảnh hợp lệ"})
		return
	}

	// Lấy phần mở rộng của file
	ext := filepath.Ext(file.Filename)
	if ext == "" {
		ext = ".png" // Fallback
	}

	// Tạo tên file ngẫu nhiên dựa trên timestamp
	newFilename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	
	// Thư mục lưu file
	uploadDir := filepath.Join("uploads", "images")
	
	// Tạo thư mục nếu chưa có
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi khi tạo thư mục: " + err.Error()})
		return
	}

	// Đường dẫn lưu file
	savePath := filepath.Join(uploadDir, newFilename)

	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể lưu file: " + err.Error()})
		return
	}

	// Trả về URL của ảnh
	host := c.Request.Host
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	url := fmt.Sprintf("%s://%s/uploads/images/%s", scheme, host, newFilename)

	c.JSON(http.StatusOK, gin.H{
		"message": "Upload thành công",
		"url":     url,
	})
}
