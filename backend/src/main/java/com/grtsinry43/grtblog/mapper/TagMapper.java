package com.grtsinry43.grtblog.mapper;

import com.grtsinry43.grtblog.entity.Tag;
import com.baomidou.mybatisplus.core.mapper.BaseMapper;
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
public interface TagMapper extends BaseMapper<Tag> {
    @Select("SELECT article_id FROM article_tag WHERE tag_id = #{tagId}")
    Long[] getArticlesByTagId(Long tagId);

    @Select("SELECT t.* FROM tag t LEFT JOIN article_tag at ON t.id = at.tag_id WHERE at.tag_id IS NULL")
    List<Tag> findTagsWithNoArticles();

    @Select("SELECT id FROM tag WHERE name = #{tagName}")
    Long getTagIdByName(@Param("tagName") String tagName);
}
