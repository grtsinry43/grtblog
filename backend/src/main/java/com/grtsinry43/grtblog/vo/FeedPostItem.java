package com.grtsinry43.grtblog.vo;

import com.fasterxml.jackson.annotation.JsonProperty;
import com.grtsinry43.grtblog.entity.Article;
import com.grtsinry43.grtblog.entity.Page;
import com.grtsinry43.grtblog.entity.StatusUpdate;
import com.grtsinry43.grtblog.util.MarkdownConverter;
import lombok.Data;

import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;
import java.util.ArrayList;
import java.util.Comparator;
import java.util.List;

/**
 * @author grtsinry43
 * @date 2024/12/27 11:25
 * @description 热爱可抵岁月漫长
 */
@Data
public class FeedPostItem {
    private String title;
    private String id;
    private String content;
    private String url;
    private LocalDateTime publishedAt;
    private String cover;

    @JsonProperty("publishedAt")
    public String getPublishedAt() {
        // 格式化时间：2024-10-27 19:43:00
        return publishedAt.format(DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm:ss"));
    }

    public static List<FeedPostItem> buildFeed(List<Article> articles, List<StatusUpdate> statusUpdates, List<Page> pages, String websiteUrl) {
        List<FeedPostItem> feedPostItems = new ArrayList<>();
        for (Article article : articles) {
            FeedPostItem feedPostItem = new FeedPostItem();
            feedPostItem.setTitle(article.getTitle());
            feedPostItem.setId(article.getId().toString());
            feedPostItem.setContent(MarkdownConverter.convertMarkdownToHtml(article.getContent()));
            feedPostItem.setUrl(websiteUrl + "/posts/" + article.getShortUrl());
            feedPostItem.setPublishedAt(article.getCreatedAt());
            feedPostItem.setCover(article.getCover());
            feedPostItems.add(feedPostItem);
        }
        for (StatusUpdate statusUpdate : statusUpdates) {
            FeedPostItem feedPostItem = new FeedPostItem();
            feedPostItem.setTitle(statusUpdate.getTitle());
            feedPostItem.setId(statusUpdate.getId().toString());
            feedPostItem.setContent(MarkdownConverter.convertMarkdownToHtml(statusUpdate.getContent()));
            feedPostItem.setUrl(websiteUrl + "/moments/" + statusUpdate.getCreatedAt().format(DateTimeFormatter.ofPattern("yyyy/MM/dd")) + "/" + statusUpdate.getShortUrl());
            feedPostItem.setPublishedAt(statusUpdate.getCreatedAt());
            feedPostItem.setCover(statusUpdate.getImg() == null ? null : statusUpdate.getImg().split(",")[0].startsWith("http") ? statusUpdate.getImg().split(",")[0] : websiteUrl + statusUpdate.getImg().split(",")[0]);
            feedPostItems.add(feedPostItem);
        }
        for (Page page : pages) {
            FeedPostItem feedPostItem = new FeedPostItem();
            feedPostItem.setTitle(page.getTitle());
            feedPostItem.setId(page.getId().toString());
            feedPostItem.setContent(MarkdownConverter.convertMarkdownToHtml(page.getContent()));
            feedPostItem.setUrl(websiteUrl+ "/" + page.getRefPath());
            feedPostItem.setPublishedAt(page.getCreatedAt());
            feedPostItem.setCover(null);
            feedPostItems.add(feedPostItem);
        }
        feedPostItems.sort(Comparator.comparing(FeedPostItem::getPublishedAt).reversed());
        return feedPostItems;
    }
}
