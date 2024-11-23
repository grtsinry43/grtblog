package com.grtsinry43.grtblog.service.impl;

import com.baomidou.mybatisplus.extension.service.impl.ServiceImpl;
import com.fasterxml.jackson.core.JsonProcessingException;
import com.grtsinry43.grtblog.common.ErrorCode;
import com.grtsinry43.grtblog.dto.PageDTO;
import com.grtsinry43.grtblog.entity.CommentArea;
import com.grtsinry43.grtblog.entity.Page;
import com.grtsinry43.grtblog.exception.BusinessException;
import com.grtsinry43.grtblog.mapper.PageMapper;
import com.grtsinry43.grtblog.service.CommentAreaService;
import com.grtsinry43.grtblog.service.ElasticsearchService;
import com.grtsinry43.grtblog.util.ArticleParser;
import com.grtsinry43.grtblog.vo.PageVO;
import org.springframework.beans.BeanUtils;
import org.springframework.stereotype.Service;

import java.time.LocalDateTime;
import java.util.List;
import java.util.Objects;

@Service
public class PageService extends ServiceImpl<PageMapper, Page> {
    private final CommentAreaService commentAreaService;
    private final ElasticsearchService elasticsearchService;

    public PageService(CommentAreaService commentAreaService, ElasticsearchService elasticsearchService) {
        this.commentAreaService = commentAreaService;
        this.elasticsearchService = elasticsearchService;
    }

    public Page getPageByPath(String path) {
        return this.lambdaQuery().eq(Page::getRefPath, path).one();
    }

    public PageVO getPageByShortUrl(String shortUrl) {
        Page pageFind = this.lambdaQuery().eq(Page::getRefPath, "/" + shortUrl).one();
        if (pageFind == null) {
            throw new BusinessException(ErrorCode.NOT_FOUND);
        } else {
            PageVO pageVO = new PageVO();
            BeanUtils.copyProperties(pageFind, pageVO);
            pageVO.setId(pageFind.getId().toString());
            pageVO.setCommentId(pageFind.getCommentId() != null ? pageFind.getCommentId().toString() : null);
            return pageVO;
        }
    }

    public String[] getAllPageRefPath() {
        return this.lambdaQuery()
                .select(Page::getRefPath)
                .list()
                .stream()
                .filter(Objects::nonNull)
                .filter(page -> page.getEnable() != null && page.getEnable())
                .filter(page -> page.getCanDelete() != null && page.getCanDelete())
                .filter(page -> page.getDeletedAt() == null) // 过滤掉已删除的页面
                .map(Page::getRefPath)
                .toArray(String[]::new);
    }

    public PageVO addPage(PageDTO pageDTO) {
        Page page = new Page();
        BeanUtils.copyProperties(pageDTO, page);
        try {
            page.setToc(ArticleParser.generateToc(page.getContent()));
        } catch (JsonProcessingException e) {
            throw new BusinessException(ErrorCode.OPERATION_ERROR);
        }
        CommentArea commentArea = commentAreaService.createCommentArea("页面", page.getTitle());
        page.setCommentId(commentArea.getId());
        save(page);
        try {
            elasticsearchService.indexPage(page);
        } catch (Exception e) {
            throw new BusinessException(ErrorCode.OPERATION_ERROR, "Failed to index page in Elasticsearch");
        }
        PageVO pageVO = new PageVO();
        BeanUtils.copyProperties(page, pageVO);
        pageVO.setId(page.getId().toString());
        pageVO.setCommentId(page.getCommentId() != null ? page.getCommentId().toString() : null);
        return pageVO;
    }

    public PageVO updatePage(String id, PageDTO pageDTO) {
        Page page = getById(id);
        if (page == null) {
            throw new BusinessException(ErrorCode.NOT_FOUND);
        }
        BeanUtils.copyProperties(pageDTO, page);
        try {
            page.setToc(ArticleParser.generateToc(page.getContent()));
        } catch (JsonProcessingException e) {
            throw new BusinessException(ErrorCode.OPERATION_ERROR);
        }
        updateById(page);
        try {
            elasticsearchService.updatePage(page);
        } catch (Exception e) {
            throw new BusinessException(ErrorCode.OPERATION_ERROR, "Failed to update page in Elasticsearch");
        }
        PageVO pageVO = new PageVO();
        BeanUtils.copyProperties(page, pageVO);
        pageVO.setId(page.getId().toString());
        pageVO.setCommentId(page.getCommentId() != null ? page.getCommentId().toString() : null);
        return pageVO;
    }

    public void deletePage(String id) {
        Page page = getById(id);
        if (page == null) {
            throw new BusinessException(ErrorCode.NOT_FOUND);
        }
        page.setDeletedAt(LocalDateTime.now());
        if (commentAreaService.isExist(page.getCommentId().toString())) {
            commentAreaService.deleteCommentArea(page.getCommentId());
        }
        updateById(page);
        try {
            elasticsearchService.deletePage(id);
        } catch (Exception e) {
            throw new BusinessException(ErrorCode.OPERATION_ERROR, "Failed to delete page from Elasticsearch");
        }
    }

    public List<PageVO> getPageListAdmin(int page, int size) {
        List<Page> pages = this.lambdaQuery()
                .orderByDesc(Page::getCreatedAt)
                .last("limit " + (page - 1) * size + "," + size)
                .list();
        return pages.stream().map(page1 -> {
            PageVO pageVO = new PageVO();
            BeanUtils.copyProperties(page1, pageVO);
            pageVO.setId(page1.getId().toString());
            pageVO.setCommentId(page1.getCommentId() != null ? page1.getCommentId().toString() : null);
            return pageVO;
        }).toList();
    }

    public PageVO getPageByIdAdmin(String id) {
        Page page = getById(id);
        if (page == null) {
            throw new BusinessException(ErrorCode.NOT_FOUND);
        }
        PageVO pageVO = new PageVO();
        BeanUtils.copyProperties(page, pageVO);
        pageVO.setId(page.getId().toString());
        pageVO.setCommentId(page.getCommentId() != null ? page.getCommentId().toString() : null);
        return pageVO;
    }
}
