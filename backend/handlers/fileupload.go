package handlers

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const maxUploadBytes = 10 << 20 // 10MB

var allowedUploadExt = map[string]struct{}{
	".pdf": {}, ".doc": {}, ".docx": {}, ".xls": {}, ".xlsx": {},
	".png": {}, ".jpg": {}, ".jpeg": {},
}

// sanitizeOriginalFilename 仅用于展示/记录，去掉路径成分并限制长度。
func sanitizeOriginalFilename(name string) string {
	name = strings.TrimSpace(name)
	name = strings.ReplaceAll(name, "\\", "/")
	if i := strings.LastIndex(name, "/"); i >= 0 {
		name = name[i+1:]
	}
	if len(name) > 200 {
		name = name[:200]
	}
	return name
}

func extAllowed(ext string) bool {
	ext = strings.ToLower(ext)
	_, ok := allowedUploadExt[ext]
	return ok
}

// validateUploadStream 根据扩展名与文件头 Content-Type（魔数）校验，防伪造后缀。
func validateUploadStream(ext string, head []byte) error {
	ext = strings.ToLower(ext)
	if !extAllowed(ext) {
		return fmt.Errorf("不支持的文件类型，仅允许 pdf/doc/docx/xls/xlsx/png/jpg/jpeg")
	}
	if len(head) == 0 {
		return fmt.Errorf("无法读取文件内容")
	}
	detected := http.DetectContentType(head)
	detected = strings.ToLower(strings.TrimSpace(strings.Split(detected, ";")[0]))

	switch ext {
	case ".pdf":
		if detected != "application/pdf" && !bytes.HasPrefix(head, []byte("%PDF")) {
			return fmt.Errorf("文件内容与扩展名不符（期望 PDF）")
		}
	case ".png":
		if detected != "image/png" {
			return fmt.Errorf("文件内容与扩展名不符（期望 PNG）")
		}
	case ".jpg", ".jpeg":
		if !strings.HasPrefix(detected, "image/jpeg") && !bytes.HasPrefix(head, []byte{0xff, 0xd8, 0xff}) {
			return fmt.Errorf("文件内容与扩展名不符（期望 JPEG）")
		}
	case ".docx", ".xlsx":
		// OOXML 为 zip；部分环境 DetectContentType 为 application/zip 或 application/octet-stream
		if detected != "application/zip" && detected != "application/octet-stream" {
			return fmt.Errorf("文件内容与扩展名不符（期望 Office Open XML / zip）")
		}
		if len(head) >= 4 && !bytes.Equal(head[0:2], []byte("PK")) {
			return fmt.Errorf("文件内容与扩展名不符（期望 zip 容器）")
		}
	case ".doc", ".xls":
		// 老版 Office 为 OLE 复合文档；部分系统返回 octet-stream
		if len(head) >= 8 && bytes.HasPrefix(head, []byte{0xd0, 0xcf, 0x11, 0xe0, 0xa1, 0xb1, 0x1a, 0xe1}) {
			return nil
		}
		if ext == ".doc" && (strings.Contains(detected, "msword") || strings.Contains(detected, "word")) {
			return nil
		}
		if ext == ".xls" && (strings.Contains(detected, "excel") || strings.Contains(detected, "spreadsheet")) {
			return nil
		}
		return fmt.Errorf("文件内容与扩展名不符（期望 Word/Excel 二进制文档）")
	default:
		return fmt.Errorf("不支持的文件类型")
	}
	return nil
}

// saveMultipartFileValidated 读首 512 字节做校验后写入 dstPath。
func saveMultipartFileValidated(fh *multipart.FileHeader, dstPath string) (origName string, err error) {
	if fh.Size > maxUploadBytes {
		return "", fmt.Errorf("单文件不能超过 10MB")
	}
	ext := strings.ToLower(filepath.Ext(fh.Filename))
	if !extAllowed(ext) {
		return "", fmt.Errorf("不支持的文件类型，仅允许 pdf/doc/docx/xls/xlsx/png/jpg/jpeg")
	}

	src, err := fh.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	head := make([]byte, 512)
	n, err := io.ReadFull(src, head)
	if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
		return "", err
	}
	if err := validateUploadStream(ext, head[:n]); err != nil {
		return "", err
	}

	out, err := os.Create(dstPath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	if _, err := io.Copy(out, io.MultiReader(bytes.NewReader(head[:n]), src)); err != nil {
		return "", err
	}
	return sanitizeOriginalFilename(fh.Filename), nil
}
