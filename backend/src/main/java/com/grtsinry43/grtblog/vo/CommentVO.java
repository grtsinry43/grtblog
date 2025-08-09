package com.grtsinry43.grtblog.vo;

import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.Data;

import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;

/**
 * @author grtsinry43
 * @date 2025/2/1 22:25
 * @description 热爱可抵岁月漫长
 */
@Data
public class CommentVO {
    /**
     * 评论 ID，会由雪花算法生成
     */
    private String id;

    /**
     * 评论区 id
     */
    private String areaId;

    /**
     * 评论内容（markdown 格式）
     */
    private String content;

    /**
     * 评论者 ID
     */
    private String authorId;

    /**
     * 评论者昵称
     */
    private String nickName;

    /**
     * 评论者 IP 地址
     */
    private String ip;

    /**
     * 评论者归属地
     */
    private String location;

    /**
     * 评论者操作系统
     */
    private String platform;

    /**
     * 评论者浏览器
     */
    private String browser;

    /**
     * 评论者邮箱
     */
    private String email;

    /**
     * 评论者网站
     */
    private String website;

    /**
     * 评论是否查看过
     */
    private Boolean isViewed;

    /**
     * 置顶
     */
    private Boolean isTop;

    private transient LocalDateTime createdAt;

    private transient LocalDateTime updatedAt;

    /**
     * 评论者是否是博主
     */
    private Boolean isOwner;

    /**
     * 评论者是否是作者
     */
    private Boolean isAuthor;

    /**
     * 评论者是否是好友
     */
    private Boolean isFriend;

    /**
     * 评论删除时间（软删除），如果不为空则表示已删除
     */
    private LocalDateTime deletedAt;

    /**
     * 父评论 ID，如果为空则表示是顶级评论，否则是回复评论
     */
    private String parentId;

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
