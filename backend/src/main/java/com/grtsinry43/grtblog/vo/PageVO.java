package com.grtsinry43.grtblog.vo;

import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.Data;

import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;

/**
 * @author grtsinry43
 * @date 2024/11/22 11:02
 * @description 热爱可抵岁月漫长
 */
@Data
public class PageVO {
    private String id;
    private String title;
    private String description;
    private String aiSummary;
    private String refPath;
    private String toc;
    private String content;
    private Integer views;
    private Integer likes;
    private Integer comments;
    private String commentId;
    private Boolean enable;
    private Boolean canDelete;
    private LocalDateTime createdAt;
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
