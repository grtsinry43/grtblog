package com.grtsinry43.grtblog.vo;

import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.Data;

import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;

/**
 * @author grtsinry43
 * @date 2025/1/31 21:47
 * @description 热爱可抵岁月漫长
 */
@Data
public class ThinkingVO {
    private String id;
    private String content;
    private String author;
    private LocalDateTime createdAt;

    @JsonProperty("createdAt")
    public String getCreatedAt() {
        // 格式化时间：2024-10-27 19:43:00
        return createdAt.format(DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm:ss"));
    }
}
