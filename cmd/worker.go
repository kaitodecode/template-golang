package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	_ "template-golang/docs"
	"template-golang/pkg/config"
	"template-golang/pkg/fileUploader"
	"template-golang/pkg/helper"
	"template-golang/pkg/logger"
	utlog "template-golang/pkg/logger"
	"template-golang/pkg/redisx"

	"github.com/goccy/go-json"
	"github.com/spf13/cobra"
)

var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "Jalankan worker",
	Run: func(cmd *cobra.Command, args []string) {

		cfg := config.LoadConfig()
		utlog.Init(cfg.Env)

		ctx := context.Background()

		client, err := redisx.New()
		if err != nil {
			panic(fmt.Errorf("failed to initialize redis client: %v", err))
		}

		if err := client.InitConsumerGroup(ctx, "upload_jobs", "worker"); err != nil {
			panic(fmt.Errorf("failed to initialize consumer group: %v", err))
		}

		logger.L().Infoln("🚀 Worker started. Listening jobs")

		for {
			msgs, err := client.ConsumeJob(ctx, "upload_jobs", "worker", "worker-1", 10, 5*time.Second)
			if err != nil {
				logger.L().Errorf("failed to consume job: %v", err)
				continue
			}

			if len(msgs) == 0 {
				logger.L().Infoln("🗑️ no jobs found")
				continue
			}

			for _, msg := range msgs {
				jobID := msg.ID

				// Ambil payload asli dari Redis
				rawPayload, ok := msg.Payload.(map[string]any)["payload"]
				if !ok {
					logger.L().Errorf("job %s: payload field missing", jobID)
					continue
				}

				jsonStr, ok := rawPayload.(string)
				if !ok {
					logger.L().Errorf("job %s: payload is not a string", jobID)
					continue
				}

				var payload fileUploader.QueueUploadFile
				if err := json.Unmarshal([]byte(jsonStr), &payload); err != nil {
					logger.L().Errorf("job %s: failed to unmarshal payload into QueueUploadFile: %v", jobID, err)
					continue
				}

				// Validasi field penting
				if payload.FilePathTmp == nil {
					logger.L().Errorf("job %s: FilePathTmp field missing", jobID)
					continue
				}
				if payload.IsCompressToWebp == nil {
					logger.L().Errorf("job %s: IsCompressToWebp field missing", jobID)
					continue
				}

				// Pastikan file tmp ada
				if _, err := os.Stat(*payload.FilePathTmp); err != nil {
					logger.L().Errorf("job %s: tmp file not found (%s): %v", jobID, *payload.FilePathTmp, err)
					continue
				}

				logger.L().Infof("job %s: start processing filePath=%s tmp=%s compress=%v",
					jobID, payload.FilePath, *payload.FilePathTmp, *payload.IsCompressToWebp)

				// Tentukan folder berdasarkan FilePath
				folder := fileUploader.ExtractFolderFromFilePath(payload.FilePath)

				// Retry logic up to 5 times
				var lastErr error
				success := false
				for attempt := 1; attempt <= 5; attempt++ {
					err := fileUploader.UploadFileFromPath(ctx, *payload.FilePathTmp, fileUploader.FileUploadOptions{
						Folder:           folder,
						NameFile:         payload.FilePath,
						IsCompressToWebp: helper.BoolPtr(*payload.IsCompressToWebp),
					})
					if err == nil {
						success = true
						break
					}

					// Jika ada old file, hapus sebelum retry
					if payload.OldFile != nil {
						if err = fileUploader.DeleteFile(ctx, *payload.OldFile); err != nil {
							logger.L().Errorf("job %s: failed to delete old file %s: %v", jobID, *payload.OldFile, err)
						} else {
							logger.L().Infof("job %s: old file %s deleted", jobID, *payload.OldFile)
						}
					}

					lastErr = err
					logger.L().Warnf("job %s: upload attempt %d failed: %v", jobID, attempt, err)
					if attempt < 5 {
						time.Sleep(time.Second * time.Duration(attempt))
					}
				}

				if !success {
					logger.L().Errorf("job %s: failed to upload file after 5 attempts: %v", jobID, lastErr)
					continue
				}

				// ACK setelah selesai
				if err := client.AckJob(ctx, "upload_jobs", "worker", msg.ID); err != nil {
					logger.L().Errorf("job %s: ack error: %v", jobID, err)
				} else {
					logger.L().Infoln("✅ Job done:", jobID)

					// Bersihkan tmp file
					if err := os.Remove(*payload.FilePathTmp); err != nil {
						logger.L().Warnf("job %s: failed to remove tmp file %s: %v", jobID, *payload.FilePathTmp, err)
					} else {
						logger.L().Infof("job %s: tmp file %s removed", jobID, *payload.FilePathTmp)
					}
				}
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(workerCmd)
}
