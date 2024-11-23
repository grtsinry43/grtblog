package com.grtsinry43.grtblog.service.impl;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.grtsinry43.grtblog.common.ErrorCode;
import com.grtsinry43.grtblog.dto.ArticleDTO;
import com.grtsinry43.grtblog.entity.Article;
import com.grtsinry43.grtblog.entity.CommentArea;
import com.grtsinry43.grtblog.exception.BusinessException;
import com.grtsinry43.grtblog.mapper.ArticleMapper;
import com.grtsinry43.grtblog.security.LoginUserDetails;
import com.grtsinry43.grtblog.service.CommentAreaService;
import com.grtsinry43.grtblog.service.ElasticsearchService;
import com.grtsinry43.grtblog.service.IArticleService;
import com.baomidou.mybatisplus.extension.service.impl.ServiceImpl;
import com.grtsinry43.grtblog.service.RecommendationService;
import com.grtsinry43.grtblog.util.ArticleParser;
import com.grtsinry43.grtblog.vo.ArticlePreview;
import com.grtsinry43.grtblog.vo.ArticleVO;
import com.grtsinry43.grtblog.vo.ArticleView;
import org.springframework.beans.BeanUtils;
import org.springframework.stereotype.Service;

import java.util.Arrays;
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
public class ArticleServiceImpl extends ServiceImpl<ArticleMapper, Article> implements IArticleService {
    private final ArticleTagServiceImpl articleTagService;
    private final TagServiceImpl tagService;
    private final CategoryServiceImpl categoryService;
    private final UserServiceImpl userService;
    private final RecommendationService recommendationService;
    private final CommentAreaService commentAreaService;
    private final ElasticsearchService elasticsearchService;

    public ArticleServiceImpl(ArticleTagServiceImpl articleTagService, TagServiceImpl tagService, CategoryServiceImpl categoryService, UserServiceImpl userService, RecommendationService recommendationService, CommentAreaService commentAreaService, ElasticsearchService elasticsearchService) {
        this.articleTagService = articleTagService;
        this.tagService = tagService;
        this.categoryService = categoryService;
        this.userService = userService;
        this.recommendationService = recommendationService;
        this.commentAreaService = commentAreaService;
        this.elasticsearchService = elasticsearchService;
    }

    @Override
    public ArticleVO addArticle(ArticleDTO articleDTO, Long userId) {
        Article article = new Article();
        BeanUtils.copyProperties(articleDTO, article);
        article.setAuthorId(userId);
        article.setCategoryId(Long.parseLong(articleDTO.getCategoryId()));
        // 解析文章并生成目录
        String toc = null;
        try {
            toc = ArticleParser.generateToc(articleDTO.getContent());
        } catch (JsonProcessingException e) {
            throw new BusinessException(ErrorCode.OPERATION_ERROR);
        }
        article.setToc(toc);
        if (!categoryService.isCategoryExist(Long.parseLong(articleDTO.getCategoryId()))) {
            throw new BusinessException(ErrorCode.PARAMS_ERROR);
        }
        // 添加短链接
        article.setShortUrl(ArticleParser.generateShortUrl(articleDTO.getTitle()));
        // 处理标签
        String[] tagNames = articleDTO.getTags().split(",");
        Long[] tagIds = Arrays.stream(tagNames)
                .map(tagService::getOrCreateTagId)
                .toArray(Long[]::new);
        if (article.getSummary() == null) {
            article.setSummary(article.getContent().length() > 200 ? article.getContent().substring(0, 200) : article.getContent());
        }
        // 创建评论区
        CommentArea commentArea = commentAreaService.createCommentArea("文章", article.getTitle());
        article.setCommentId(commentArea.getId());
        this.baseMapper.insert(article);
        articleTagService.syncArticleTag(article.getId(), tagIds);
        recommendationService.updateArticleStatus(article);
        try {
            elasticsearchService.indexArticle(article);
        } catch (Exception e) {
            throw new BusinessException(ErrorCode.OPERATION_ERROR, "Failed to index article in Elasticsearch");
        }
        ArticleVO articleVO = new ArticleVO();
        BeanUtils.copyProperties(article, articleVO);
        articleVO.setId(article.getId().toString());
        articleVO.setAuthor(userService.getById(userId).getNickname());
        return articleVO;
    }

    @Override
    public void deleteArticle(Long articleId, LoginUserDetails principal) {
        Article article = this.baseMapper.selectById(articleId);
        if (article == null) {
            throw new BusinessException(ErrorCode.NOT_FOUND);
        }
        // 这里角色是管理员或者超级管理员可以删除所有文章，合作作者只能删除自己的文章
        if (principal.getRoleNames().contains("ROLE_WRITER") && !article.getAuthorId().equals(principal.getUser().getId())) {
            throw new BusinessException(ErrorCode.UNAUTHORIZED);
        }
        this.baseMapper.deleteById(articleId);
        // 移除文章标签
        articleTagService.deleteArticleTag(articleId);
        // 移除文章推荐状态
        recommendationService.deleteArticleStatus(articleId);
        // 删除评论区
        commentAreaService.deleteCommentArea(article.getCommentId());
        try {
            elasticsearchService.deleteArticle(articleId.toString());
        } catch (Exception e) {
            throw new BusinessException(ErrorCode.OPERATION_ERROR, "Failed to delete article from Elasticsearch");
        }
    }

    @Override
    public ArticleVO updateArticle(Long articleId, ArticleDTO articleDTO, Long userId) {
        Article article = this.baseMapper.selectById(articleId);
        BeanUtils.copyProperties(articleDTO, article);
        article.setCategoryId(Long.parseLong(articleDTO.getCategoryId()));
        article.setAuthorId(userId);
        // 解析文章并生成目录
        String toc = null;
        try {
            toc = ArticleParser.generateToc(articleDTO.getContent());
        } catch (JsonProcessingException e) {
            throw new BusinessException(ErrorCode.OPERATION_ERROR);
        }
        article.setToc(toc);
        if (!categoryService.isCategoryExist(Long.parseLong(articleDTO.getCategoryId()))) {
            throw new BusinessException(ErrorCode.PARAMS_ERROR);
        }
        // 处理标签
        String[] tagNames = articleDTO.getTags().split(",");
        Long[] tagIds = Arrays.stream(tagNames)
                .map(tagService::getOrCreateTagId)
                .toArray(Long[]::new);
        if (article.getSummary() == null) {
            article.setSummary(article.getContent().length() > 200 ? article.getContent().substring(0, 200) : article.getContent());
        }
        this.baseMapper.updateById(article);
        articleTagService.syncArticleTag(article.getId(), tagIds);
        recommendationService.updateArticleStatus(article);
        try {
            elasticsearchService.updateArticle(article);
        } catch (Exception e) {
            throw new BusinessException(ErrorCode.OPERATION_ERROR, "Failed to update article in Elasticsearch");
        }
        ArticleVO articleVO = new ArticleVO();
        BeanUtils.copyProperties(article, articleVO);
        articleVO.setId(article.getId().toString());
        articleVO.setAuthor(userService.getById(userId).getNickname());
        return articleVO;
    }

    @Override
    public ArticleView viewOneArticle(String shortUrl) {
        Article article = this.baseMapper.getArticleByShortUrl(shortUrl);
        if (article == null) {
            throw new BusinessException(ErrorCode.NOT_FOUND);
        }
        ArticleView articleView = new ArticleView();
        BeanUtils.copyProperties(article, articleView);
        articleView.setAuthorName(userService.getById(article.getAuthorId()).getNickname());
        articleView.setCategoryName(categoryService.getById(article.getCategoryId()).getName());
        articleView.setTags(String.join(",", tagService.getTagNamesByArticleId(article.getId())));
        articleView.setCommentId(article.getCommentId().toString());
        return articleView;
    }


    @Override
    public List<String> getAllArticleShortLinks() {
        return this.baseMapper.getAllArticleShortLinks();
    }

    @Override
    public List<ArticlePreview> getLastFiveArticleList() {
        try {
            List<Article> articles = this.baseMapper.getLastFiveArticles();
            return articles.stream().map(article -> {
                ArticlePreview articlePreview = new ArticlePreview();
                BeanUtils.copyProperties(article, articlePreview);
                articlePreview.setId(article.getId().toString());
                articlePreview.setAvatar(userService.getById(article.getAuthorId()).getAvatar());
                articlePreview.setAuthorName(userService.getById(article.getAuthorId()).getNickname());
                articlePreview.setTags(String.join(",", tagService.getTagNamesByArticleId(article.getId())));
                return articlePreview;
            }).collect(Collectors.toList());
        } catch (Exception e) {
            System.out.println(e.getMessage());
        }
        return null;
    }

    @Override
    public List<ArticlePreview> getRecommendArticleList(String shortUrl) {
        Long articleId = this.baseMapper.getArticleByShortUrl(shortUrl).getId();
        String[] recommendArticleIdsArray = recommendationService.getRecommendations(articleId, 5);
        List<Long> recommendArticleIds = Arrays.stream(recommendArticleIdsArray)
                .map(Long::parseLong)
                .toList();
        return recommendArticleIds.stream().map(id -> {
            Article article = this.baseMapper.selectById(id);
            ArticlePreview articlePreview = new ArticlePreview();
            BeanUtils.copyProperties(article, articlePreview);
            articlePreview.setId(article.getId().toString());
            articlePreview.setAvatar(userService.getById(article.getAuthorId()).getAvatar());
            articlePreview.setTags(String.join(",", tagService.getTagNamesByArticleId(article.getId())));
            articlePreview.setAuthorName(userService.getById(article.getAuthorId()).getNickname());
            return articlePreview;
        }).collect(Collectors.toList());
    }

    @Override
    public List<ArticlePreview> getArticleListByPage(Integer page, Integer pageSize) {
        List<Article> articles = this.baseMapper.getArticleListByPage((page - 1) * pageSize, pageSize);
        return articles.stream().map(article -> {
            ArticlePreview articlePreview = new ArticlePreview();
            BeanUtils.copyProperties(article, articlePreview);
            articlePreview.setId(article.getId().toString());
            articlePreview.setSummary(!"".equals(article.getSummary()) ? article.getSummary() : article.getContent().length() > 200 ?
                    article.getContent().substring(0, 200) + "..." : article.getContent());
            articlePreview.setCategoryName(article.getCategoryId() != null ? categoryService.getById(article.getCategoryId()).getName() : "未分类");
            articlePreview.setAvatar(userService.getById(article.getAuthorId()).getAvatar());
            articlePreview.setTags(String.join(",", tagService.getTagNamesByArticleId(article.getId())));
            articlePreview.setAuthorName(userService.getById(article.getAuthorId()).getNickname());
            return articlePreview;
        }).collect(Collectors.toList());
    }

    @Override
    public List<ArticlePreview> getArticleListByCategory(String shortUrl, Integer page, Integer pageSize) {
        Long categoryId = categoryService.getCategoryIdByShortUrl(shortUrl);
        List<Article> articles = this.baseMapper.getArticleListByCategory(categoryId, (page - 1) * pageSize, pageSize);
        return articles.stream().map(article -> {
            ArticlePreview articlePreview = new ArticlePreview();
            BeanUtils.copyProperties(article, articlePreview);
            articlePreview.setId(article.getId().toString());
            articlePreview.setSummary(!"".equals(article.getSummary()) ? article.getSummary() : article.getContent().length() > 200 ?
                    article.getContent().substring(0, 200) + "..." : article.getContent());
            articlePreview.setTags(String.join(",", tagService.getTagNamesByArticleId(article.getId())));
            articlePreview.setCategoryName(article.getCategoryId() != null ? categoryService.getById(article.getCategoryId()).getName() : "未分类");
            articlePreview.setAvatar(userService.getById(article.getAuthorId()).getAvatar());
            articlePreview.setAuthorName(userService.getById(article.getAuthorId()).getNickname());
            return articlePreview;
        }).collect(Collectors.toList());
    }

    @Override
    public List<ArticleVO> getArticleListAdmin(Integer page, Integer pageSize) {
        List<Article> articles = this.baseMapper.getArticleListByPage((page - 1) * pageSize, pageSize);
        return articles.stream().map(article -> {
            ArticleVO articleVO = new ArticleVO();
            BeanUtils.copyProperties(article, articleVO);
            articleVO.setId(article.getId().toString());
            articleVO.setAuthor(userService.getById(article.getAuthorId()).getNickname());
            articleVO.setTags(String.join(",", tagService.getTagNamesByArticleId(article.getId())));
            articleVO.setCategoryId(article.getCategoryId().toString());
            return articleVO;
        }).collect(Collectors.toList());
    }

    @Override
    public ArticleVO getArticleByIdAdmin(Long id) {
        Article article = this.baseMapper.selectById(id);
        ArticleVO articleVO = new ArticleVO();
        BeanUtils.copyProperties(article, articleVO);
        articleVO.setId(article.getId().toString());
        articleVO.setAuthor(userService.getById(article.getAuthorId()).getNickname());
        articleVO.setTags(String.join(",", tagService.getTagNamesByArticleId(article.getId())));
        articleVO.setCategoryId(article.getCategoryId().toString());
        return articleVO;
    }


}
