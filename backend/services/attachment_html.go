package services

import (
	"encoding/json"
	"fmt"
	"strings"

	"quotepro-backend/config"
)

// GenerateAttachmentHTML 让 AI 基于报价上下文生成“附件内容”的 HTML（可直接落盘并通过 /uploads 静态访问预览/下载）。
// 注意：这里不做 PDF 转换，先确保“有真实文件可打开”，后续再升级为后端转 PDF。
func GenerateAttachmentHTML(cfg *config.Config, attachmentName string, params map[string]interface{}, normalizedQuote map[string]interface{}) (string, error) {
	ctx := map[string]interface{}{
		"attachmentName": attachmentName,
		"params":         params,
		"quote":          normalizedQuote,
	}
	ctxJSON, _ := json.Marshal(ctx)

	sys := strings.TrimSpace(`
你是一个外贸报价助手与文档排版助手。请基于输入 JSON，生成一份“附件资料”的 HTML 文档内容。

严格规则：
1) 只输出完整 HTML（必须包含 <html> <head> <meta charset="utf-8"> <body>），不要 markdown 代码块，不要解释文字。
2) 使用内联 CSS（<style>）保证打印/导出友好：A4 宽度、字号适中、表格可读。
3) 内容必须围绕 attachmentName 生成：例如包含 "Spec Sheet/参数" 则输出规格参数表；包含 "Color Box/彩盒" 则输出彩盒/包装参数；否则输出通用的报价资料说明与参数汇总。
4) 若某字段缺失，显示 "-"，禁止编造客户未提供的硬性承诺（如交期/价格/认证等），可用“待确认/仅供参考”表达。
5) 文档语言：若 quote.replyVersions[0].language 为 zh 则以中文为主；否则以英文为主（可夹带必要的中英文标题）。
`)

	user := fmt.Sprintf("输入 JSON：\n%s", string(ctxJSON))
	return ChatWithAI(cfg, []ChatMessage{
		{Role: "system", Content: sys},
		{Role: "user", Content: user},
	})
}

