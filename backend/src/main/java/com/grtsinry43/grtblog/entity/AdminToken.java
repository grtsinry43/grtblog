package com.grtsinry43.grtblog.entity;

import com.baomidou.mybatisplus.annotation.*;
import lombok.Data;

import java.sql.Timestamp;

/**
 * @author grtsinry43
 * @date 2025/1/26 20:25
 * @description 热爱可抵岁月漫长
 */
@Data
@TableName("admin_token")
public class AdminToken {

    @TableId(type = IdType.ASSIGN_ID)
    private Long id;

    private String token;

    private Long userId;

    private String description;

    private Timestamp expireAt;

    @TableField(value = "created_at", insertStrategy = FieldStrategy.NEVER)
    private Timestamp createdAt;

    @TableField(value = "updated_at", insertStrategy = FieldStrategy.NEVER)
    private Timestamp updatedAt;
}
