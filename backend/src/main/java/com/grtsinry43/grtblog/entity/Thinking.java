package com.grtsinry43.grtblog.entity;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import lombok.Data;

import java.time.LocalDateTime;

/**
 * @author grtsinry43
 * @date 2025/1/31 21:38
 * @description 热爱可抵岁月漫长
 */
@Data
public class Thinking {
    @TableId(type = IdType.ASSIGN_ID)
    private Long id;
    private String content;
    private String author;
    private LocalDateTime createdAt;
}
