package com.grtsinry43.grtblog.dto;

import lombok.Data;

import java.sql.Timestamp;

/**
 * @author grtsinry43
 * @date 2025/1/26 20:29
 * @description 热爱可抵岁月漫长
 */
@Data
public class AdminTokenRequest {
    private String description;
    private Timestamp expireAt;
}
