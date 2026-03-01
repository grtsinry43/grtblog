package article

import (
	"encoding/json"
	"fmt"
	"html"
	"strings"

	appfed "github.com/grtsinry43/grtblog-v2/server/internal/app/federation"
	"github.com/grtsinry43/grtblog-v2/server/internal/domain/federation"
)

// ExpandFederationSignals replaces federation signal markers in article content
// with enriched component blocks / HTML tags for public rendering.
//
//   - <cite:instance|post-id> → ::: fed-citation instance="..." post-id="..." ... :::
//   - <@user@instance>        → <fed-mention data-user="..." data-instance="..." data-status="...">@user@instance</fed-mention>
func ExpandFederationSignals(
	content string,
	deliveries []federation.OutboundDelivery,
	postCache []federation.FederatedPostCache,
	instances []federation.FederationInstance,
) string {
	if strings.TrimSpace(content) == "" {
		return content
	}

	// Build lookup maps.
	// Delivery keyed by "type|targetInstance|identifier" for quick lookup.
	deliveryByKey := make(map[string]*federation.OutboundDelivery, len(deliveries))
	for i := range deliveries {
		d := &deliveries[i]
		deliveryByKey[deliveryKey(d)] = d
	}

	// Post cache keyed by "instanceBaseURL|remotePostID".
	postByKey := make(map[string]*federation.FederatedPostCache, len(postCache))
	instanceByID := make(map[int64]*federation.FederationInstance, len(instances))
	for i := range instances {
		instanceByID[instances[i].ID] = &instances[i]
	}
	for i := range postCache {
		p := &postCache[i]
		inst := instanceByID[p.InstanceID]
		if inst == nil || p.RemotePostID == nil {
			continue
		}
		baseHost := extractHost(inst.BaseURL)
		if baseHost != "" {
			postByKey[baseHost+"|"+*p.RemotePostID] = p
		}
	}

	result := content

	// Expand citations: <cite:instance|post-id>
	result = appfed.CitationPatternPublic().ReplaceAllStringFunc(result, func(match string) string {
		sub := appfed.CitationPatternPublic().FindStringSubmatch(match)
		if len(sub) < 3 {
			return match
		}
		instance := sub[1]
		postID := sub[2]

		// Find delivery status.
		status := "pending"
		dKey := "citation|" + instance + "|" + postID
		if d, ok := deliveryByKey[dKey]; ok {
			status = d.Status
		}

		// Find cached post data.
		cacheKey := instance + "|" + postID
		cached := postByKey[cacheKey]

		title := ""
		summary := ""
		postURL := ""
		coverImage := ""
		authorName := ""
		if cached != nil {
			title = cached.Title
			summary = cached.Summary
			postURL = cached.URL
			if cached.CoverImage != nil {
				coverImage = *cached.CoverImage
			}
			if len(cached.Author) > 0 {
				var author struct {
					Name string `json:"name"`
				}
				if err := json.Unmarshal(cached.Author, &author); err == nil {
					authorName = author.Name
				}
			}
		}
		if postURL == "" {
			postURL = "https://" + instance + "/posts/" + postID
		}

		return buildCitationBlock(instance, postID, title, summary, postURL, coverImage, authorName, status)
	})

	// Expand mentions: <@user@instance>
	result = appfed.MentionPatternPublic().ReplaceAllStringFunc(result, func(match string) string {
		sub := appfed.MentionPatternPublic().FindStringSubmatch(match)
		if len(sub) < 3 {
			return match
		}
		user := sub[1]
		instance := sub[2]

		status := "pending"
		dKey := "mention|" + instance + "|" + user
		if d, ok := deliveryByKey[dKey]; ok {
			status = d.Status
		}

		return fmt.Sprintf(
			`<fed-mention data-user="%s" data-instance="%s" data-status="%s">@%s@%s</fed-mention>`,
			html.EscapeString(user),
			html.EscapeString(instance),
			html.EscapeString(status),
			html.EscapeString(user),
			html.EscapeString(instance),
		)
	})

	return result
}

func buildCitationBlock(instance, postID, title, summary, url, coverImage, authorName, status string) string {
	var b strings.Builder
	b.WriteString("::: fed-citation")
	b.WriteString(fmt.Sprintf(` instance="%s"`, escAttr(instance)))
	b.WriteString(fmt.Sprintf(` post-id="%s"`, escAttr(postID)))
	if title != "" {
		b.WriteString(fmt.Sprintf(` title="%s"`, escAttr(title)))
	}
	if summary != "" {
		b.WriteString(fmt.Sprintf(` summary="%s"`, escAttr(summary)))
	}
	if url != "" {
		b.WriteString(fmt.Sprintf(` url="%s"`, escAttr(url)))
	}
	if coverImage != "" {
		b.WriteString(fmt.Sprintf(` cover-image="%s"`, escAttr(coverImage)))
	}
	if authorName != "" {
		b.WriteString(fmt.Sprintf(` author-name="%s"`, escAttr(authorName)))
	}
	b.WriteString(fmt.Sprintf(` status="%s"`, escAttr(status)))
	b.WriteString("\n\n:::")
	return b.String()
}

func escAttr(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, "\n", `\n`)
	return s
}

func deliveryKey(d *federation.OutboundDelivery) string {
	host := extractHost(d.TargetInstanceURL)
	// Extract identifier from payload.
	var payload map[string]any
	if err := json.Unmarshal(d.Payload, &payload); err == nil {
		switch d.DeliveryType {
		case federation.DeliveryTypeCitation:
			if pid, ok := payload["target_post_id"].(string); ok {
				return "citation|" + host + "|" + pid
			}
		case federation.DeliveryTypeMention:
			if user, ok := payload["mentioned_user"].(string); ok {
				return "mention|" + host + "|" + user
			}
		}
	}
	return d.DeliveryType + "|" + host + "|" + d.RequestID
}

func extractHost(rawURL string) string {
	u := strings.TrimSpace(rawURL)
	u = strings.TrimPrefix(u, "https://")
	u = strings.TrimPrefix(u, "http://")
	u = strings.TrimRight(u, "/")
	if idx := strings.Index(u, "/"); idx > 0 {
		u = u[:idx]
	}
	return u
}
