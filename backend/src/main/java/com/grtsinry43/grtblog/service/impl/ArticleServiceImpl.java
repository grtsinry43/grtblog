package com.grtsinry43.grtblog.service.impl;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.grtsinry43.grtblog.common.ErrorCode;
import com.grtsinry43.grtblog.dto.ArticleDTO;
import com.grtsinry43.grtblog.dto.PostStatusToggle;
import com.grtsinry43.grtblog.entity.Article;
import com.grtsinry43.grtblog.entity.CommentArea;
import com.grtsinry43.grtblog.entity.User;
import com.grtsinry43.grtblog.exception.BusinessException;
import com.grtsinry43.grtblog.mapper.ArticleMapper;
import com.grtsinry43.grtblog.mapper.TagMapper;
import com.grtsinry43.grtblog.security.LoginUserDetails;
import com.grtsinry43.grtblog.service.CommentAreaService;
import com.grtsinry43.grtblog.service.IArticleService;
import com.baomidou.mybatisplus.extension.service.impl.ServiceImpl;
import com.grtsinry43.grtblog.service.RecommendationService;
import com.grtsinry43.grtblog.service.SocketIOService;
import com.grtsinry43.grtblog.util.ArticleParser;
import com.grtsinry43.grtblog.vo.ArticlePreview;
import com.grtsinry43.grtblog.vo.ArticleVO;
import com.grtsinry43.grtblog.vo.ArticleView;
import org.springframework.beans.BeanUtils;
import org.springframework.context.annotation.Lazy;
import org.springframework.stereotype.Service;
import org.springframework.web.multipart.MultipartFile;
import org.yaml.snakeyaml.Yaml;

import java.io.InputStream;
import java.time.LocalDateTime;
import java.time.ZoneId;
import java.util.*;
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
    private final SocketIOService socketIOService;
    private final ArticleMapper articleMapper;
    private final TagMapper tagMapper;

    public ArticleServiceImpl(ArticleTagServiceImpl articleTagService, TagServiceImpl tagService, CategoryServiceImpl categoryService, UserServiceImpl userService, RecommendationService recommendationService, CommentAreaService commentAreaService, @Lazy SocketIOService socketIOService, ArticleMapper articleMapper, TagMapper tagMapper) {
        this.articleTagService = articleTagService;
        this.tagService = tagService;
        this.categoryService = categoryService;
        this.userService = userService;
        this.recommendationService = recommendationService;
        this.commentAreaService = commentAreaService;
        this.socketIOService = socketIOService;
        this.articleMapper = articleMapper;
        this.tagMapper = tagMapper;
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
        if (articleDTO.getShortUrl() == null || articleDTO.getShortUrl().isEmpty()) {
            article.setShortUrl(ArticleParser.generateShortUrl(articleDTO.getTitle()));
        } else {
            article.setShortUrl(articleDTO.getShortUrl());
        }
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
        if (article.getIsPublished()) {
            recommendationService.updateArticleStatus(article);
        }
        ArticleVO articleVO = new ArticleVO();
        BeanUtils.copyProperties(article, articleVO);
        articleVO.setId(article.getId().toString());
        articleVO.setAuthor(userService.getById(userId).getNickname());
        if (article.getIsPublished()) {
            socketIOService.broadcastNotification("文章：" + article.getTitle() + " 已发布，新鲜出炉，快来看看吧！");
        }
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
        // 删除文章
        article.setDeletedAt(LocalDateTime.now());
        this.baseMapper.updateById(article);
        // 移除文章标签
        articleTagService.deleteArticleTag(articleId);
        // 移除文章推荐状态
        recommendationService.deleteArticleStatus(articleId);
        // 删除评论区
        commentAreaService.deleteCommentArea(article.getCommentId());
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
        article.setUpdatedAt(LocalDateTime.now());
        this.baseMapper.updateById(article);
        articleTagService.syncArticleTag(article.getId(), tagIds);
        if (article.getIsPublished()) {
            recommendationService.updateArticleStatus(article);
        } else {
            recommendationService.deleteArticleStatus(article.getId());
        }
        ArticleVO articleVO = new ArticleVO();
        BeanUtils.copyProperties(article, articleVO);
        articleVO.setId(article.getId().toString());
        articleVO.setAuthor(userService.getById(userId).getNickname());
        if (article.getIsPublished()) {
            socketIOService.broadcastNotification("文章：" + article.getTitle() + " 已更新");
        }
        return articleVO;
    }

    @Override
    public ArticleVO updateArticleStatus(Long articleId, PostStatusToggle postStatusToggle) {
        Article article = this.baseMapper.selectById(articleId);
        if (article == null) {
            throw new BusinessException(ErrorCode.NOT_FOUND);
        }
        article.setIsPublished(postStatusToggle.getIsPublished() == null ? article.getIsPublished() : postStatusToggle.getIsPublished());
        article.setIsTop(postStatusToggle.getIsTop() == null ? article.getIsTop() : postStatusToggle.getIsTop());
        article.setIsHot(postStatusToggle.getIsHot() == null ? article.getIsHot() : postStatusToggle.getIsHot());
        article.setIsOriginal(postStatusToggle.getIsOriginal() == null ? article.getIsOriginal() : postStatusToggle.getIsOriginal());

        this.baseMapper.updateById(article);

        if (article.getIsPublished()) {
            recommendationService.updateArticleStatus(article);
        } else {
            recommendationService.deleteArticleStatus(article.getId());
        }

        ArticleVO articleVO = new ArticleVO();
        BeanUtils.copyProperties(article, articleVO);
        articleVO.setId(article.getId().toString());
        articleVO.setAuthor(userService.getById(article.getAuthorId()).getNickname());
        if (article.getIsPublished()) {
            socketIOService.broadcastNotification("文章：" + article.getTitle() + " 的状态已更新");
        }
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
        articleView.setId(article.getId().toString());
        articleView.setAuthorName(userService.getById(article.getAuthorId()).getNickname());
        articleView.setCategoryName(categoryService.getById(article.getCategoryId()) == null ? "未分类" : categoryService.getById(article.getCategoryId()).getName());
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
            articlePreview.setCategoryShortUrl(categoryService.getShortUrlById(article.getCategoryId()));
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
            articlePreview.setCategoryShortUrl(categoryService.getShortUrlById(article.getCategoryId()));
            articlePreview.setAvatar(userService.getById(article.getAuthorId()).getAvatar());
            articlePreview.setAuthorName(userService.getById(article.getAuthorId()).getNickname());
            return articlePreview;
        }).collect(Collectors.toList());
    }

    public List<ArticlePreview> getArticleListByTag(String tagName, Integer page, Integer pageSize) {
        Long tagId = tagMapper.getTagIdByName(tagName);
        if (tagId == null) {
            return Collections.emptyList();
        }
        int start = (page - 1) * pageSize;
        List<Article> articles = articleMapper.getArticleByTag(tagId, start, pageSize);
        return articles.stream().map(article -> {
            ArticlePreview articlePreview = new ArticlePreview();
            BeanUtils.copyProperties(article, articlePreview);
            articlePreview.setId(article.getId().toString());
            articlePreview.setSummary(!"".equals(article.getSummary()) ? article.getSummary() : article.getContent().length() > 200 ?
                    article.getContent().substring(0, 200) + "..." : article.getContent());
            articlePreview.setTags(String.join(",", tagService.getTagNamesByArticleId(article.getId())));
            articlePreview.setCategoryShortUrl(categoryService.getShortUrlById(article.getCategoryId()));
            articlePreview.setCategoryName(article.getCategoryId() != null ? categoryService.getById(article.getCategoryId()).getName() : "未分类");
            articlePreview.setAvatar(userService.getById(article.getAuthorId()).getAvatar());
            articlePreview.setAuthorName(userService.getById(article.getAuthorId()).getNickname());
            return articlePreview;
        }).collect(Collectors.toList());
    }

    @Override
    public List<ArticleVO> getArticleListAdmin(Integer page, Integer pageSize) {
        List<Article> articles = this.baseMapper.getArticleListByPageAdmin((page - 1) * pageSize, pageSize);
        return articles.stream().map(article -> {
            ArticleVO articleVO = new ArticleVO();
            BeanUtils.copyProperties(article, articleVO);
            articleVO.setId(article.getId().toString());
            articleVO.setAuthor(userService.getById(article.getAuthorId()).getNickname());
            articleVO.setTags(String.join(",", tagService.getTagNamesByArticleId(article.getId())));
            articleVO.setCategoryId(article.getCategoryId() != null ? article.getCategoryId().toString() : null);
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
        articleVO.setCategoryId(article.getCategoryId() != null ? article.getCategoryId().toString() : null);
        return articleVO;
    }

    public ArticleVO importFromHexoMd(MultipartFile file, Long userId) throws Exception {
        InputStream inputStream = file.getInputStream();
        String content = new String(inputStream.readAllBytes());

        // 分割 YAML 前置数据和 Markdown 内容
        String[] parts = content.split("---", 3);
        if (parts.length < 3) {
            throw new BusinessException(ErrorCode.PARAMS_ERROR);
        }

        String yamlContent = parts[1];
        String markdownContent = parts[2];

        // 解析文章并生成目录
        String toc = null;
        try {
            toc = ArticleParser.generateToc(markdownContent);
        } catch (JsonProcessingException e) {
            throw new BusinessException(ErrorCode.OPERATION_ERROR);
        }

        // 解析 YAML 前置数据
        Yaml yaml = new Yaml();
        Map<String, Object> metadata = yaml.load(yamlContent);

        // 验证分类是否存在，存在则设置分类 ID，不存在则创建分类并设置分类 ID
        Long categoryId = null;
        if (metadata.get("category") != null) {
            categoryId = categoryService.getOrCreateCategoryId((String) metadata.get("category"));
        }
        Long[] tagIds = null;

        synchronized (this) {
            try {
                // 同步标签，将 List 转换为 String 数组
                List<String> tagsList = (List<String>) metadata.get("tags");
                String[] tags = tagsList == null ? new String[0] : tagsList.toArray(new String[0]);
                tagIds = Arrays.stream(tags)
                        .map(tagService::getOrCreateTagId)
                        .toArray(Long[]::new);
            } catch (Exception e) {
                throw new BusinessException(ErrorCode.PARAMS_ERROR);
            }
        }
        System.out.println(metadata.get("title"));

        // 创建并保存文章
        Article article = new Article();
        article.setToc(toc);
        article.setTitle(metadata.get("title") == null ? "未命名" : (String) metadata.get("title"));
        User author = userService.getUserByNickname((String) metadata.get("author"));
        if (author != null) {
            article.setAuthorId(author.getId());
        } else {
            article.setAuthorId(userId);
        }
        article.setCategoryId(categoryId);
        article.setContent(markdownContent);
        article.setSummary(markdownContent.length() > 200 ? markdownContent.substring(0, 200) : markdownContent);
        article.setIsPublished(true);
        article.setIsTop(false);
        article.setIsHot(false);
        article.setIsOriginal(true);
        article.setShortUrl(ArticleParser.generateShortUrl(article.getTitle()));
        this.baseMapper.insert(article);
        articleTagService.syncArticleTag(article.getId(), tagIds);
        this.recommendationService.updateArticleStatus(article);
        CommentArea commentArea = this.commentAreaService.createCommentArea("文章", article.getTitle());
        article.setCommentId(commentArea.getId());
        // 设置文章发行时间
        Date date = (Date) metadata.get("date");
        LocalDateTime createdAt = date == null ? LocalDateTime.now() : date.toInstant().atZone(ZoneId.systemDefault()).toLocalDateTime();
        article.setCreatedAt(createdAt);
        article.setUpdatedAt(LocalDateTime.now());
        this.baseMapper.updateById(article);

        ArticleVO articleVO = new ArticleVO();
        BeanUtils.copyProperties(article, articleVO);
        articleVO.setId(article.getId().toString());
        articleVO.setAuthor(userService.getById(userId).getNickname());
        articleVO.setCategoryId(categoryId == null ? null : categoryId.toString());
        return articleVO;
    }


}
