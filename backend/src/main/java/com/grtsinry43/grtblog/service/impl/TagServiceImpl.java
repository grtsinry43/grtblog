package com.grtsinry43.grtblog.service.impl;

import com.grtsinry43.grtblog.common.ErrorCode;
import com.grtsinry43.grtblog.entity.Tag;
import com.grtsinry43.grtblog.exception.BusinessException;
import com.grtsinry43.grtblog.exception.TestException;
import com.grtsinry43.grtblog.mapper.TagMapper;
import com.grtsinry43.grtblog.service.ITagService;
import com.baomidou.mybatisplus.extension.service.impl.ServiceImpl;
import org.springframework.stereotype.Service;

import java.time.LocalDateTime;

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

    @Override
    public Tag addNewTag(String tagName) {
        Tag tag = new Tag();
        tag.setName(tagName);
        tag.setCreated(LocalDateTime.now());
        try {
            save(tag);
        } catch (Exception e) {
            throw new TestException(500, e.getMessage());
        }
        return tag;
    }
}
