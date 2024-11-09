package com.grtsinry43.grtblog.entity;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableField;
import com.baomidou.mybatisplus.annotation.TableId;

import java.io.Serial;
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
public class Article implements Serializable {

    @Serial
    private static final long serialVersionUID = 1L;

    /**
     * 文章 ID，会由雪花算法生成
     */
    @TableId(value = "id", type = IdType.ASSIGN_ID)
    private Long id;

    /**
     * 文章标题
     */
    @TableField("title")
    private String title;

    /**
     * 文章内容，markdown 格式，交由前端解析
     */
    @TableField("content")
    private String content;

    /**
     * 作者 ID，逻辑限制
     */
    @TableField("author_id")
    private Long authorId;

    /**
     * 文章封面
     */
    private String cover;

    /**
     * 分类 ID
     */
    private Long categoryId;

    /**
     * 文章浏览量
     */
    private Integer views;

    /**
     * 文章点赞量
     */
    private Integer likes;

    /**
     * 文章评论量
     */
    private Integer comments;

    /**
     * 文章状态（PUBLISHED, DRAFT）
     */
    private Status status;

    /**
     * 文章创建时间
     */
    private LocalDateTime createdAt;

    /**
     * 文章更新时间
     */
    private LocalDateTime updatedAt;

    /**
     * 文章删除时间（软删除），如果不为空则表示已删除
     */
    private LocalDateTime deletedAt;

    public enum Status {
        /**
         * 已发布
         */
        PUBLISHED,
        /**
         * 草稿
         */
        DRAFT
    }
}
