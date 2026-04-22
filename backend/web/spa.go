package web

import (
	"io/fs"
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// RegisterSPAFallback 在注册完 /api、/uploads 等路由后调用。未匹配到的 GET/HEAD
// 从嵌入的 dist 提供静态文件；无扩展名路径回退到 index.html 以支持 Vue Router history 模式。
func RegisterSPAFallback(r *gin.Engine) {
	sub, err := fs.Sub(Dist, "dist")
	if err != nil {
		panic("web embed: " + err.Error())
	}

	r.NoRoute(func(c *gin.Context) {
		if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead {
			c.String(http.StatusNotFound, "not found")
			return
		}

		rel := strings.Trim(c.Request.URL.Path, "/")
		if rel != "" {
			rel = path.Clean("/" + rel)
			rel = strings.TrimPrefix(rel, "/")
		}
		if strings.Contains(rel, "..") {
			c.String(http.StatusBadRequest, "bad path")
			return
		}
		if rel == "" || rel == "." {
			rel = "index.html"
		}

		b, err := fs.ReadFile(sub, rel)
		if err != nil {
			if e := path.Ext(rel); e != "" {
				c.String(http.StatusNotFound, "not found")
				return
			}
			b, err = fs.ReadFile(sub, "index.html")
		}
		if err != nil {
			c.String(http.StatusNotFound, "not found")
			return
		}

		ct := contentTypeFor(rel, b)
		if c.Request.Method == http.MethodHead {
			c.Header("Content-Type", ct)
			c.Header("Content-Length", strconv.Itoa(len(b)))
			c.Status(http.StatusOK)
			return
		}
		c.Data(http.StatusOK, ct, b)
	})
}

func contentTypeFor(rel string, b []byte) string {
	lower := strings.ToLower(rel)
	switch {
	case strings.HasSuffix(lower, ".html"):
		return "text/html; charset=utf-8"
	case strings.HasSuffix(lower, ".js"):
		return "text/javascript; charset=utf-8"
	case strings.HasSuffix(lower, ".css"):
		return "text/css; charset=utf-8"
	case strings.HasSuffix(lower, ".svg"):
		return "image/svg+xml"
	case strings.HasSuffix(lower, ".json"):
		return "application/json; charset=utf-8"
	case strings.HasSuffix(lower, ".woff2"):
		return "font/woff2"
	case strings.HasSuffix(lower, ".woff"):
		return "font/woff"
	case strings.HasSuffix(lower, ".ttf"):
		return "font/ttf"
	case strings.HasSuffix(lower, ".ico"):
		return "image/x-icon"
	}
	return http.DetectContentType(b)
}
