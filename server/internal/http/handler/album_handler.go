package handler

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"

	appalbum "github.com/grtsinry43/grtblog-v2/server/internal/app/album"
	mediaapp "github.com/grtsinry43/grtblog-v2/server/internal/app/media"
	domainalbum "github.com/grtsinry43/grtblog-v2/server/internal/domain/album"
	domaincomment "github.com/grtsinry43/grtblog-v2/server/internal/domain/comment"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/contract"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/middleware"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/response"
)

type AlbumHandler struct {
	svc         *appalbum.Service
	albumRepo   domainalbum.Repository
	commentRepo domaincomment.CommentRepository
	mediaSvc    *mediaapp.Service
}

func NewAlbumHandler(svc *appalbum.Service, albumRepo domainalbum.Repository, commentRepo domaincomment.CommentRepository, mediaSvc *mediaapp.Service) *AlbumHandler {
	return &AlbumHandler{svc: svc, albumRepo: albumRepo, commentRepo: commentRepo, mediaSvc: mediaSvc}
}

// --------------- Album CRUD ---------------

func (h *AlbumHandler) CreateAlbum(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.ErrorFromBiz[any](c, response.NotLogin)
	}

	var req contract.CreateAlbumReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}

	cmd := appalbum.CreateAlbumCmd{
		Title:        req.Title,
		Description:  req.Description,
		Cover:        req.Cover,
		ShortURL:     req.ShortURL,
		IsPublished:  req.IsPublished,
		AllowComment: req.AllowComment,
		CreatedAt:    req.CreatedAt,
	}
	if cmd.AllowComment == nil {
		defaultAllow := true
		cmd.AllowComment = &defaultAllow
	}

	created, err := h.svc.CreateAlbum(c.Context(), claims.UserID, cmd)
	if err != nil {
		if errors.Is(err, domainalbum.ErrAlbumShortURLExists) {
			return response.NewBizErrorWithMsg(response.ParamsError, "短链接已存在")
		}
		return err
	}

	resp, err := h.toAlbumResp(c.Context(), created)
	if err != nil {
		return err
	}

	Audit(c, "album.create", map[string]any{
		"albumId": created.ID,
		"title":   created.Title,
		"userId":  claims.UserID,
	})

	return response.SuccessWithMessage(c, resp, "相册创建成功")
}

func (h *AlbumHandler) UpdateAlbum(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.ErrorFromBiz[any](c, response.NotLogin)
	}

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的相册ID")
	}

	var req contract.UpdateAlbumReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}

	cmd := appalbum.UpdateAlbumCmd{
		ID:           id,
		Title:        req.Title,
		Description:  req.Description,
		Cover:        req.Cover,
		ShortURL:     req.ShortURL,
		IsPublished:  req.IsPublished,
		AllowComment: req.AllowComment,
	}

	updated, err := h.svc.UpdateAlbum(c.Context(), cmd)
	if err != nil {
		if errors.Is(err, domainalbum.ErrAlbumShortURLExists) {
			return response.NewBizErrorWithMsg(response.ParamsError, "短链接已存在")
		}
		if errors.Is(err, domainalbum.ErrAlbumNotFound) {
			return response.NewBizErrorWithMsg(response.NotFound, "相册不存在")
		}
		return err
	}

	resp, err := h.toAlbumResp(c.Context(), updated)
	if err != nil {
		return err
	}

	Audit(c, "album.update", map[string]any{
		"albumId": updated.ID,
		"title":   updated.Title,
		"userId":  claims.UserID,
	})

	return response.SuccessWithMessage(c, resp, "相册更新成功")
}

func (h *AlbumHandler) DeleteAlbum(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的相册ID")
	}

	if err := h.svc.DeleteAlbum(c.Context(), id); err != nil {
		if errors.Is(err, domainalbum.ErrAlbumNotFound) {
			return response.NewBizErrorWithMsg(response.NotFound, "相册不存在")
		}
		return err
	}

	Audit(c, "album.delete", map[string]any{"albumId": id})

	return response.SuccessWithMessage[any](c, nil, "相册删除成功")
}

func (h *AlbumHandler) GetAlbum(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的相册ID")
	}

	a, err := h.svc.GetAlbumByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, domainalbum.ErrAlbumNotFound) {
			return response.NewBizErrorWithMsg(response.NotFound, "相册不存在")
		}
		return err
	}
	if !a.IsPublished {
		return response.NewBizErrorWithMsg(response.NotFound, "相册不存在")
	}

	resp, err := h.toAlbumResp(c.Context(), a)
	if err != nil {
		return err
	}
	return response.Success(c, resp)
}

func (h *AlbumHandler) GetAlbumByShortURL(c *fiber.Ctx) error {
	shortURL := c.Params("shortUrl")
	if shortURL == "" {
		return response.NewBizErrorWithMsg(response.ParamsError, "短链接不能为空")
	}

	a, err := h.svc.GetAlbumByShortURL(c.Context(), shortURL)
	if err != nil {
		if errors.Is(err, domainalbum.ErrAlbumNotFound) {
			return response.NewBizErrorWithMsg(response.NotFound, "相册不存在")
		}
		return err
	}
	if !a.IsPublished {
		return response.NewBizErrorWithMsg(response.NotFound, "相册不存在")
	}

	photos, err := h.svc.ListAlbumPhotos(c.Context(), a.ID)
	if err != nil {
		return err
	}
	metrics, _ := h.svc.GetAlbumMetrics(c.Context(), a.ID)

	detail := contract.AlbumDetailResp{
		AlbumResp: h.buildAlbumResp(a, int64(len(photos)), metrics),
		Photos:    h.mapPhotosResp(photos),
	}

	return response.Success(c, detail)
}

func (h *AlbumHandler) ListAlbums(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var search *string
	if s := c.Query("search"); s != "" {
		search = &s
	}

	albums, total, err := h.svc.ListPublicAlbums(c.Context(), domainalbum.AlbumListOptions{
		Page: page, PageSize: pageSize, Search: search,
	})
	if err != nil {
		return err
	}

	items := make([]contract.AlbumListItemResp, len(albums))
	for i, a := range albums {
		count, _ := h.svc.CountAlbumPhotos(c.Context(), a.ID)
		metrics, _ := h.svc.GetAlbumMetrics(c.Context(), a.ID)
		items[i] = h.buildAlbumListItemResp(a, count, metrics)
	}

	return response.Success(c, contract.AlbumListResp{
		Items: items, Total: total, Page: page, Size: pageSize,
	})
}

// --------------- Admin ---------------

func (h *AlbumHandler) GetAlbumAdmin(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的相册ID")
	}

	a, err := h.svc.GetAlbumByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, domainalbum.ErrAlbumNotFound) {
			return response.NewBizErrorWithMsg(response.NotFound, "相册不存在")
		}
		return err
	}

	photos, err := h.svc.ListAlbumPhotos(c.Context(), a.ID)
	if err != nil {
		return err
	}
	metrics, _ := h.svc.GetAlbumMetrics(c.Context(), a.ID)

	detail := contract.AlbumDetailResp{
		AlbumResp: h.buildAlbumResp(a, int64(len(photos)), metrics),
		Photos:    h.mapPhotosResp(photos),
	}

	return response.Success(c, detail)
}

func (h *AlbumHandler) ListAlbumsAdmin(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var search *string
	if s := c.Query("search"); s != "" {
		search = &s
	}
	var published *bool
	if p := c.Query("published"); p != "" {
		val := p == "true"
		published = &val
	}

	albums, total, err := h.svc.ListAlbums(c.Context(), domainalbum.AlbumListOptionsInternal{
		Page: page, PageSize: pageSize, Published: published, Search: search,
	})
	if err != nil {
		return err
	}

	items := make([]contract.AlbumListItemResp, len(albums))
	for i, a := range albums {
		count, _ := h.svc.CountAlbumPhotos(c.Context(), a.ID)
		metrics, _ := h.svc.GetAlbumMetrics(c.Context(), a.ID)
		items[i] = h.buildAlbumListItemResp(a, count, metrics)
	}

	return response.Success(c, contract.AlbumListResp{
		Items: items, Total: total, Page: page, Size: pageSize,
	})
}

func (h *AlbumHandler) BatchSetAlbumPublished(c *fiber.Ctx) error {
	var req contract.BatchSetAlbumPublishedReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}
	if len(req.IDs) == 0 {
		return response.NewBizErrorWithMsg(response.ParamsError, "ids 不能为空")
	}

	if err := h.svc.BatchSetPublished(c.Context(), appalbum.BatchSetPublishedCmd{
		IDs: req.IDs, IsPublished: req.IsPublished,
	}); err != nil {
		return err
	}

	if req.IsPublished {
		return response.SuccessWithMessage[any](c, nil, "相册已批量发布")
	}
	return response.SuccessWithMessage[any](c, nil, "相册已批量取消发布")
}

func (h *AlbumHandler) BatchDeleteAlbums(c *fiber.Ctx) error {
	var req contract.BatchDeleteAlbumReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}
	if len(req.IDs) == 0 {
		return response.NewBizErrorWithMsg(response.ParamsError, "ids 不能为空")
	}

	if err := h.svc.BatchDelete(c.Context(), appalbum.BatchDeleteCmd{IDs: req.IDs}); err != nil {
		return err
	}

	return response.SuccessWithMessage[any](c, nil, "相册已批量删除")
}

// --------------- Photo sub-resource ---------------

func (h *AlbumHandler) AddPhotos(c *fiber.Ctx) error {
	albumID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的相册ID")
	}

	var req contract.BatchCreatePhotosReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}
	if len(req.Photos) == 0 {
		return response.NewBizErrorWithMsg(response.ParamsError, "photos 不能为空")
	}

	cmds := make([]appalbum.CreatePhotoCmd, len(req.Photos))
	for i, p := range req.Photos {
		exifMap := map[string]any{}
		if p.Exif != nil {
			_ = json.Unmarshal([]byte(*p.Exif), &exifMap)
		}

		// If missing dimensions or dominantColor, extract from local file
		_, hasWidth := exifMap["imageWidth"]
		_, hasColor := exifMap["dominantColor"]
		if (!hasWidth || !hasColor) && h.mediaSvc != nil {
			_, meta := h.mediaSvc.ExtractImageMetaFromURL(p.URL)
			if meta != nil {
				if !hasWidth && meta.Width > 0 {
					exifMap["imageWidth"] = meta.Width
					exifMap["imageHeight"] = meta.Height
				}
				if !hasColor && meta.DominantColor != "" {
					exifMap["dominantColor"] = meta.DominantColor
				}
			}
		}

		var exifBytes []byte
		if len(exifMap) > 0 {
			exifBytes, _ = json.Marshal(exifMap)
		}

		cmds[i] = appalbum.CreatePhotoCmd{
			URL:         p.URL,
			Description: p.Description,
			Caption:     p.Caption,
			Exif:        exifBytes,
			SortOrder:   p.SortOrder,
		}
	}

	photos, err := h.svc.AddPhotos(c.Context(), appalbum.BatchCreatePhotosCmd{
		AlbumID: albumID, Photos: cmds,
	})
	if err != nil {
		if errors.Is(err, domainalbum.ErrAlbumNotFound) {
			return response.NewBizErrorWithMsg(response.NotFound, "相册不存在")
		}
		return err
	}

	respPhotos := make([]contract.PhotoResp, len(photos))
	for i, p := range photos {
		respPhotos[i] = h.mapPhotoResp(p)
	}

	Audit(c, "album.photos.add", map[string]any{
		"albumId": albumID,
		"count":   len(photos),
	})

	return response.SuccessWithMessage(c, respPhotos, "照片添加成功")
}

func (h *AlbumHandler) UpdatePhoto(c *fiber.Ctx) error {
	photoID, err := strconv.ParseInt(c.Params("photoId"), 10, 64)
	if err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的照片ID")
	}

	var req contract.UpdatePhotoReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}

	var exifBytes []byte
	if req.Exif != nil {
		exifBytes = []byte(*req.Exif)
	}
	updated, err := h.svc.UpdatePhoto(c.Context(), appalbum.UpdatePhotoCmd{
		ID:          photoID,
		URL:         req.URL,
		Description: req.Description,
		Caption:     req.Caption,
		Exif:        exifBytes,
		SortOrder:   req.SortOrder,
	})
	if err != nil {
		if errors.Is(err, domainalbum.ErrPhotoNotFound) {
			return response.NewBizErrorWithMsg(response.NotFound, "照片不存在")
		}
		return err
	}

	return response.SuccessWithMessage(c, h.mapPhotoResp(updated), "照片更新成功")
}

func (h *AlbumHandler) DeletePhoto(c *fiber.Ctx) error {
	photoID, err := strconv.ParseInt(c.Params("photoId"), 10, 64)
	if err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的照片ID")
	}

	if err := h.svc.DeletePhoto(c.Context(), photoID); err != nil {
		if errors.Is(err, domainalbum.ErrPhotoNotFound) {
			return response.NewBizErrorWithMsg(response.NotFound, "照片不存在")
		}
		return err
	}

	return response.SuccessWithMessage[any](c, nil, "照片删除成功")
}

func (h *AlbumHandler) ReorderPhotos(c *fiber.Ctx) error {
	albumID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的相册ID")
	}

	var req contract.ReorderPhotosReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}
	if len(req.PhotoIDs) == 0 {
		return response.NewBizErrorWithMsg(response.ParamsError, "photoIds 不能为空")
	}

	if err := h.svc.ReorderPhotos(c.Context(), appalbum.ReorderPhotosCmd{
		AlbumID: albumID, PhotoIDs: req.PhotoIDs,
	}); err != nil {
		if errors.Is(err, domainalbum.ErrAlbumNotFound) {
			return response.NewBizErrorWithMsg(response.NotFound, "相册不存在")
		}
		return err
	}

	return response.SuccessWithMessage[any](c, nil, "照片排序更新成功")
}

// --------------- Response mappers ---------------

func (h *AlbumHandler) toAlbumResp(ctx context.Context, a *domainalbum.Album) (*contract.AlbumResp, error) {
	count, _ := h.svc.CountAlbumPhotos(ctx, a.ID)
	metrics, _ := h.svc.GetAlbumMetrics(ctx, a.ID)
	resp := h.buildAlbumResp(a, count, metrics)
	return &resp, nil
}

func (h *AlbumHandler) buildAlbumResp(a *domainalbum.Album, photoCount int64, metrics *domainalbum.AlbumMetrics) contract.AlbumResp {
	resp := contract.AlbumResp{
		ID:          a.ID,
		Title:       a.Title,
		Description: a.Description,
		Cover:       a.Cover,
		ShortURL:    a.ShortURL,
		AuthorID:    a.AuthorID,
		CommentID:   a.CommentID,
		IsPublished: a.IsPublished,
		PhotoCount:  photoCount,
		CreatedAt:   a.CreatedAt,
		UpdatedAt:   a.UpdatedAt,
	}

	resp.AllowComment = true
	if a.CommentID != nil && h.commentRepo != nil {
		area, err := h.commentRepo.GetAreaByID(context.Background(), *a.CommentID)
		if err == nil {
			resp.AllowComment = !area.IsClosed
		}
	}

	if metrics != nil {
		resp.Metrics = &contract.MetricsResp{
			Views:    metrics.Views,
			Likes:    metrics.Likes,
			Comments: metrics.Comments,
		}
	}

	return resp
}

func (h *AlbumHandler) buildAlbumListItemResp(a *domainalbum.Album, photoCount int64, metrics *domainalbum.AlbumMetrics) contract.AlbumListItemResp {
	item := contract.AlbumListItemResp{
		ID:          a.ID,
		Title:       a.Title,
		Description: a.Description,
		Cover:       a.Cover,
		ShortURL:    a.ShortURL,
		IsPublished: a.IsPublished,
		PhotoCount:  photoCount,
		CreatedAt:   a.CreatedAt,
		UpdatedAt:   a.UpdatedAt,
	}
	if metrics != nil {
		item.Views = metrics.Views
		item.Likes = metrics.Likes
		item.Comments = metrics.Comments
	}
	return item
}

func (h *AlbumHandler) mapPhotosResp(photos []*domainalbum.Photo) []contract.PhotoResp {
	resp := make([]contract.PhotoResp, len(photos))
	for i, p := range photos {
		resp[i] = h.mapPhotoResp(p)
	}
	return resp
}

func (h *AlbumHandler) mapPhotoResp(p *domainalbum.Photo) contract.PhotoResp {
	resp := contract.PhotoResp{
		ID:          p.ID,
		AlbumID:     p.AlbumID,
		URL:         p.URL,
		Description: p.Description,
		Caption:     p.Caption,
		SortOrder:   p.SortOrder,
		CreatedAt:   p.CreatedAt,
	}
	if len(p.Exif) > 0 {
		raw := json.RawMessage(p.Exif)
		resp.Exif = &raw
	}
	if h.mediaSvc != nil {
		resp.ThumbnailURL = h.mediaSvc.ThumbnailURLFor(p.URL)
	}
	return resp
}
