package fileUploader

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"template-golang/pkg/apperror"
	"template-golang/pkg/config"
	"template-golang/pkg/logger"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/chai2010/webp"
)

var (
	sessionInstance *session.Session
	svcInstance     *s3.S3
	once            sync.Once
)

type QueueUploadFile struct {
	FilePath         string
	IsCompressToWebp *bool
	File             *multipart.FileHeader
	FilePathTmp      *string
	OldFile          *string
}

// InitS3Client initialize singleton S3 client
func InitS3Client() error {
	var err error
	cfg := config.GetConfig()
	once.Do(func() {
		region := cfg.S3Region
		if region == "" {
			region = "us-west-2"
		}
		sessionInstance, err = session.NewSession(&aws.Config{
			Region:   aws.String(region),
			Endpoint: aws.String(cfg.S3End),
			Credentials: credentials.NewStaticCredentials(
				cfg.S3Access,
				cfg.S3Secret,
				"",
			),
		})
		if err == nil && sessionInstance != nil {
			svcInstance = s3.New(sessionInstance)
		}
	})

	if svcInstance == nil {
		return fmt.Errorf("S3 session is not initialized")
	}
	return err
}

type FileUploadOptions struct {
	Folder           string
	NameFile         string
	MaxSizeMB        *int64
	AllowedMimeTypes []string
	IsCompressToWebp *bool
}

func ExtractFolderFromFilePath(filePath string) string {
	// If filePath is a URL (e.g. https://is3.***/bucket/folder/file.jpg)
	if strings.HasPrefix(strings.ToLower(filePath), "http://") || strings.HasPrefix(strings.ToLower(filePath), "https://") {
		parts := strings.Split(filePath, "/")
		// Find bucket index (e.g. is3.***)
		bucketIndex := -1
		for i, part := range parts {
			if strings.Contains(part, "is3") { // Adjust to your bucket pattern
				bucketIndex = i + 1 // Bucket index is usually followed by bucket name
				break
			}
		}
		if bucketIndex != -1 && bucketIndex+1 < len(parts)-1 {
			// Take the single folder name right after the bucket
			return parts[bucketIndex+1]
		}
		return ""
	}
	// If not a URL, use filepath.Dir as usual
	return filepath.Dir(filePath)
}

// GenerateFileURLCompressed hanya buat URL file di S3
func GenerateFileURLCompressed(folder string, image *multipart.FileHeader) (string, error) {
	if err := InitS3Client(); err != nil {
		return "", err
	}
	if folder == "" {
		return "", fmt.Errorf("folder cannot be empty")
	}
	if image == nil {
		return "", fmt.Errorf("image cannot be nil")
	}

	// Jika filename sudah URL → return langsung
	filename := image.Filename
	if strings.HasPrefix(strings.ToLower(filename), "http://") ||
		strings.HasPrefix(strings.ToLower(filename), "https://") {
		return filename, nil
	}

	// Generate nama unik
	timestamp := time.Now().UnixNano()
	ext := ".webp" // default ke webp
	key := fmt.Sprintf("%s/%d%s", strings.Trim(folder, "/"), timestamp, ext)

	cfg := config.GetConfig()
	endpoint := strings.TrimRight(cfg.S3End, "/")
	if !strings.HasPrefix(endpoint, "http") {
		endpoint = "https://" + endpoint
	}

	fileURL := fmt.Sprintf("%s/%s/%s", endpoint, cfg.S3Bucket, key)
	return fileURL, nil
}

func ValidateFileOptions(file *multipart.FileHeader, opts FileUploadOptions) error {
	if opts.Folder == "" {
		return apperror.New("BAD", "folder cannot be empty", 400, nil, string(debug.Stack()))
	}
	if opts.NameFile == "" {
		return apperror.New("BAD", "nameFile cannot be empty", 400, nil, string(debug.Stack()))
	}
	if file == nil {
		return apperror.New("BAD", "file cannot be nil", 400, nil, string(debug.Stack()))
	}
	if opts.MaxSizeMB != nil && file.Size > *opts.MaxSizeMB*1024*1024 {
		return apperror.New("BAD", fmt.Sprintf("file exceeds maximum size of %d MB", *opts.MaxSizeMB), 400, nil, string(debug.Stack()))
	}
	if len(opts.AllowedMimeTypes) > 0 {
		contentType := file.Header.Get("Content-Type")
		allowed := false
		for _, t := range opts.AllowedMimeTypes {
			if contentType == t {
				allowed = true
				break
			}
		}
		if !allowed {
			return apperror.New("BAD", fmt.Sprintf("file type %s is not allowed", contentType), 400, nil, string(debug.Stack()))
		}
	}
	return nil
}

// UploadFile upload file ke S3, support compress ke WebP
func UploadFile(ctx context.Context, file *multipart.FileHeader, opts FileUploadOptions) (string, error) {
	if err := InitS3Client(); err != nil {
		return "", err
	}

	if file == nil {
		return "", fmt.Errorf("file cannot be nil")
	}
	if opts.NameFile == "" {
		return "", fmt.Errorf("NameFile cannot be empty")
	}

	// Validasi size
	if opts.MaxSizeMB != nil && file.Size > *opts.MaxSizeMB*1024*1024 {
		return "", fmt.Errorf("file exceeds maximum size of %d MB", *opts.MaxSizeMB)
	}

	// Validasi MIME
	if len(opts.AllowedMimeTypes) > 0 {
		contentType := file.Header.Get("Content-Type")
		allowed := false
		for _, t := range opts.AllowedMimeTypes {
			if contentType == t {
				allowed = true
				break
			}
		}
		if !allowed {
			return "", apperror.New("BAD", fmt.Sprintf("file type %s is not allowed", contentType), 400, nil, string(debug.Stack()))
		}
	}

	// Open file
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer src.Close()

	// Read file into bytes
	data, err := io.ReadAll(src)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	contentType := file.Header.Get("Content-Type")
	filename := filepath.Base(opts.NameFile)

	// Kompres ke WebP jika diminta
	if opts.IsCompressToWebp != nil && *opts.IsCompressToWebp {
		img, _, err := image.Decode(bytes.NewReader(data))
		if err != nil {
			return "", fmt.Errorf("failed to decode image: %w", err)
		}

		buf, err := webp.EncodeRGB(img, 80)
		if err != nil {
			return "", fmt.Errorf("failed to encode webp: %w", err)
		}

		data = buf
		contentType = "image/webp"
		filename = strings.TrimSuffix(filename, filepath.Ext(filename)) + ".webp"

		logger.L().Printf("[UploadFile] compressed to webp, final size: %d bytes", len(data))
	} else {
		// Pastikan kalau bukan webp tetap sesuai ext
		switch contentType {
		case "image/png":
			img, err := png.Decode(bytes.NewReader(data))
			if err == nil {
				var buf bytes.Buffer
				_ = png.Encode(&buf, img)
				data = buf.Bytes()
			}
		case "image/jpeg":
			img, err := jpeg.Decode(bytes.NewReader(data))
			if err == nil {
				var buf bytes.Buffer
				_ = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90})
				data = buf.Bytes()
			}
		}
	}

	// Buat key relatif
	key := fmt.Sprintf("%s/%s", strings.Trim(opts.Folder, "/"), filename)

	// Upload ke S3
	cfg := config.GetConfig()
	_, err = svcInstance.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(cfg.S3Bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
		ACL:         aws.String("public-read"),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	// Generate URL publik
	endpoint := strings.TrimRight(cfg.S3End, "/")
	if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		endpoint = "https://" + endpoint
	}
	fileURL := fmt.Sprintf("%s/%s/%s", endpoint, cfg.S3Bucket, key)

	logger.L().Printf("[UploadFile] successfully uploaded file: %s", fileURL)
	return fileURL, nil
}

func UploadFileFromPath(ctx context.Context, filePath string, opts FileUploadOptions) error {
	if err := InitS3Client(); err != nil {
		return err
	}

	if filePath == "" {
		return fmt.Errorf("filePath cannot be empty")
	}
	if opts.NameFile == "" {
		return fmt.Errorf("NameFile cannot be empty")
	}

	// open file from local path
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	// get size
	fi, err := f.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}
	fileSize := fi.Size()

	// Validasi size
	if opts.MaxSizeMB != nil && fileSize > *opts.MaxSizeMB*1024*1024 {
		return fmt.Errorf("file exceeds maximum size of %d MB", *opts.MaxSizeMB)
	}

	// detect MIME
	buf := make([]byte, 512)
	n, _ := io.ReadFull(f, buf)
	contentType := http.DetectContentType(buf[:n])
	// reset cursor
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("failed to rewind file: %w", err)
	}

	// Validasi MIME
	if len(opts.AllowedMimeTypes) > 0 {
		allowed := false
		for _, t := range opts.AllowedMimeTypes {
			if contentType == t {
				allowed = true
				break
			}
		}
		if !allowed {
			return apperror.New("BAD", fmt.Sprintf("file type %s is not allowed", contentType), 400, nil, string(debug.Stack()))
		}
	}

	// Read full file
	data, err := io.ReadAll(f)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	filename := filepath.Base(opts.NameFile)

	// Kompres ke WebP jika diminta
	if opts.IsCompressToWebp != nil && *opts.IsCompressToWebp {
		img, _, err := image.Decode(bytes.NewReader(data))
		if err != nil {
			return fmt.Errorf("failed to decode image: %w", err)
		}

		buf, err := webp.EncodeRGB(img, 10)
		if err != nil {
			return fmt.Errorf("failed to encode webp: %w", err)
		}

		data = buf
		contentType = "image/webp"
		filename = strings.TrimSuffix(filename, filepath.Ext(filename)) + ".webp"

		logger.L().Printf("[UploadFileFromPath] compressed to webp, final size: %d bytes", len(data))
	} else {
		// Pastikan kalau bukan webp tetap sesuai ext
		switch contentType {
		case "image/png":
			img, err := png.Decode(bytes.NewReader(data))
			if err == nil {
				var buf bytes.Buffer
				_ = png.Encode(&buf, img)
				data = buf.Bytes()
			}
		case "image/jpeg":
			img, err := jpeg.Decode(bytes.NewReader(data))
			if err == nil {
				var buf bytes.Buffer
				_ = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90})
				data = buf.Bytes()
			}
		}
	}

	// Buat key relatif
	key := fmt.Sprintf("%s/%s", strings.Trim(opts.Folder, "/"), filename)

	// Upload ke S3
	cfg := config.GetConfig()
	_, err = svcInstance.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(cfg.S3Bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
		ACL:         aws.String("public-read"),
	})
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	// Generate URL publik
	endpoint := strings.TrimRight(cfg.S3End, "/")
	if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		endpoint = "https://" + endpoint
	}
	fileURL := fmt.Sprintf("%s/%s/%s", endpoint, cfg.S3Bucket, key)

	logger.L().Printf("[UploadFileFromPath] successfully uploaded file: %s", fileURL)
	return nil
}

// GenerateFileURL returns a public URL for a file (without uploading)
func GenerateFileURL(folder string, image *multipart.FileHeader) (string, error) {
	if err := InitS3Client(); err != nil {
		return "", err
	}
	if folder == "" {
		return "", fmt.Errorf("folder cannot be empty")
	}
	if image == nil {
		return "", fmt.Errorf("image cannot be nil")
	}

	// Jika filename sudah URL, return langsung
	filename := image.Filename
	if strings.HasPrefix(strings.ToLower(filename), "http://") || strings.HasPrefix(strings.ToLower(filename), "https://") {
		return filename, nil
	}

	// Buat key baru
	timestamp := time.Now().UnixNano()
	ext := filepath.Ext(filename)
	key := fmt.Sprintf("%s/%d%s", strings.Trim(folder, "/"), timestamp, ext)

	cfg := config.GetConfig()
	endpoint := strings.TrimRight(cfg.S3End, "/")
	if !strings.HasPrefix(endpoint, "http") {
		endpoint = "https://" + endpoint
	}

	fileURL := fmt.Sprintf("%s/%s/%s", endpoint, cfg.S3Bucket, key)
	return fileURL, nil
}

// DeleteFile deletes a file from S3 by its public URL
func DeleteFile(ctx context.Context, fileURL string) error {
	if err := InitS3Client(); err != nil {
		return err
	}

	folder := ExtractFolderFromFilePath(fileURL)
	if folder == "example" {
		return nil
	}

	if fileURL == "" {
		return fmt.Errorf("file URL cannot be empty")
	}

	// Parse URL untuk mendapatkan key
	parts := strings.Split(fileURL, "/")
	if len(parts) < 2 {
		return fmt.Errorf("invalid file URL format")
	}

	bucketName := config.GetConfig().S3Bucket
	urlBucketIndex := -1
	for i, part := range parts {
		if part == bucketName {
			urlBucketIndex = i
			break
		}
	}

	if urlBucketIndex == -1 || urlBucketIndex >= len(parts)-1 {
		return fmt.Errorf("could not find valid key in URL")
	}

	key := strings.Join(parts[urlBucketIndex+1:], "/")

	_, err := svcInstance.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}
