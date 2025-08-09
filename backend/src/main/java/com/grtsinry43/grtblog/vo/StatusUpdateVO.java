package com.grtsinry43.grtblog.vo;

import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.Data;

import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;

/**
 * @author grtsinry43
 * @date 2024/11/21 19:50
 * @description 热爱可抵岁月漫长
 */
@Data
public class StatusUpdateVO {
    /**
     * 分享 ID，会由雪花算法生成
     */
    private String id;

    /**
     * 分享标题
     */
    private String title;

    /**
     * 分享摘要
     */
    private String summary;

    private String aiSummary;

    /**
     * 分享内容，markdown 格式，交由前端解析
     */
    private String content;

    /**
     * 作者 ID，逻辑限制
     */
    private String authorName;

    /**
     * 分享图片，多个图片用逗号分隔
     */
    private String img;

    /**
     * 分类 ID
     */
    private String categoryId;

    /**
     * 分享浏览量
     */
    private Integer views;

    /**
     * 分享点赞量
     */
    private Integer likes;

    /**
     * 分享评论量
     */
    private Integer comments;

    /**
     * 分享短链接
     */
    private String shortUrl;

    /**
     * 是否发布（0：否，1：是）
     */
    private Boolean isPublished;

    /**
     * 分享创建时间
     */
    private LocalDateTime createdAt;

    /**
     * 分享更新时间
     */
    private LocalDateTime updatedAt;

    /**
     * 分享删除时间（软删除），如果不为空则表示已删除
     */
    private LocalDateTime deletedAt;

    /**
     * 是否置顶（0：否，1：是）
     */
    private Boolean isTop;

    /**
     * 是否热门（0：否，1：是）
     */
    private Boolean isHot;

    /**
     * 是否原创（0：否，1：是）
     */
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
