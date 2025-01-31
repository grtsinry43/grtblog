package com.grtsinry43.grtblog.vo;

import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.Data;

import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;

/**
 * @author grtsinry43
 * @date 2024/10/31 19:12
 * @description 热爱可抵岁月漫长
 */
@Data
public class ArticlePreview {
    private String id;

    /**
     * 文章标题
     */
    private String title;

    /**
     * 文章短链接
     */
    private String shortUrl;

    /**
     * 作者名字
     */
    private String authorName;

    /**
     * 文章简介
     */
    private String summary;

    /**
     * 作者头像
     */
    private String avatar;

    /**
     * 文章封面
     */
    private String cover;

    /**
     * 文章浏览量
     */
    private Integer views;

    /**
     * 文章分类名称
     */
    private String categoryName;

    /**
     * 文章分类短链接
     */
    private String categoryShortUrl;

    /**
     * 文章标签
     */
    private String tags;

    /**
     * 文章点赞量
     */
    private Integer likes;

    /**
     * 文章评论量
     */
    private Integer comments;

    /**
     * 是否置顶
     */
    private Boolean isTop;

    /**
     * 文章创建时间（发布时间，以这个为准）
     */
    private transient LocalDateTime createdAt;

    /**
     * 文章更新时间
     */
    private transient LocalDateTime updatedAt;

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
}
