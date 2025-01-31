package com.grtsinry43.grtblog.service.impl;

import com.grtsinry43.grtblog.entity.Tag;
import com.grtsinry43.grtblog.exception.TestException;
import com.grtsinry43.grtblog.mapper.ArticleTagMapper;
import com.grtsinry43.grtblog.mapper.TagMapper;
import com.grtsinry43.grtblog.service.ITagService;
import com.baomidou.mybatisplus.extension.service.impl.ServiceImpl;
import com.grtsinry43.grtblog.vo.TagVO;
import org.springframework.stereotype.Service;

import java.time.LocalDateTime;
import java.util.List;
import java.util.stream.Collectors;

/**
 * <p>
 * 服务实现类
 * </p>
 *
 * @author grtsinry43
 * @since 2024-10-09
 */
@Service
public class TagServiceImpl extends ServiceImpl<TagMapper, Tag> implements ITagService {
    private final ArticleTagMapper articleTagMapper;

    public TagServiceImpl(ArticleTagMapper articleTagMapper) {
        this.articleTagMapper = articleTagMapper;
    }

    @Override
    public Tag addNewTag(String tagName) {
        Tag tag = new Tag();
        tag.setName(tagName);
        tag.setCreatedAt(LocalDateTime.now());
        try {
            save(tag);
        } catch (Exception e) {
            throw new TestException(500, e.getMessage());
        }
        return tag;
    }

    @Override
    public Long getOrCreateTagId(String tagName) {
        Tag tag = getTagByName(tagName);
        if (tag == null) {
            tag = addNewTag(tagName);
        }
        return tag.getId();
    }

    @Override
    public Tag getTagByName(String tagName) {
        return lambdaQuery().eq(Tag::getName, tagName).one();
    }

    @Override
    public Long[] getArticlesByTagId(Long tagId) {
        return baseMapper.getArticlesByTagId(tagId);
    }

    @Override
    public List<TagVO> getTagInfoList() {
        List<Tag> tags = list();
        return tags.stream().map(tag -> {
            TagVO tagVO = new TagVO();
            tagVO.setTagId(tag.getId().toString());
            tagVO.setTagName(tag.getName());
            tagVO.setArticleCount(articleTagMapper.countArticlesByTagId(tag.getId()));
            return tagVO;
        }).collect(Collectors.toList());
    }

    @Override
    public List<String> getAllTagNames(){
        List<Tag> tags = list();
        return tags.stream().map(Tag::getName).toList();
    }

    @Override
    public String[] getTagNamesByArticleId(Long articleId) {
        return articleTagMapper.getTagNamesByArticleId(articleId);
    }

}
