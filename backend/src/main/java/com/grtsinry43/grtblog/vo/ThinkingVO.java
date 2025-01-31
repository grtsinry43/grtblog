package com.grtsinry43.grtblog.vo;

import lombok.Data;

import java.time.LocalDateTime;

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
}
