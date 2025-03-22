package com.grtsinry43.grtblog.vo;

import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.Data;

import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;

/**
 * @author grtsinry43
 * @date 2025/2/3 11:03
 * @description 热爱可抵岁月漫长
 */
@Data
public class FriendLinkVO {
    private String id;

    /**
     * 友链名称
     */
    private String name;

    /**
     * 友链URL
     */
    private String url;

    /**
     * 友链Logo
     */
    private String logo;

    /**
     * 友链描述
     */
    private String description;

    private String userId;

    /**
     * 是否激活
     */
    private Boolean isActive;

    private LocalDateTime createdAt;

    private LocalDateTime updatedAt;

    /**
     * 友链删除时间（软删除），如果不为空则表示已删除
     */
    private LocalDateTime deletedAt;

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
        return deletedAt.format(DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm:ss"));
    }
}
