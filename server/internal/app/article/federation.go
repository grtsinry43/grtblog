package article

import (
	"context"
	"log"
	"time"

	appEvent "github.com/grtsinry43/grtblog-v2/server/internal/app/event"
	appfed "github.com/grtsinry43/grtblog-v2/server/internal/app/federation"
	"github.com/grtsinry43/grtblog-v2/server/internal/domain/content"
)

func publishFederationSignals(ctx context.Context, bus appEvent.Bus, article *content.Article, contentBody string) {
	if bus == nil || article == nil {
		return
	}
	mentions, citations := appfed.ParseSignals(contentBody)
	if len(mentions) == 0 && len(citations) == 0 {
		return
	}
	deliveredMentions, deliveredCitations := deliveredSignalKeys(article.ExtInfo)
	now := time.Now()
	newMentionKeys := make([]string, 0, len(mentions))
	newCitationKeys := make([]string, 0, len(citations))
	for _, mention := range mentions {
		key := mention.User + "@" + mention.Instance
		if _, exists := deliveredMentions[key]; exists {
			continue
		}
		// Only record the signal as delivered when the dispatch handler
		// succeeded; otherwise the marker would be silently lost forever.
		if err := bus.Publish(ctx, appfed.MentionDetected{
			ArticleID:      article.ID,
			AuthorID:       article.AuthorID,
			Title:          article.Title,
			ShortURL:       article.ShortURL,
			TargetUser:     mention.User,
			TargetInstance: mention.Instance,
			Context:        mention.Context,
			MentionType:    "",
			At:             now,
		}); err != nil {
			log.Printf("[federation] 提及事件发布失败 article_id=%d target=%s@%s err=%v", article.ID, mention.User, mention.Instance, err)
			continue
		}
		newMentionKeys = append(newMentionKeys, key)
	}
	for _, citation := range citations {
		key := citation.Instance + "|" + citation.PostID
		if _, exists := deliveredCitations[key]; exists {
			continue
		}
		if err := bus.Publish(ctx, appfed.CitationDetected{
			ArticleID:      article.ID,
			AuthorID:       article.AuthorID,
			Title:          article.Title,
			ShortURL:       article.ShortURL,
			TargetInstance: citation.Instance,
			TargetPostID:   citation.PostID,
			Context:        citation.Context,
			CitationType:   "",
			At:             now,
		}); err != nil {
			log.Printf("[federation] 引用事件发布失败 article_id=%d target=%s|%s err=%v", article.ID, citation.Instance, citation.PostID, err)
			continue
		}
		newCitationKeys = append(newCitationKeys, key)
	}
	if len(newMentionKeys) == 0 && len(newCitationKeys) == 0 {
		return
	}
	if updated, changed := markDeliveredSignals(article.ExtInfo, newMentionKeys, newCitationKeys); changed {
		article.ExtInfo = updated
	}
}
