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
	"io/fs"
	"log"
	"math"
	"mime/multipart"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	goexif "github.com/rwcarlsen/goexif/exif"

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

type SyncResult struct {
	Scanned           int
	Indexed           int
	Created           int
	Updated           int
	Deleted           int
	SkippedDuplicates int
}

type indexedDiskFile struct {
	Name string
	Path string
	Type string
	Size int64
	Hash string
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

func (s *Service) SyncIndex(ctx context.Context) (*SyncResult, error) {
	existing, err := s.repo.ListAll(ctx)
	if err != nil {
		return nil, err
	}

	diskFiles, scanned, skippedDuplicates, err := s.scanUploadFiles()
	if err != nil {
		return nil, err
	}

	result := &SyncResult{
		Scanned:           scanned,
		Indexed:           len(diskFiles),
		SkippedDuplicates: skippedDuplicates,
	}

	existingByPath := make(map[string]*media.UploadFile, len(existing))
	existingByHash := make(map[string]*media.UploadFile, len(existing))
	for i := range existing {
		file := &existing[i]
		existingByPath[file.Path] = file
		if strings.TrimSpace(file.Hash) != "" {
			existingByHash[file.Hash] = file
		}
	}

	usedIDs := make(map[int64]struct{}, len(diskFiles))
	for _, diskFile := range diskFiles {
		pathRecord := existingByPath[diskFile.Path]
		hashRecord := existingByHash[diskFile.Hash]

		target := resolveSyncTarget(pathRecord, hashRecord)
		if target == nil {
			record := &media.UploadFile{
				Name: diskFile.Name,
				Path: diskFile.Path,
				Type: diskFile.Type,
				Size: diskFile.Size,
				Hash: diskFile.Hash,
			}
			if err := s.repo.Create(ctx, record); err != nil {
				return nil, err
			}
			result.Created++
			usedIDs[record.ID] = struct{}{}
			existingByPath[record.Path] = record
			existingByHash[record.Hash] = record
			continue
		}

		usedIDs[target.ID] = struct{}{}
		if !needsSyncUpdate(*target, diskFile) {
			continue
		}

		originalPath := target.Path
		originalHash := target.Hash
		target.Name = diskFile.Name
		target.Path = diskFile.Path
		target.Type = diskFile.Type
		target.Size = diskFile.Size
		target.Hash = diskFile.Hash
		if err := s.repo.Update(ctx, target); err != nil {
			return nil, err
		}
		result.Updated++
		if originalPath != target.Path {
			delete(existingByPath, originalPath)
		}
		existingByPath[target.Path] = target
		if strings.TrimSpace(originalHash) != "" && originalHash != target.Hash {
			delete(existingByHash, originalHash)
		}
		existingByHash[target.Hash] = target
	}

	for i := range existing {
		file := &existing[i]
		if _, ok := usedIDs[file.ID]; ok {
			continue
		}
		if err := s.repo.DeleteByID(ctx, file.ID); err != nil {
			return nil, err
		}
		result.Deleted++
	}

	return result, nil
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
	thumbURL, meta, _ = s.ExtractPhotoMetadataFromURL(publicURL)
	return thumbURL, meta
}

// ExtractPhotoMetadataFromURL 根据本站公开 URL 提取图片元信息和 EXIF 摘要。
// 外链或不存在的本地文件返回空结果。
func (s *Service) ExtractPhotoMetadataFromURL(publicURL string) (thumbURL string, meta *ImageMeta, exifData map[string]any) {
	const prefix = "/uploads"
	if !strings.HasPrefix(publicURL, prefix) {
		return "", nil, nil
	}
	storedPath := strings.TrimPrefix(publicURL, prefix)
	diskPath := s.diskPathFromStored(storedPath)
	if !fileExists(diskPath) {
		return "", nil, nil
	}
	thumbURL, meta = s.processImage(diskPath, storedPath, "pictures")
	exifData = extractExifSummary(diskPath)
	return thumbURL, meta, exifData
}

func resolveSyncTarget(pathRecord *media.UploadFile, hashRecord *media.UploadFile) *media.UploadFile {
	switch {
	case pathRecord != nil && hashRecord != nil:
		if pathRecord.ID == hashRecord.ID {
			return pathRecord
		}
		return hashRecord
	case pathRecord != nil:
		return pathRecord
	default:
		return hashRecord
	}
}

func needsSyncUpdate(record media.UploadFile, diskFile indexedDiskFile) bool {
	return record.Name != diskFile.Name ||
		record.Path != diskFile.Path ||
		record.Type != diskFile.Type ||
		record.Size != diskFile.Size ||
		record.Hash != diskFile.Hash
}

func (s *Service) scanUploadFiles() ([]indexedDiskFile, int, int, error) {
	root := filepath.Clean(s.uploadDir)
	candidatesByHash := make(map[string]indexedDiskFile)
	scanned := 0
	skippedDuplicates := 0

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			if path == root {
				return nil
			}
			if s.shouldSkipSyncDir(path) {
				return filepath.SkipDir
			}
			return nil
		}
		if !d.Type().IsRegular() {
			return nil
		}

		scanned++
		hash, err := hashDiskPath(path)
		if err != nil {
			return err
		}
		if _, exists := candidatesByHash[hash]; exists {
			skippedDuplicates++
			return nil
		}

		relPath, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}

		storedPath := "/" + filepath.ToSlash(relPath)
		candidatesByHash[hash] = indexedDiskFile{
			Name: filepath.Base(path),
			Path: storedPath,
			Type: detectIndexedFileType(storedPath),
			Size: info.Size(),
			Hash: hash,
		}
		return nil
	})
	if err != nil {
		return nil, 0, 0, err
	}

	candidates := make([]indexedDiskFile, 0, len(candidatesByHash))
	for _, file := range candidatesByHash {
		candidates = append(candidates, file)
	}
	slices.SortFunc(candidates, func(a, b indexedDiskFile) int {
		return strings.Compare(a.Path, b.Path)
	})
	return candidates, scanned, skippedDuplicates, nil
}

func (s *Service) shouldSkipSyncDir(path string) bool {
	root := filepath.Clean(s.uploadDir)
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	first := filepath.ToSlash(rel)
	if idx := strings.IndexByte(first, '/'); idx >= 0 {
		first = first[:idx]
	}
	return first == thumbnailDir
}

func detectIndexedFileType(storedPath string) string {
	clean := strings.ToLower(filepath.ToSlash(strings.TrimSpace(storedPath)))
	if strings.HasPrefix(clean, "/pictures/") {
		return "picture"
	}

	switch filepath.Ext(clean) {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp", ".svg", ".avif", ".heic", ".heif", ".tif", ".tiff":
		return "picture"
	default:
		return "file"
	}
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

func hashDiskPath(path string) (string, error) {
	src, err := os.Open(path)
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

func extractExifSummary(diskPath string) map[string]any {
	file, err := os.Open(diskPath)
	if err != nil {
		return nil
	}
	defer file.Close()

	x, err := goexif.Decode(file)
	if err != nil {
		return nil
	}

	result := map[string]any{}
	if makeValue := exifString(x, goexif.Make); makeValue != "" {
		result["make"] = makeValue
	}
	if modelValue := exifString(x, goexif.Model); modelValue != "" {
		result["model"] = modelValue
	}
	if lensModel := exifString(x, goexif.LensModel); lensModel != "" {
		result["lensModel"] = lensModel
	}
	if focalLength := exifFocalLength(x); focalLength != "" {
		result["focalLength"] = focalLength
	}
	if fNumber := exifDecimal(x, goexif.FNumber, 2); fNumber != "" {
		result["fNumber"] = fNumber
	}
	if exposureTime := exifExposureTime(x); exposureTime != "" {
		result["exposureTime"] = exposureTime
	}
	if iso, ok := exifInt(x, goexif.ISOSpeedRatings); ok && iso > 0 {
		result["iso"] = iso
	}
	if dateTime, ok := exifDateTime(x); ok {
		result["dateTimeOriginal"] = dateTime
	}
	if lat, long, ok := exifLatLong(x); ok {
		result["gpsLatitude"] = lat
		result["gpsLongitude"] = long
	}
	if altitude, ok := exifAltitude(x); ok {
		result["gpsAltitude"] = altitude
	}
	if width, ok := exifInt(x, goexif.PixelXDimension); ok && width > 0 {
		result["imageWidth"] = width
	}
	if height, ok := exifInt(x, goexif.PixelYDimension); ok && height > 0 {
		result["imageHeight"] = height
	}
	if orientation, ok := exifInt(x, goexif.Orientation); ok && orientation > 0 {
		result["orientation"] = orientation
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func exifString(x *goexif.Exif, field goexif.FieldName) string {
	tag, err := x.Get(field)
	if err != nil {
		return ""
	}
	value, err := tag.StringVal()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(value)
}

func exifInt(x *goexif.Exif, field goexif.FieldName) (int, bool) {
	tag, err := x.Get(field)
	if err != nil {
		return 0, false
	}
	value, err := tag.Int(0)
	if err != nil {
		return 0, false
	}
	return value, true
}

func exifDecimal(x *goexif.Exif, field goexif.FieldName, precision int) string {
	tag, err := x.Get(field)
	if err != nil {
		return ""
	}
	value, ok := exifRat(tag)
	if !ok {
		return ""
	}
	return formatExifFloat(value, precision)
}

func exifFocalLength(x *goexif.Exif) string {
	tag, err := x.Get(goexif.FocalLength)
	if err != nil {
		return ""
	}
	value, ok := exifRat(tag)
	if !ok {
		return ""
	}
	return formatExifFloat(value, 2) + " mm"
}

func exifExposureTime(x *goexif.Exif) string {
	tag, err := x.Get(goexif.ExposureTime)
	if err != nil {
		return ""
	}
	num, den, err := tag.Rat2(0)
	if err != nil || den == 0 {
		return ""
	}
	if num > 0 && num < den {
		return fmt.Sprintf("%d/%d", num, den)
	}
	return formatExifFloat(float64(num)/float64(den), 4)
}

func exifDateTime(x *goexif.Exif) (string, bool) {
	tm, err := x.DateTime()
	if err != nil || tm.IsZero() {
		return "", false
	}
	return tm.Format("2006:01:02 15:04:05"), true
}

func exifLatLong(x *goexif.Exif) (float64, float64, bool) {
	lat, long, err := x.LatLong()
	if err != nil {
		return 0, 0, false
	}
	return lat, long, true
}

func exifAltitude(x *goexif.Exif) (float64, bool) {
	tag, err := x.Get(goexif.GPSAltitude)
	if err != nil {
		return 0, false
	}
	value, ok := exifRat(tag)
	if !ok {
		return 0, false
	}
	if ref, ok := exifInt(x, goexif.GPSAltitudeRef); ok && ref == 1 {
		value = -value
	}
	return value, true
}

func exifRat(tag interface {
	Rat2(i int) (num, den int64, err error)
}) (float64, bool) {
	num, den, err := tag.Rat2(0)
	if err != nil || den == 0 {
		return 0, false
	}
	return float64(num) / float64(den), true
}

func formatExifFloat(value float64, precision int) string {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return ""
	}
	if precision < 0 {
		precision = 0
	}
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.*f", precision, value), "0"), ".")
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
