package com.grtsinry43.grtblog.service;

import com.grtsinry43.grtblog.entity.Tag;
import com.baomidou.mybatisplus.extension.service.IService;
import com.grtsinry43.grtblog.vo.TagVO;

import java.util.List;

/**
 * <p>
 * 服务类
 * </p>
 *
 * @author grtsinry43
 * @since 2024-10-09
 */
public interface ITagService extends IService<Tag> {
    Tag addNewTag(String tagName);

    Long getOrCreateTagId(String tagName);

    Tag getTagByName(String tagName);

    Long[] getArticlesByTagId(Long tagId);

    List<TagVO> getTagInfoList();

    List<String> getAllTagNames();

    String[] getTagNamesByArticleId(Long articleId);

}
