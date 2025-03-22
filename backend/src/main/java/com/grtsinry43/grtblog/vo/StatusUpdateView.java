package com.grtsinry43.grtblog.vo;

import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.Data;

import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;

/**
 * @author grtsinry43
 * @date 2024/11/1 13:04
 * @description 热爱可抵岁月漫长
 */
@Data
public class StatusUpdateView {
    private String id;
    /**
     * 说说短链接
     */
    private String shortUrl;

    /**
     * 作者名字
     */
    private String authorName;

    /**
     * 作者头像
     */
    private String authorAvatar;

    private String categoryName;

    /**
     * 图片
     */
    private String[] images;

    private String toc;

    /**
     * 说说标题
     */
    private String title;

    /**
     * 内容（Markdown 格式，由前端解析，不解析图片）
     */
    private String content;

    private String summary;

    private String aiSummary;

    /**
     * 查看次数
     */
    private Integer views;

    /**
     * 评论次数
     */
    private Integer comments;

    /**
     * 点赞次数
     */
    private Integer likes;

    /**
     * 评论区 id
     */
    private String commentId;

    /**
     * 是否置顶
     */
    private Boolean isTop;

    /**
     * 是否热门
     */
    private Boolean isHot;

    /**
     * 是否原创
     */
    private Boolean isOriginal;

    /**
     * 说说创建时间
     */
    private LocalDateTime createdAt;

    /**
     * 说说更新时间
     */
    private LocalDateTime updatedAt;

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
