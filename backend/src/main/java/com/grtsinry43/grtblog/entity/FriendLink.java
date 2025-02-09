package com.grtsinry43.grtblog.entity;

import com.baomidou.mybatisplus.annotation.*;

import java.io.Serial;
import java.io.Serializable;
import java.time.LocalDateTime;

import lombok.Data;
import lombok.Getter;
import lombok.Setter;

/**
 * <p>
 * 
 * </p>
 *
 * @author grtsinry43
 * @since 2024-10-09
 */
@Data
@TableName("friend_link")
public class FriendLink implements Serializable {

    @Serial
    private static final long serialVersionUID = 1L;

    /**
     * 友链ID，会由雪花算法生成
     */
    @TableId(value = "id", type = IdType.AUTO)
    private Long id;

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

    private long userId;

    /**
     * 是否激活
     */
    private Boolean isActive;

    @TableField(value = "created_at", insertStrategy = FieldStrategy.NEVER, updateStrategy = FieldStrategy.NEVER)
    private LocalDateTime createdAt;
    @TableField(value = "updated_at", insertStrategy = FieldStrategy.NEVER, updateStrategy = FieldStrategy.NEVER)
    private LocalDateTime updatedAt;

    /**
     * 友链删除时间（软删除），如果不为空则表示已删除
     */
    private LocalDateTime deletedAt;
}
