package com.grtsinry43.grtblog.vo;

import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.Data;

import java.time.LocalDate;
import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;

/**
 * @author grtsinry43
 * @date 2025/2/8 12:53
 * @description 热爱可抵岁月漫长
 */
@Data
public class PhotoPreview {
    private String id;
    private String url;
    private String device;
    private String location;
    private String description;
    private LocalDateTime date;
    private String shade;

    @JsonProperty("date")
    public String getDate() {
        // 格式化时间：2024-10-27 19:43:00
        return date.format(DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm:ss"));
    }

}
