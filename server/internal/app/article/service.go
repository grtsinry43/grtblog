package article

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"time"

	"github.com/grtsinry43/grtblog-v2/server/internal/app/contentutil"
	appEvent "github.com/grtsinry43/grtblog-v2/server/internal/app/event"
	domaincomment "github.com/grtsinry43/grtblog-v2/server/internal/domain/comment"
	"github.com/grtsinry43/grtblog-v2/server/internal/domain/content"
)

type Service struct {
	repo        content.Repository
	commentRepo domaincomment.CommentRepository
	events      appEvent.Bus
}

func NewService(repo content.Repository, commentRepo domaincomment.CommentRepository, events appEvent.Bus) *Service {
	if events == nil {
		events = appEvent.NopBus{}
	}
	return &Service{repo: repo, commentRepo: commentRepo, events: events}
}

// CreateArticle 创建文章
func (s *Service) CreateArticle(ctx context.Context, authorID int64, cmd CreateArticleCmd) (*content.Article, error) {
	shortURL := ""
	if cmd.ShortURL != nil {
		shortURL = strings.TrimSpace(*cmd.ShortURL)
	}
	if shortURL == "" {
		shortURL = contentutil.GenerateShortURLFromTitle(cmd.Title)
	}
	shortURL, err := s.ensureShortURLAvailable(ctx, shortURL)
	if err != nil {
		return nil, err
	}

	if cmd.CategoryID != nil {
		if _, err := s.repo.GetCategoryByID(ctx, *cmd.CategoryID); err != nil {
			return nil, err
		}
	}
	if err := s.ensureTagsExist(ctx, cmd.TagIDs); err != nil {
		return nil, err
	}

	// 设置创建时间
	createdAt := time.Now()
	if cmd.CreatedAt != nil {
		createdAt = *cmd.CreatedAt
	}

	toc := contentutil.GenerateTOC(cmd.Content)
	summary := contentutil.BuildSummary(cmd.Summary, cmd.Content)

	article := &content.Article{
		Title:            cmd.Title,
		Summary:          summary,
		AISummary:        cmd.AISummary,
		LeadIn:           cmd.LeadIn,
		TOC:              toc,
		Content:          cmd.Content,
		ContentHash:      content.ArticleContentHash(cmd.Title, cmd.LeadIn, cmd.Content),
		AuthorID:         authorID,
		Cover:            cmd.Cover,
		CategoryID:       cmd.CategoryID,
		ShortURL:         shortURL,
		IsPublished:      cmd.IsPublished,
		IsTop:            cmd.IsTop,
		IsHot:            false,
		IsOriginal:       cmd.IsOriginal,
		ExtInfo:          mergeExtInfoKeepingFederation(nil, cmd.ExtInfo),
		ContentUpdatedAt: createdAt,
		CreatedAt:        createdAt,
	}
	if cmd.Views != nil && *cmd.Views > 0 {
		article.InitialViews = *cmd.Views
	}

	if err := s.repo.CreateArticle(ctx, article); err != nil {
		return nil, err
	}
	if err := s.applyCommentAreaStatus(ctx, article.CommentID, cmd.AllowComment); err != nil {
		return nil, err
	}

	// 如果有标签，则关联标签
	if len(cmd.TagIDs) > 0 {
		if err := s.repo.SyncTagsToArticle(ctx, article.ID, cmd.TagIDs); err != nil {
			return nil, err
		}
	}

	now := time.Now()
	_ = s.events.Publish(ctx, ArticleCreated{
		ID:        article.ID,
		AuthorID:  article.AuthorID,
		Title:     article.Title,
		ShortURL:  article.ShortURL,
		Published: article.IsPublished,
		At:        now,
	})
	if article.IsPublished {
		prevExtInfo := append([]byte(nil), article.ExtInfo...)
		_ = s.events.Publish(ctx, ArticlePublished{
			ID:       article.ID,
			AuthorID: article.AuthorID,
			Title:    article.Title,
			ShortURL: article.ShortURL,
			At:       now,
		})
		publishFederationSignals(ctx, s.events, article, cmd.Content)
		if !bytes.Equal(prevExtInfo, article.ExtInfo) {
			if err := s.repo.UpdateArticle(ctx, article); err != nil {
				return nil, err
			}
		}
	}

	return article, nil
}

// UpdateArticle 更新文章
func (s *Service) UpdateArticle(ctx context.Context, cmd UpdateArticleCmd) (*content.Article, error) {
	// 先获取现有文章
	existing, err := s.repo.GetArticleByID(ctx, cmd.ID)
	if err != nil {
		return nil, err
	}
	prevPublished := existing.IsPublished
	prevContentHash := existing.ContentHash

	if cmd.CategoryID != nil {
		if _, err := s.repo.GetCategoryByID(ctx, *cmd.CategoryID); err != nil {
			return nil, err
		}
	}
	if err := s.ensureTagsExist(ctx, cmd.TagIDs); err != nil {
		return nil, err
	}

	toc := contentutil.GenerateTOC(cmd.Content)
	summary := contentutil.BuildSummary(cmd.Summary, cmd.Content)

	// 更新字段
	existing.Title = cmd.Title
	existing.Summary = summary
	existing.AISummary = cmd.AISummary
	existing.LeadIn = cmd.LeadIn
	existing.TOC = toc
	existing.Content = cmd.Content
	existing.ContentHash = content.ArticleContentHash(cmd.Title, cmd.LeadIn, cmd.Content)
	if prevContentHash != existing.ContentHash {
		existing.ContentUpdatedAt = time.Now()
	}
	existing.Cover = cmd.Cover
	existing.CategoryID = cmd.CategoryID
	shortURL := strings.TrimSpace(cmd.ShortURL)
	if shortURL == "" {
		shortURL = existing.ShortURL
	}
	if shortURL != existing.ShortURL {
		shortURL, err = s.ensureShortURLAvailable(ctx, shortURL)
		if err != nil {
			return nil, err
		}
	}
	existing.ShortURL = shortURL
	existing.IsPublished = cmd.IsPublished
	existing.IsTop = cmd.IsTop
	existing.IsOriginal = cmd.IsOriginal
	existing.ExtInfo = mergeExtInfoKeepingFederation(existing.ExtInfo, cmd.ExtInfo)

	if err := s.repo.UpdateArticle(ctx, existing); err != nil {
		return nil, err
	}
	if err := s.applyCommentAreaStatus(ctx, existing.CommentID, cmd.AllowComment); err != nil {
		return nil, err
	}

	// 同步标签
	if err := s.repo.SyncTagsToArticle(ctx, existing.ID, cmd.TagIDs); err != nil {
		return nil, err
	}

	now := time.Now()
	_ = s.events.Publish(ctx, ArticleUpdated{
		ID:          existing.ID,
		AuthorID:    existing.AuthorID,
		Title:       existing.Title,
		ShortURL:    existing.ShortURL,
		Published:   existing.IsPublished,
		ContentHash: existing.ContentHash,
		LeadIn:      existing.LeadIn,
		TOC:         existing.TOC,
		Content:     existing.Content,
		At:          now,
	})
	if !prevPublished && existing.IsPublished {
		_ = s.events.Publish(ctx, ArticlePublished{
			ID:       existing.ID,
			AuthorID: existing.AuthorID,
			Title:    existing.Title,
			ShortURL: existing.ShortURL,
			At:       now,
		})
	}
	if prevPublished && !existing.IsPublished {
		_ = s.events.Publish(ctx, ArticleUnpublished{
			ID:       existing.ID,
			AuthorID: existing.AuthorID,
			Title:    existing.Title,
			ShortURL: existing.ShortURL,
			At:       now,
		})
	}
	if existing.IsPublished && (!prevPublished || prevContentHash != existing.ContentHash) {
		prevExtInfo := append([]byte(nil), existing.ExtInfo...)
		publishFederationSignals(ctx, s.events, existing, existing.Content)
		if !bytes.Equal(prevExtInfo, existing.ExtInfo) {
			if err := s.repo.UpdateArticle(ctx, existing); err != nil {
				return nil, err
			}
		}
	}

	return existing, nil
}

// GetArticleByID 根据 ID 获取文章
func (s *Service) GetArticleByID(ctx context.Context, id int64) (*content.Article, error) {
	return s.repo.GetArticleByID(ctx, id)
}

// GetArticleByShortURL 根据短链接获取文章
func (s *Service) GetArticleByShortURL(ctx context.Context, shortURL string) (*content.Article, error) {
	article, err := s.repo.GetArticleByShortURL(ctx, shortURL)
	if err != nil {
		return nil, err
	}

	return article, nil
}

// ListArticles 获取文章列表
func (s *Service) ListArticles(ctx context.Context, options content.ArticleListOptionsInternal) ([]*content.Article, int64, error) {
	return s.repo.ListArticles(ctx, options)
}

// ListPublicArticles 获取公开文章列表
func (s *Service) ListPublicArticles(ctx context.Context, options content.ArticleListOptions) ([]*content.Article, int64, error) {
	return s.repo.ListPublicArticles(ctx, options)
}

// BatchSetPublished 批量设置文章发布状态。
func (s *Service) BatchSetPublished(ctx context.Context, cmd BatchSetPublishedCmd) error {
	ids := normalizeIDs(cmd.IDs)
	for _, id := range ids {
		articleItem, err := s.repo.GetArticleByID(ctx, id)
		if err != nil {
			return err
		}
		prevPublished := articleItem.IsPublished
		if prevPublished == cmd.IsPublished {
			continue
		}
		articleItem.IsPublished = cmd.IsPublished
		if err := s.repo.UpdateArticle(ctx, articleItem); err != nil {
			return err
		}

		now := time.Now()
		_ = s.events.Publish(ctx, ArticleUpdated{
			ID:          articleItem.ID,
			AuthorID:    articleItem.AuthorID,
			Title:       articleItem.Title,
			ShortURL:    articleItem.ShortURL,
			Published:   articleItem.IsPublished,
			ContentHash: articleItem.ContentHash,
			LeadIn:      articleItem.LeadIn,
			TOC:         articleItem.TOC,
			Content:     articleItem.Content,
			At:          now,
		})
		if cmd.IsPublished {
			_ = s.events.Publish(ctx, ArticlePublished{
				ID:       articleItem.ID,
				AuthorID: articleItem.AuthorID,
				Title:    articleItem.Title,
				ShortURL: articleItem.ShortURL,
				At:       now,
			})
		} else {
			_ = s.events.Publish(ctx, ArticleUnpublished{
				ID:       articleItem.ID,
				AuthorID: articleItem.AuthorID,
				Title:    articleItem.Title,
				ShortURL: articleItem.ShortURL,
				At:       now,
			})
		}
	}
	return nil
}

// BatchSetTop 批量设置文章置顶状态。
func (s *Service) BatchSetTop(ctx context.Context, cmd BatchSetTopCmd) error {
	ids := normalizeIDs(cmd.IDs)
	for _, id := range ids {
		articleItem, err := s.repo.GetArticleByID(ctx, id)
		if err != nil {
			return err
		}
		if articleItem.IsTop == cmd.IsTop {
			continue
		}
		articleItem.IsTop = cmd.IsTop
		if err := s.repo.UpdateArticle(ctx, articleItem); err != nil {
			return err
		}
		_ = s.events.Publish(ctx, ArticleUpdated{
			ID:          articleItem.ID,
			AuthorID:    articleItem.AuthorID,
			Title:       articleItem.Title,
			ShortURL:    articleItem.ShortURL,
			Published:   articleItem.IsPublished,
			ContentHash: articleItem.ContentHash,
			LeadIn:      articleItem.LeadIn,
			TOC:         articleItem.TOC,
			Content:     articleItem.Content,
			At:          time.Now(),
		})
	}
	return nil
}

// DeleteArticle 删除文章
func (s *Service) DeleteArticle(ctx context.Context, id int64) error {
	article, err := s.repo.GetArticleByID(ctx, id)
	if err != nil {
		return err
	}
	if err := s.repo.DeleteArticle(ctx, id); err != nil {
		return err
	}
	_ = s.events.Publish(ctx, ArticleDeleted{
		ID:       article.ID,
		AuthorID: article.AuthorID,
		Title:    article.Title,
		ShortURL: article.ShortURL,
		At:       time.Now(),
	})
	return nil
}

// BatchDelete 批量删除文章。
func (s *Service) BatchDelete(ctx context.Context, cmd BatchDeleteCmd) error {
	ids := normalizeIDs(cmd.IDs)
	for _, id := range ids {
		if err := s.DeleteArticle(ctx, id); err != nil {
			return err
		}
	}
	return nil
}

// GetArticleWithTags 获取文章及其标签
func (s *Service) GetArticleWithTags(ctx context.Context, id int64) (*content.Article, []*content.Tag, error) {
	article, err := s.repo.GetArticleByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	tags, err := s.repo.GetTagsByArticleID(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	return article, tags, nil
}

// GetArticleMetrics 获取文章指标
func (s *Service) GetArticleMetrics(ctx context.Context, articleID int64) (*content.ArticleMetrics, error) {
	return s.repo.GetArticleMetrics(ctx, articleID)
}

// GetArticleTags 获取文章标签。
func (s *Service) GetArticleTags(ctx context.Context, articleID int64) ([]*content.Tag, error) {
	return s.repo.GetTagsByArticleID(ctx, articleID)
}

// UpdateHotArticles 根据指标更新热门文章状态
func (s *Service) UpdateHotArticles(ctx context.Context, viewThreshold, likeThreshold, commentThreshold int64) error {
	marked, err := s.repo.SyncHotArticles(ctx, viewThreshold, likeThreshold, commentThreshold)
	if err != nil {
		return err
	}

	now := time.Now()
	for _, item := range marked {
		if !item.IsPublished || strings.TrimSpace(item.ShortURL) == "" {
			continue
		}
		_ = s.events.Publish(ctx, ArticleHotMarked{
			ID:       item.ID,
			Title:    item.Title,
			ShortURL: item.ShortURL,
			At:       now,
		})
	}
	return nil
}

// generateShortURL 生成短链接
func (s *Service) ensureShortURLAvailable(ctx context.Context, shortURL string) (string, error) {
	shortURL = strings.TrimSpace(shortURL)
	if shortURL == "" {
		for i := 0; i < 5; i++ {
			candidate := contentutil.GenerateRandomShortURL()
			_, err := s.repo.GetArticleByShortURL(ctx, candidate)
			if err != nil {
				if errors.Is(err, content.ErrArticleNotFound) {
					return candidate, nil
				}
				return "", err
			}
		}
		return "", content.ErrArticleShortURLExists
	}

	existing, err := s.repo.GetArticleByShortURL(ctx, shortURL)
	if err != nil && !errors.Is(err, content.ErrArticleNotFound) {
		return "", err
	}
	if err == nil && existing != nil {
		return "", content.ErrArticleShortURLExists
	}
	return shortURL, nil
}

func (s *Service) ensureTagsExist(ctx context.Context, tagIDs []int64) error {
	if len(tagIDs) == 0 {
		return nil
	}
	unique := make(map[int64]struct{}, len(tagIDs))
	for _, id := range tagIDs {
		if id <= 0 {
			return content.ErrTagNotFound
		}
		unique[id] = struct{}{}
	}
	ids := make([]int64, 0, len(unique))
	for id := range unique {
		ids = append(ids, id)
	}
	ok, err := s.repo.TagIDsExist(ctx, ids)
	if err != nil {
		return err
	}
	if !ok {
		return content.ErrTagNotFound
	}
	return nil
}

func (s *Service) applyCommentAreaStatus(ctx context.Context, areaID *int64, allowComment *bool) error {
	if s.commentRepo == nil || areaID == nil || *areaID <= 0 || allowComment == nil {
		return nil
	}
	return s.commentRepo.SetAreaClosed(ctx, *areaID, !*allowComment)
}

func normalizeIDs(ids []int64) []int64 {
	if len(ids) == 0 {
		return nil
	}
	seen := make(map[int64]struct{}, len(ids))
	out := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}
