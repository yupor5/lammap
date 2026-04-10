package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const maxUploadBytes = 10 << 20 // 10MB

var allowedUploadExt = map[string]struct{}{
	".pdf": {}, ".doc": {}, ".docx": {}, ".xls": {}, ".xlsx": {},
	".png": {}, ".jpg": {}, ".jpeg": {},
}

func Upload() gin.HandlerFunc {
	return func(c *gin.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			Error(c, http.StatusBadRequest, "请选择文件")
			return
		}

		if file.Size > maxUploadBytes {
			Error(c, http.StatusBadRequest, "单文件不能超过 10MB")
			return
		}

		ext := strings.ToLower(filepath.Ext(file.Filename))
		if _, ok := allowedUploadExt[ext]; !ok {
			Error(c, http.StatusBadRequest, "不支持的文件类型，仅允许 pdf/doc/docx/xls/xlsx/png/jpg/jpeg")
			return
		}

		uploadDir := filepath.Join("uploads", time.Now().Format("2006-01"))
		if err := os.MkdirAll(uploadDir, 0755); err != nil {
			Error(c, http.StatusInternalServerError, "创建目录失败")
			return
		}

		filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
		dst := filepath.Join(uploadDir, filename)

		if err := c.SaveUploadedFile(file, dst); err != nil {
			Error(c, http.StatusInternalServerError, "保存文件失败")
			return
		}

		webPath := "/" + filepath.ToSlash(dst)
		Success(c, gin.H{
			"filename": file.Filename,
			"path":     webPath,
			"url":      webPath,
			"size":     file.Size,
		})
	}
}
