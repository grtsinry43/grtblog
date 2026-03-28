package media

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/disintegration/imaging"

	appEvent "github.com/grtsinry43/grtblog-v2/server/internal/app/event"
	"github.com/grtsinry43/grtblog-v2/server/internal/domain/media"
)

type Service struct {
	repo      media.Repository
	uploadDir string
	events    appEvent.Bus
}

func NewService(repo media.Repository, uploadDir string, events appEvent.Bus) *Service {
	trimmed := strings.TrimSpace(uploadDir)
	if trimmed == "" {
		trimmed = filepath.Join("storage", "uploads")
	}
	if events == nil {
		events = appEvent.NopBus{}
	}
	return &Service{
		repo:      repo,
		uploadDir: trimmed,
		events:    events,
	}
}

const thumbnailMaxWidth = 1200
const thumbnailDir = "thumbnails"
const thumbnailQuality = 82

// ImageMeta 图片元信息，上传图片时自动提取。
type ImageMeta struct {
	Width         int    `json:"width,omitempty"`
	Height        int    `json:"height,omitempty"`
	DominantColor string `json:"dominantColor,omitempty"` // hex, e.g. "#a3b2c1"
}

type UploadResult struct {
	File         media.UploadFile
	Created      bool
	ThumbnailURL string     // 缩略图公开路径（仅 picture 类型）
	ImageMeta    *ImageMeta // 图片元信息（仅 picture 类型）
}

func (s *Service) Upload(ctx context.Context, file *multipart.FileHeader, fileType string) (*UploadResult, error) {
	if file == nil {
		return nil, errors.New("file is required")
	}

	dir, err := dirForType(fileType)
	if err != nil {
		return nil, err
	}

	hash, err := hashFile(file)
	if err != nil {
		return nil, err
	}

	existing, err := s.repo.FindByHash(ctx, hash)
	if err != nil && !errors.Is(err, media.ErrUploadFileNotFound) {
		return nil, err
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	filename := s.buildFilename(dir, ext)
	storedPath := "/" + dir + "/" + filename
	diskPath := s.diskPathFromStored(storedPath)

	if existing != nil {
		existingDisk := s.diskPathFromStored(existing.Path)
		if fileExists(existingDisk) {
			thumbURL, meta := s.processImage(existingDisk, existing.Path, dir)
			return &UploadResult{File: *existing, Created: false, ThumbnailURL: thumbURL, ImageMeta: meta}, nil
		}
		if err := s.saveFile(file, diskPath); err != nil {
			return nil, err
		}
		if existing.Path != storedPath {
			if err := s.repo.UpdatePath(ctx, existing.ID, storedPath); err != nil {
				return nil, err
			}
			existing.Path = storedPath
		}
		thumbURL, meta := s.processImage(diskPath, storedPath, dir)
		return &UploadResult{File: *existing, Created: false, ThumbnailURL: thumbURL, ImageMeta: meta}, nil
	}

	if err := s.saveFile(file, diskPath); err != nil {
		return nil, err
	}

	record := &media.UploadFile{
		Name: file.Filename,
		Path: storedPath,
		Type: strings.ToLower(strings.TrimSpace(fileType)),
		Size: file.Size,
		Hash: hash,
	}
	if err := s.repo.Create(ctx, record); err != nil {
		return nil, err
	}
	_ = s.events.Publish(ctx, appEvent.Generic{
		EventName: "media.uploaded",
		At:        time.Now(),
		Payload: map[string]any{
			"ID":   record.ID,
			"Name": record.Name,
			"Path": record.Path,
			"Type": record.Type,
			"Size": record.Size,
		},
	})
	thumbURL, meta := s.processImage(diskPath, storedPath, dir)
	return &UploadResult{File: *record, Created: true, ThumbnailURL: thumbURL, ImageMeta: meta}, nil
}

type ListResult struct {
	Items []media.UploadFile
	Total int64
	Page  int
	Size  int
}

func (s *Service) List(ctx context.Context, page int, size int) (*ListResult, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	if size > 100 {
		size = 100
	}
	offset := (page - 1) * size
	items, total, err := s.repo.List(ctx, offset, size)
	if err != nil {
		return nil, err
	}
	return &ListResult{
		Items: items,
		Total: total,
		Page:  page,
		Size:  size,
	}, nil
}

func (s *Service) Rename(ctx context.Context, id int64, name string) (*media.UploadFile, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return nil, errors.New("name is required")
	}
	file, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if file.Name == trimmed {
		return file, nil
	}
	if err := s.repo.UpdateName(ctx, id, trimmed); err != nil {
		return nil, err
	}
	file.Name = trimmed
	return file, nil
}

func (s *Service) Delete(ctx context.Context, id int64) (*media.UploadFile, error) {
	file, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	diskPath := s.diskPathFromStored(file.Path)
	if err := removeFile(diskPath); err != nil {
		return nil, err
	}
	if err := s.repo.DeleteByID(ctx, id); err != nil {
		return nil, err
	}
	_ = s.events.Publish(ctx, appEvent.Generic{
		EventName: "media.deleted",
		At:        time.Now(),
		Payload: map[string]any{
			"ID":   file.ID,
			"Name": file.Name,
			"Path": file.Path,
			"Type": file.Type,
			"Size": file.Size,
		},
	})
	return file, nil
}

func (s *Service) GetByID(ctx context.Context, id int64) (*media.UploadFile, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) ResolveDiskPath(storedPath string) (string, error) {
	diskPath := s.diskPathFromStored(storedPath)
	if diskPath == "" {
		return "", errors.New("empty stored path")
	}
	return diskPath, nil
}

// processImage 为图片生成缩略图并提取元信息（尺寸 + 主色调）。
func (s *Service) processImage(diskPath string, storedPath string, dir string) (thumbURL string, meta *ImageMeta) {
	if dir != "pictures" {
		return "", nil
	}

	f, err := os.Open(diskPath)
	if err != nil {
		log.Printf("[image] open failed for %s: %v", diskPath, err)
		return "", nil
	}
	defer f.Close()

	src, _, err := image.Decode(f)
	if err != nil {
		log.Printf("[image] decode failed for %s: %v", diskPath, err)
		return "", nil
	}

	bounds := src.Bounds()
	meta = &ImageMeta{
		Width:         bounds.Dx(),
		Height:        bounds.Dy(),
		DominantColor: calcDominantColor(src),
	}

	// Generate thumbnail
	thumbStoredPath := "/" + thumbnailDir + storedPath
	thumbDiskPath := s.diskPathFromStored(thumbStoredPath)

	if !fileExists(thumbDiskPath) {
		thumb := imaging.Resize(src, thumbnailMaxWidth, 0, imaging.Lanczos)
		if err := os.MkdirAll(filepath.Dir(thumbDiskPath), 0o755); err != nil {
			log.Printf("[thumbnail] mkdir failed: %v", err)
			return "", meta
		}
		out, err := os.Create(thumbDiskPath)
		if err != nil {
			log.Printf("[thumbnail] create failed: %v", err)
			return "", meta
		}
		defer out.Close()
		if err := jpeg.Encode(out, thumb, &jpeg.Options{Quality: thumbnailQuality}); err != nil {
			log.Printf("[thumbnail] encode failed: %v", err)
			return "", meta
		}
	}

	return "/uploads" + thumbStoredPath, meta
}

// calcDominantColor 采样缩小后取平均色。
func calcDominantColor(img image.Image) string {
	small := imaging.Resize(img, 32, 0, imaging.Box)
	bounds := small.Bounds()
	var r, g, b uint64
	var count uint64
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			cr, cg, cb, ca := small.At(x, y).RGBA()
			if ca < 0x1000 {
				continue
			}
			r += uint64(cr >> 8)
			g += uint64(cg >> 8)
			b += uint64(cb >> 8)
			count++
		}
	}
	if count == 0 {
		return ""
	}
	return fmt.Sprintf("#%02x%02x%02x", r/count, g/count, b/count)
}

// ThumbnailURLFor 根据原图公开 URL 返回对应缩略图的公开 URL。
// 如果缩略图不存在于磁盘，返回空字符串。
func (s *Service) ThumbnailURLFor(publicURL string) string {
	// publicURL = /uploads/pictures/2026-...
	const prefix = "/uploads"
	if !strings.HasPrefix(publicURL, prefix) {
		return ""
	}
	storedPath := strings.TrimPrefix(publicURL, prefix) // /pictures/2026-...
	thumbStoredPath := "/" + thumbnailDir + storedPath
	thumbDiskPath := s.diskPathFromStored(thumbStoredPath)
	if fileExists(thumbDiskPath) {
		return prefix + thumbStoredPath
	}
	return ""
}

// ExtractImageMetaFromURL 根据本站公开 URL 提取图片元信息（尺寸+主色调）并确保缩略图存在。
// 外链返回 nil。
func (s *Service) ExtractImageMetaFromURL(publicURL string) (thumbURL string, meta *ImageMeta) {
	const prefix = "/uploads"
	if !strings.HasPrefix(publicURL, prefix) {
		return "", nil
	}
	storedPath := strings.TrimPrefix(publicURL, prefix)
	diskPath := s.diskPathFromStored(storedPath)
	if !fileExists(diskPath) {
		return "", nil
	}
	return s.processImage(diskPath, storedPath, "pictures")
}

func (s *Service) saveFile(file *multipart.FileHeader, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if fileExists(path) {
		return nil
	}
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	return err
}

func (s *Service) diskPathFromStored(storedPath string) string {
	trimmed := strings.TrimSpace(storedPath)
	if trimmed == "" {
		return ""
	}
	clean := filepath.Clean(trimmed)
	clean = strings.TrimPrefix(clean, string(filepath.Separator))
	uploadDir := filepath.Clean(s.uploadDir)
	if strings.HasPrefix(clean, uploadDir+string(filepath.Separator)) || clean == uploadDir {
		return clean
	}
	return filepath.Join(uploadDir, clean)
}

func (s *Service) buildFilename(dir string, ext string) string {
	base := time.Now().Format("2006-01-02-15:04:05")
	ext = strings.TrimSpace(ext)
	if ext != "" && !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	for i := 0; i < 5; i++ {
		suffix := randomHex(2)
		filename := base + "-" + suffix + ext
		if !fileExists(filepath.Join(s.uploadDir, dir, filename)) {
			return filename
		}
	}
	suffix := randomHex(4)
	return base + "-" + suffix + ext
}

func randomHex(n int) string {
	if n <= 0 {
		return ""
	}
	byteLen := (n + 1) / 2
	buf := make([]byte, byteLen)
	if _, err := rand.Read(buf); err == nil {
		return hex.EncodeToString(buf)[:n]
	}
	fallback := hex.EncodeToString([]byte(time.Now().Format("150405.000")))
	if len(fallback) >= n {
		return fallback[:n]
	}
	return fallback
}

func dirForType(fileType string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(fileType)) {
	case "picture":
		return "pictures", nil
	case "file":
		return "files", nil
	default:
		return "", media.ErrInvalidUploadType
	}
}

func hashFile(file *multipart.FileHeader) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, src); err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func fileExists(path string) bool {
	if strings.TrimSpace(path) == "" {
		return false
	}
	_, err := os.Stat(path)
	return err == nil
}

func removeFile(path string) error {
	if strings.TrimSpace(path) == "" {
		return nil
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
