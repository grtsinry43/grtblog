package com.grtsinry43.grtblog.vo;

import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.Data;

import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;

/**
 * @author grtsinry43
 * @date 2024/10/11 18:55
 * @description 热爱可抵岁月漫长
 */
@Data
public class UserVO {
    private String id;
    private String nickname;
    private String email;
    private String avatar;
    private LocalDateTime createdAt;
    private String oauthProvider;

    @JsonProperty("createdAt")
    public String getCreatedAt() {
        // 格式化时间：2024-10-27 19:43:00
        return createdAt.format(DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm:ss"));
    }
}
