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

func Upload() gin.HandlerFunc {
	return func(c *gin.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			Error(c, http.StatusBadRequest, "请选择文件")
			return
		}

		ext := strings.ToLower(filepath.Ext(file.Filename))
		uploadDir := filepath.Join("uploads", time.Now().Format("2006-01"))
		if err := os.MkdirAll(uploadDir, 0755); err != nil {
			Error(c, http.StatusInternalServerError, "创建目录失败")
			return
		}

		filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
		dst := filepath.Join(uploadDir, filename)

		orig, err := saveMultipartFileValidated(file, dst)
		if err != nil {
			_ = os.Remove(dst)
			Error(c, http.StatusBadRequest, err.Error())
			return
		}

		st, _ := os.Stat(dst)
		sz := file.Size
		if st != nil {
			sz = st.Size()
		}

		webPath := "/" + filepath.ToSlash(dst)
		Success(c, gin.H{
			"filename": orig,
			"path":     webPath,
			"url":      webPath,
			"size":     sz,
		})
	}
}
