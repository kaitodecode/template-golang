package redisx

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"

	"template-golang/pkg/apperror"
	"template-golang/pkg/config"
	"template-golang/pkg/fileUploader"
	"template-golang/pkg/logger"

	"github.com/goccy/go-json"
	"github.com/nrednav/cuid2"
	"github.com/redis/go-redis/v9"
)

// Client wrapper Redis
type Client struct {
	rdb *redis.Client
}

// Job struct generic untuk representasi pesan di Stream
type Job struct {
	ID      string      `json:"id"`
	Payload interface{} `json:"payload"`
}

// New membuat koneksi ke Redis
func New() (*Client, error) {
	conf := config.GetConfig()
	rdb := redis.NewClient(&redis.Options{
		Addr:     conf.RedisAddr,
		Password: conf.RedisPass,
		DB:       0,
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, apperror.New("redisx", "New", 500, err, "failed to connect to Redis")
	}
	return &Client{rdb: rdb}, nil
}

// ================== BASIC CACHE ==================

// Set key-value dengan TTL
func (c *Client) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	var v interface{} = value
	switch value.(type) {
	case string, int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64, bool:
	default:
		jsonBytes, err := json.Marshal(value)
		if err != nil {
			logger.L().Error("Failed to marshal value", "key", key, "err", err)
			return apperror.New("redisx", "Set", 500, err, "failed to marshal value")
		}
		v = string(jsonBytes)
	}
	if err := c.rdb.Set(ctx, key, v, ttl).Err(); err != nil && err.Error() != "redis: nil" {
		return apperror.New("redisx", "Set", 500, err, string(debug.Stack()))
	}
	return nil
}

// Get key
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	res, err := c.rdb.Get(ctx, key).Result()
	if err != nil {
		return "", apperror.New("redisx", "Get", 500, err, string(debug.Stack()))
	}
	return res, nil
}

// Del key
func (c *Client) Del(ctx context.Context, key string) error {
	if err := c.rdb.Del(ctx, key).Err(); err != nil {
		return apperror.New("redisx", "Del", 500, err, "failed to delete key")
	}
	return nil
}

// DelByPattern hapus key by pattern
func (c *Client) DelByPattern(ctx context.Context, pattern string) error {
	keys, err := c.rdb.Keys(ctx, pattern).Result()
	if err != nil {
		return apperror.New("redisx", "DelByPattern", 500, err, "failed to fetch keys with pattern")
	}
	if len(keys) == 0 {
		return nil
	}
	if err := c.rdb.Del(ctx, keys...).Err(); err != nil {
		return apperror.New("redisx", "DelByPattern", 500, err, "failed to delete keys")
	}
	return nil
}

// ================== STREAM / QUEUE ==================

// EnqueueJob menambahkan job ke Redis Stream
func (c *Client) EnqueueJob(ctx context.Context, stream string, payload Job) error {
	jsonPayload, err := json.Marshal(payload.Payload)
	if err != nil {
		return apperror.New("redisx", "EnqueueJob", 500, err, "failed to marshal job payload")
	}

	_, err = c.rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: stream,
		ID:     "*", // biarkan Redis generate ID
		Values: map[string]interface{}{
			"id":   payload.ID,
			"data": string(jsonPayload),
		},
	}).Result()
	if err != nil {
		return apperror.New("redisx", "EnqueueJob", 500, err, "failed to enqueue job")
	}
	return nil
}

func QueuePDFUpload(ctx context.Context, client *Client, filePath string) error {
	job := Job{
		ID: cuid2.Generate(),
		Payload: map[string]interface{}{
			"file_path": filePath,
			"folder":    "bukti_pendaftaran",
			"name":      filepath.Base(filePath),
		},
	}

	return client.EnqueueJob(ctx, "pdf_upload_jobs", job)
}

// EnqueueJobFile menambahkan job ke Redis Stream dengan menyimpan payload ke file
func (c *Client) EnqueueJobFile(ctx context.Context, payload Job) error {
	// Pastikan folder tmp ada
	if err := os.MkdirAll("tmp", 0755); err != nil {
		return apperror.New("redisx", "EnqueueJobFile", 500, err, "failed to create tmp dir")
	}

	// Cast to the concrete type so we can access the file data
	qf, ok := payload.Payload.(fileUploader.QueueUploadFile)
	if !ok {
		return apperror.New("redisx", "EnqueueJobFile", 500, nil, "payload is not of type QueueUploadFile")
	}

	// Buat nama file unik di tmp dengan ekstensi asli
	if qf.File == nil {
		return apperror.New("redisx", "EnqueueJobFile", 400, nil, "QueueUploadFile.File cannot be nil")
	}

	// Dapatkan ekstensi dari nama file asli
	ext := filepath.Ext(qf.File.Filename)

	// Baca konten file dari *multipart.FileHeader
	src, err := qf.File.Open()
	if err != nil {
		return apperror.New("redisx", "EnqueueJobFile", 500, err, "failed to open multipart file")
	}
	defer src.Close()

	data, err := io.ReadAll(src)
	if err != nil {
		return apperror.New("redisx", "EnqueueJobFile", 500, err, "failed to read multipart file")
	}

	// Buat nama file unik di tmp
	fileName := fmt.Sprintf("tmp/%s_%d%s", payload.ID, time.Now().UnixNano(), ext)
	qf.FilePathTmp = &fileName
	payload.Payload = qf

	// Tulis file gambarnya ke tmp
	if err := os.WriteFile(fileName, data, 0644); err != nil {
		return apperror.New("redisx", "EnqueueJobFile", 500, err, "failed to write image file to tmp")
	}

	// Marshal payload tanpa FileData (sudah tersimpan di tmp)
	jsonPayload, err := json.Marshal(payload.Payload)
	if err != nil {
		return apperror.New("redisx", "EnqueueJobFile", 500, err, "failed to marshal job payload")
	}

	// Enqueue job dengan path file
	_, err = c.rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: "upload_jobs",
		ID:     "*", // biarkan Redis generate ID
		Values: map[string]interface{}{
			"id":      payload.ID,
			"payload": string(jsonPayload),
		},
	}).Result()
	if err != nil {
		return apperror.New("redisx", "EnqueueJobFile", 500, err, "failed to enqueue job with file")
	}
	return nil
}

// InitConsumerGroup buat consumer group
func (c *Client) InitConsumerGroup(ctx context.Context, stream, group string) error {
	if err := c.rdb.XGroupCreateMkStream(ctx, stream, group, "0").Err(); err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		return apperror.New("redisx", "InitConsumerGroup", 500, err, "failed to create consumer group")
	}
	return nil
}

// ConsumeJob baca job dari stream
func (c *Client) ConsumeJob(ctx context.Context, stream, group, consumer string, count int, block time.Duration) ([]Job, error) {
	res, err := c.rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    group,
		Consumer: consumer,
		Streams:  []string{stream, ">"},
		Count:    int64(count),
		Block:    block,
	}).Result()

	if err == redis.Nil || len(res) == 0 {
		return nil, nil
	}
	if err != nil {
		return nil, apperror.New("redisx", "ConsumeJob", 500, err, "failed to consume job")
	}

	var jobs []Job
	for _, msg := range res[0].Messages {
		payload := make(map[string]interface{})
		for k, v := range msg.Values {
			payload[k] = v
		}
		jobs = append(jobs, Job{ID: msg.ID, Payload: payload})
	}
	return jobs, nil
}

// AckJob tandai job selesai
func (c *Client) AckJob(ctx context.Context, stream, group string, msgID string) error {
	if err := c.rdb.XAck(ctx, stream, group, msgID).Err(); err != nil {
		return apperror.New("redisx", "AckJob", 500, err, "failed to ack job")
	}
	return nil
}

// ================== PUB / SUB ==================

// Publish kirim pesan ke channel
func (c *Client) Publish(ctx context.Context, channel, message string) error {
	if err := c.rdb.Publish(ctx, channel, message).Err(); err != nil {
		return apperror.New("redisx", "Publish", 500, err, "failed to publish message")
	}
	return nil
}

// Subscribe mendengarkan channel
func (c *Client) Subscribe(ctx context.Context, channel string) (<-chan *redis.Message, error) {
	sub := c.rdb.Subscribe(ctx, channel)
	_, err := sub.Receive(ctx)
	if err != nil {
		return nil, apperror.New("redisx", "Subscribe", 500, err, "failed to subscribe")
	}
	return sub.Channel(), nil
}

// ================== Debug Helper ==================
func (c *Client) Ping(ctx context.Context) error {
	val, err := c.rdb.Ping(ctx).Result()
	if err != nil {
		log.Println("Redis connection error:", err)
	} else {
		fmt.Println("Redis PING:", val)
	}
	return nil
}
