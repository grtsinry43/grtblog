package com.grtsinry43.grtblog.vo;

import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.Data;

import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;

/**
 * @author grtsinry43
 * @date 2025/1/1 13:17
 * @description 热爱可抵岁月漫长
 */
@Data
public class GlobalNotificationVO {
    private String id;
    private String content;
    private LocalDateTime publishAt;
    private LocalDateTime expireAt;
    private Boolean allowClose;

    @JsonProperty("publishAt")
    public String getPublishAt() {
        // 格式化时间：2024-10-27 19:43:00
        return publishAt.format(DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm:ss"));
    }

    @JsonProperty("expireAt")
    public String getExpireAt() {
        // 格式化时间：2024-10-27 19:43:00
        return expireAt.format(DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm:ss"));
    }
}
