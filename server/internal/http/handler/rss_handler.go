package handler

import (
	"encoding/xml"
	"strings"

	"github.com/aclr/feeds"
	"github.com/gofiber/fiber/v2"
	apprss "github.com/grtsinry43/grtblog-v2/server/internal/app/rss"
)

type RSSHandler struct {
	svc *apprss.Service
}

func NewRSSHandler(svc *apprss.Service) *RSSHandler {
	return &RSSHandler{svc: svc}
}

// GetFeed godoc
// @Summary 获取 RSS 聚合订阅
// @Tags Public
// @Produce application/rss+xml
// @Param limit query int false "条目数量" default(20)
// @Success 200 {string} string "rss xml"
// @Router /public/rss.xml [get]
func (h *RSSHandler) GetFeed(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 20)
	aggFeed, err := h.svc.Build(c.Context(), c.BaseURL(), limit)
	if err != nil {
		return err
	}

	rssFeed := &feeds.Feed{
		Title:       aggFeed.Title,
		Link:        &feeds.Link{Href: aggFeed.Link},
		Description: aggFeed.Description,
		Created:     aggFeed.LastBuildDate,
		Updated:     aggFeed.LastBuildDate,
		Items:       make([]*feeds.Item, 0, len(aggFeed.Items)),
	}
	if aggFeed.AuthorName != "" || aggFeed.AuthorEmail != "" {
		rssFeed.Author = &feeds.Author{
			Name:  aggFeed.AuthorName,
			Email: aggFeed.AuthorEmail,
		}
	}
	if aggFeed.ImageURL != "" {
		rssFeed.Image = &feeds.Image{
			Url:   aggFeed.ImageURL,
			Title: aggFeed.Title,
			Link:  aggFeed.Link,
		}
	}

	for _, item := range aggFeed.Items {
		var author *feeds.Author
		if item.AuthorName != "" || item.AuthorEmail != "" {
			author = &feeds.Author{
				Name:  item.AuthorName,
				Email: item.AuthorEmail,
			}
		}
		rssFeed.Items = append(rssFeed.Items, &feeds.Item{
			Title:       item.Title,
			Link:        &feeds.Link{Href: item.Link},
			Id:          item.GUID,
			Author:      author,
			Created:     item.PublishedAt,
			Updated:     item.PublishedAt,
			Description: item.Description,
		})
	}

	channel := (&feeds.Rss{Feed: rssFeed}).RssFeed()
	channel.Language = "zh-CN"
	channel.Generator = "grtblog-v2/server"
	output, err := feeds.ToXML(channel)
	if err != nil {
		return err
	}
	output = appendRSSExtensions(output, c, aggFeed.FollowFeedID, aggFeed.FollowUserID)

	c.Set(fiber.HeaderContentType, "application/xml; charset=utf-8")
	return c.SendString(output)
}

func appendRSSExtensions(raw string, c *fiber.Ctx, feedID string, userID string) string {
	result := raw
	if strings.Contains(result, "<rss ") && !strings.Contains(result, "xmlns:atom=") {
		result = strings.Replace(
			result,
			"<rss ",
			"<rss xmlns:atom=\"http://www.w3.org/2005/Atom\" ",
			1,
		)
	}

	selfLink := strings.TrimSpace(c.BaseURL() + c.OriginalURL())
	extra := strings.Builder{}
	if selfLink != "" {
		extra.WriteString(`<atom:link href="`)
		extra.WriteString(xmlEscape(selfLink))
		extra.WriteString(`" rel="self" type="application/rss+xml"/>`)
	}
	if strings.TrimSpace(feedID) != "" || strings.TrimSpace(userID) != "" {
		extra.WriteString("<follow_challenge>")
		extra.WriteString("<feedId>")
		extra.WriteString(xmlEscape(strings.TrimSpace(feedID)))
		extra.WriteString("</feedId>")
		extra.WriteString("<userId>")
		extra.WriteString(xmlEscape(strings.TrimSpace(userID)))
		extra.WriteString("</userId>")
		extra.WriteString("</follow_challenge>")
	}
	if extra.Len() == 0 {
		return result
	}
	return strings.Replace(result, "</channel>", extra.String()+"</channel>", 1)
}

func xmlEscape(v string) string {
	var b strings.Builder
	if err := xml.EscapeText(&b, []byte(v)); err != nil {
		return v
	}
	return b.String()
}
