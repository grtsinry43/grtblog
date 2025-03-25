package com.grtsinry43.grtblog.vo;

import com.fasterxml.jackson.annotation.JsonProperty;
import com.grtsinry43.grtblog.entity.Article;
import lombok.Data;

import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;

/**
 * @author grtsinry43
 * @date 2024/10/27 18:47
 * @description 热爱可抵岁月漫长
 */
@Data
public class ArticleVO {
    private String id;

    private String title;

    private String summary;

    private String aiSummary;

    private String toc;

    private String content;

    private String author;

    private String cover;

    private String categoryId;

    private String tags;

    private Integer views;

    private Integer likes;

    private Integer comments;

    private String shortUrl;

    private Boolean isPublished;

    private LocalDateTime createdAt;

    private LocalDateTime updatedAt;

    private LocalDateTime deletedAt;

    private Boolean isTop;

    private Boolean isHot;

    private Boolean isOriginal;

    @JsonProperty("createdAt")
    public String getCreatedAt() {
        // 格式化时间：2024-10-27 19:43:00
        return createdAt.format(DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm:ss"));
    }

    @JsonProperty("updatedAt")
    public String getUpdatedAt() {
        // 格式化时间：2024-10-27 19:43:00
        return updatedAt.format(DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm:ss"));
    }

    @JsonProperty("deletedAt")
    public String getDeletedAt() {
        // 格式化时间：2024-10-27 19:43:00
        return deletedAt != null ? deletedAt.format(DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm:ss")) : null;
    }
}
