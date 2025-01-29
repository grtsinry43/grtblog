package com.grtsinry43.grtblog.controller;

import com.grtsinry43.grtblog.dto.ApiResponse;
import com.grtsinry43.grtblog.entity.Article;
import com.grtsinry43.grtblog.entity.User;
import com.grtsinry43.grtblog.service.RecommendationService;
import com.grtsinry43.grtblog.service.impl.ArticleServiceImpl;
import com.grtsinry43.grtblog.service.impl.TagServiceImpl;
import com.grtsinry43.grtblog.service.impl.UserServiceImpl;
import com.grtsinry43.grtblog.util.SecurityUtils;
import com.grtsinry43.grtblog.vo.ArticlePreview;
import com.grtsinry43.grtblog.vo.ArticleRecommendation;
import com.grtsinry43.grtblog.vo.ArticleVO;
import jakarta.servlet.http.HttpServletRequest;
import org.springframework.beans.BeanUtils;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.Arrays;
import java.util.List;
import java.util.stream.Collectors;

/**
 * @author grtsinry43
 * @date 2025/1/28 14:01
 * @description 热爱可抵岁月漫长
 */
@RestController
@RequestMapping("/recommend")
public class RecommendController {
    private final RecommendationService recommendationService;
    private final UserServiceImpl userService;
    private final ArticleServiceImpl articleService;
    private final TagServiceImpl tagService;

    public RecommendController(RecommendationService recommendationService, UserServiceImpl userService, ArticleServiceImpl articleService, TagServiceImpl tagService) {
        this.recommendationService = recommendationService;
        this.userService = userService;
        this.articleService = articleService;
        this.tagService = tagService;
    }

    @GetMapping
    public ApiResponse<List<ArticlePreview>> recommend(HttpServletRequest request) {
        User currentUser = SecurityUtils.getCurrentUser();
        String userId = currentUser == null ? request.getSession().getId() : currentUser.getId().toString();
        String[] articles = recommendationService.getUserRecommendations(userId, 5);
        if (articles.length == 0) {
            // 取出最热门的文章
            List<Article> filteredArticles = articleService.lambdaQuery()
                    .list().stream()
                    .filter(article -> article.getIsPublished() && article.getDeletedAt() == null)
                    .sorted((a1, a2) -> Long.compare(a2.getViews(), a1.getViews()))
                    .limit(5)
                    .toList();

            articles = filteredArticles.stream()
                    .map(article -> String.valueOf(article.getId()))
                    .toArray(String[]::new);
        }
        List<Long> recommendArticleIds = Arrays.stream(articles)
                .map(Long::parseLong)
                .toList();
        List<ArticlePreview> list = recommendArticleIds.stream().map(id -> {
            Article article = articleService.getById(id);
            ArticlePreview articlePreview = new ArticlePreview();
            BeanUtils.copyProperties(article, articlePreview);
            articlePreview.setId(article.getId().toString());
            articlePreview.setAvatar(userService.getById(article.getAuthorId()).getAvatar());
            articlePreview.setTags(String.join(",", tagService.getTagNamesByArticleId(article.getId())));
            articlePreview.setAuthorName(userService.getById(article.getAuthorId()).getNickname());
            return articlePreview;
        }).toList();
        return ApiResponse.success(list);
    }
}
