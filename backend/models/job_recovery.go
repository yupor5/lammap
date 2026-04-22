package models

import (
	"log"

	"gorm.io/gorm"
)

// RecoverStuckAsyncJobs 将上次进程异常退出时卡在 running 的任务标记为失败，避免前端永久轮询。
func RecoverStuckAsyncJobs(db *gorm.DB) {
	msg := "service restarted"
	running := string(GenerateJobRunning)
	updates := map[string]interface{}{
		"status":    string(GenerateJobFailed),
		"error_msg": msg,
	}

	n1 := db.Model(&GenerateJob{}).Where("status = ?", running).Updates(updates).RowsAffected
	n2 := db.Model(&AttachmentGenerateJob{}).Where("status = ?", running).Updates(updates).RowsAffected
	n3 := db.Model(&AttachmentZipJob{}).Where("status = ?", running).Updates(updates).RowsAffected
	n4 := db.Model(&AttachmentPackJob{}).Where("status = ?", running).Updates(updates).RowsAffected
	n5 := db.Model(&InquiryExampleJob{}).Where("status = ?", running).Updates(updates).RowsAffected
	n6 := db.Model(&ProductExampleJob{}).Where("status = ?", running).Updates(updates).RowsAffected
	if n1+n2+n3+n4+n5+n6 > 0 {
		log.Printf("[jobs] marked stuck running as failed: generate=%d attachmentGen=%d zip=%d pack=%d inquiryEx=%d productEx=%d", n1, n2, n3, n4, n5, n6)
	}
}
