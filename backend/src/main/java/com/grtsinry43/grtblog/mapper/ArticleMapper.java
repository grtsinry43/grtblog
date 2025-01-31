package com.grtsinry43.grtblog.mapper;

import com.grtsinry43.grtblog.entity.Article;
import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import com.grtsinry43.grtblog.vo.ArticleVO;
import org.apache.ibatis.annotations.Param;
import org.apache.ibatis.annotations.Select;

import java.util.List;

/**
 * <p>
 * Mapper 接口
 * </p>
 *
 * @author grtsinry43
 * @since 2024-10-09
 */
public interface ArticleMapper extends BaseMapper<Article> {
    @Select("SELECT * FROM article WHERE is_published = 1 and deleted_at is null ORDER BY created_at DESC LIMIT 5")
    public List<Article> getLastFiveArticles();

    /**
     * 分页查询文章
     */
    List<Article> getArticleListByPage(@Param("start") Integer start, @Param("pageSize") Integer pageSize);

    /**
     * 分页查询文章
     */
    List<Article> getArticleListByPageAdmin(@Param("start") Integer start, @Param("pageSize") Integer pageSize);

    /**
     * 获取所有文章的短链接
     */
    @Select("SELECT short_url FROM article WHERE is_published = 1 and deleted_at is null")
    List<String> getAllArticleShortLinks();

    /**
     * 根据文章短链接获取文章
     */
    @Select("SELECT * FROM article WHERE is_published = 1 and deleted_at is null and short_url = #{shortUrl}")
    Article getArticleByShortUrl(String shortUrl);

    /**
     * 根据分类 id 分页获取文章列表
     */
    @Select("SELECT * FROM article WHERE category_id = #{categoryId} AND is_published = 1 AND deleted_at is null ORDER BY created_at DESC LIMIT #{start}, #{pageSize}")
    List<Article> getArticleListByCategory(@Param("categoryId") Long categoryId, @Param("start") Integer start, @Param("pageSize") Integer pageSize);

    @Select("SELECT a.* FROM article a " +
            "JOIN article_tag at ON a.id = at.article_id " +
            "WHERE at.tag_id = #{tagId} AND a.is_published = 1 AND a.deleted_at IS NULL " +
            "ORDER BY a.created_at DESC LIMIT #{start}, #{pageSize}")
    List<Article> getArticleByTag(@Param("tagId") Long tagId, @Param("start") Integer start, @Param("pageSize") Integer pageSize);
}
