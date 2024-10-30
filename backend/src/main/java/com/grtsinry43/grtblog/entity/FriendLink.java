package com.grtsinry43.grtblog.entity;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import java.io.Serializable;
import java.time.LocalDateTime;
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
@Getter
@Setter
@TableName("friend_link")
public class FriendLink implements Serializable {

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

    /**
     * 友链创建时间
     */
    private LocalDateTime created;

    /**
     * 友链更新时间
     */
    private LocalDateTime updated;

    /**
     * 友链删除时间（软删除），如果不为空则表示已删除
     */
    private LocalDateTime deleted;
}