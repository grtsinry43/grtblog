package media

import (
	"context"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	domainmedia "github.com/grtsinry43/grtblog-v2/server/internal/domain/media"
)

func TestSyncIndexCreatesAndDeletesRecords(t *testing.T) {
	t.Parallel()

	uploadDir := t.TempDir()
	writePNG(t, filepath.Join(uploadDir, "pictures", "sample.png"))
	writeText(t, filepath.Join(uploadDir, "files", "readme.txt"), "hello")
	writeText(t, filepath.Join(uploadDir, "thumbnails", "pictures", "ignored.txt"), "thumb")

	repo := newMemoryRepo()
	repo.mustSeed(domainmedia.UploadFile{
		ID:   99,
		Name: "stale.txt",
		Path: "/files/stale.txt",
		Type: "file",
		Size: 1,
		Hash: "stale-hash",
	})

	svc := NewService(repo, uploadDir, nil)
	result, err := svc.SyncIndex(context.Background())
	if err != nil {
		t.Fatalf("SyncIndex() error = %v", err)
	}

	if result.Scanned != 2 {
		t.Fatalf("SyncIndex() scanned = %d, want 2", result.Scanned)
	}
	if result.Indexed != 2 {
		t.Fatalf("SyncIndex() indexed = %d, want 2", result.Indexed)
	}
	if result.Created != 2 {
		t.Fatalf("SyncIndex() created = %d, want 2", result.Created)
	}
	if result.Deleted != 1 {
		t.Fatalf("SyncIndex() deleted = %d, want 1", result.Deleted)
	}

	files, err := repo.ListAll(context.Background())
	if err != nil {
		t.Fatalf("ListAll() error = %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("ListAll() len = %d, want 2", len(files))
	}

	byPath := make(map[string]domainmedia.UploadFile, len(files))
	for _, file := range files {
		byPath[file.Path] = file
	}

	picture, ok := byPath["/pictures/sample.png"]
	if !ok {
		t.Fatalf("missing synced picture record")
	}
	if picture.Type != "picture" {
		t.Fatalf("picture.Type = %q, want picture", picture.Type)
	}

	text, ok := byPath["/files/readme.txt"]
	if !ok {
		t.Fatalf("missing synced file record")
	}
	if text.Type != "file" {
		t.Fatalf("text.Type = %q, want file", text.Type)
	}
}

func TestSyncIndexUpdatesExistingRecordMetadata(t *testing.T) {
	t.Parallel()

	uploadDir := t.TempDir()
	imagePath := filepath.Join(uploadDir, "files", "legacy-photo.jpg")
	writePNG(t, imagePath)

	repo := newMemoryRepo()
	repo.mustSeed(domainmedia.UploadFile{
		ID:   7,
		Name: "legacy.bin",
		Path: "/files/legacy-photo.jpg",
		Type: "file",
		Size: 1,
		Hash: "",
	})

	svc := NewService(repo, uploadDir, nil)
	result, err := svc.SyncIndex(context.Background())
	if err != nil {
		t.Fatalf("SyncIndex() error = %v", err)
	}

	if result.Created != 0 {
		t.Fatalf("SyncIndex() created = %d, want 0", result.Created)
	}
	if result.Updated != 1 {
		t.Fatalf("SyncIndex() updated = %d, want 1", result.Updated)
	}

	files, err := repo.ListAll(context.Background())
	if err != nil {
		t.Fatalf("ListAll() error = %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("ListAll() len = %d, want 1", len(files))
	}

	file := files[0]
	if file.Name != "legacy-photo.jpg" {
		t.Fatalf("file.Name = %q, want legacy-photo.jpg", file.Name)
	}
	if file.Type != "picture" {
		t.Fatalf("file.Type = %q, want picture", file.Type)
	}
	if file.Hash == "" {
		t.Fatalf("file.Hash should be populated")
	}
	if file.Size <= 1 {
		t.Fatalf("file.Size = %d, want > 1", file.Size)
	}
}

type memoryRepo struct {
	nextID int64
	files  []domainmedia.UploadFile
}

func newMemoryRepo() *memoryRepo {
	return &memoryRepo{nextID: 1}
}

func (r *memoryRepo) FindByHash(_ context.Context, hash string) (*domainmedia.UploadFile, error) {
	for i := range r.files {
		if r.files[i].Hash == hash {
			file := r.files[i]
			return &file, nil
		}
	}
	return nil, domainmedia.ErrUploadFileNotFound
}

func (r *memoryRepo) FindByID(_ context.Context, id int64) (*domainmedia.UploadFile, error) {
	for i := range r.files {
		if r.files[i].ID == id {
			file := r.files[i]
			return &file, nil
		}
	}
	return nil, domainmedia.ErrUploadFileNotFound
}

func (r *memoryRepo) Create(_ context.Context, file *domainmedia.UploadFile) error {
	if file.ID == 0 {
		file.ID = r.nextID
		r.nextID++
	}
	r.files = append(r.files, *file)
	return nil
}

func (r *memoryRepo) Update(_ context.Context, file *domainmedia.UploadFile) error {
	for i := range r.files {
		if r.files[i].ID == file.ID {
			r.files[i] = *file
			return nil
		}
	}
	return domainmedia.ErrUploadFileNotFound
}

func (r *memoryRepo) UpdatePath(_ context.Context, id int64, path string) error {
	for i := range r.files {
		if r.files[i].ID == id {
			r.files[i].Path = path
			return nil
		}
	}
	return domainmedia.ErrUploadFileNotFound
}

func (r *memoryRepo) UpdateName(_ context.Context, id int64, name string) error {
	for i := range r.files {
		if r.files[i].ID == id {
			r.files[i].Name = name
			return nil
		}
	}
	return domainmedia.ErrUploadFileNotFound
}

func (r *memoryRepo) List(_ context.Context, offset int, limit int) ([]domainmedia.UploadFile, int64, error) {
	if offset > len(r.files) {
		return nil, int64(len(r.files)), nil
	}
	end := offset + limit
	if end > len(r.files) {
		end = len(r.files)
	}
	items := append([]domainmedia.UploadFile(nil), r.files[offset:end]...)
	return items, int64(len(r.files)), nil
}

func (r *memoryRepo) ListAll(_ context.Context) ([]domainmedia.UploadFile, error) {
	return append([]domainmedia.UploadFile(nil), r.files...), nil
}

func (r *memoryRepo) DeleteByID(_ context.Context, id int64) error {
	for i := range r.files {
		if r.files[i].ID == id {
			r.files = append(r.files[:i], r.files[i+1:]...)
			return nil
		}
	}
	return domainmedia.ErrUploadFileNotFound
}

func (r *memoryRepo) mustSeed(file domainmedia.UploadFile) {
	r.files = append(r.files, file)
	if file.ID >= r.nextID {
		r.nextID = file.ID + 1
	}
}

func writePNG(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", path, err)
	}

	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("Create(%q) error = %v", path, err)
	}
	defer f.Close()

	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			img.Set(x, y, color.RGBA{R: 120, G: 80, B: 40, A: 255})
		}
	}

	if err := png.Encode(f, img); err != nil {
		t.Fatalf("png.Encode(%q) error = %v", path, err)
	}
}

func writeText(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
}
