package com.grtsinry43.grtblog.controller;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.grtsinry43.grtblog.dto.ApiResponse;
import com.grtsinry43.grtblog.dto.ArticleDTO;
import com.grtsinry43.grtblog.entity.Article;
import com.grtsinry43.grtblog.service.impl.ArticleServiceImpl;
import com.grtsinry43.grtblog.util.ArticleParser;
import com.grtsinry43.grtblog.vo.ArticlePreview;
import com.grtsinry43.grtblog.vo.ArticleVO;
import com.grtsinry43.grtblog.vo.ArticleView;
import jakarta.annotation.security.PermitAll;
import jakarta.servlet.http.HttpServletRequest;
import org.springframework.beans.BeanUtils;
import org.springframework.security.access.prepost.PreAuthorize;
import org.springframework.web.bind.annotation.*;

import java.util.List;

/**
 * 文章控制器
 *
 * @author grtsinry43
 * @since 2024-10-09
 */
@RestController
@RequestMapping("/article")
public class ArticleController {
    private final ArticleServiceImpl articleService;

    public ArticleController(ArticleServiceImpl articleService) {
        this.articleService = articleService;
    }

    /**
     * 获取所有文章短链接
     */
    @PermitAll
    @GetMapping("/shortLinks")
    public ApiResponse<List<String>> getAllArticleShortLinks() {
        List<String> articleShortLinks = articleService.getAllArticleShortLinks();
        return ApiResponse.success(articleShortLinks);
    }

    @PermitAll
    @GetMapping("/all")
    public ApiResponse<List<ArticlePreview>> getArticleListByPage(@RequestParam Integer page, @RequestParam Integer pageSize) {
        List<ArticlePreview> articleList = articleService.getArticleListByPage(page, pageSize);
        return ApiResponse.success(articleList);
    }

    @PermitAll
    @GetMapping("/lastFive")
    public ApiResponse<List<ArticlePreview>> getLastFiveArticles() {
        List<ArticlePreview> lastFiveArticles = articleService.getLastFiveArticleList();
        return ApiResponse.success(lastFiveArticles);
    }

    @PermitAll
    @GetMapping("/recommend/{shortUrl}")
    public ApiResponse<List<ArticlePreview>> getRecommendArticles(@PathVariable String shortUrl) {
        List<ArticlePreview> recommendArticles = articleService.getRecommendArticleList(shortUrl);
        return ApiResponse.success(recommendArticles);
    }

    @PermitAll
    @GetMapping("/category/{shortUrl}")
    public ApiResponse<List<ArticlePreview>> getArticleListByCategory(@RequestParam Integer page, @RequestParam Integer pageSize, @PathVariable String shortUrl) {
        List<ArticlePreview> articleList = articleService.getArticleListByCategory(shortUrl, page, pageSize);
        return ApiResponse.success(articleList);
    }

    @PermitAll
    @GetMapping("/tag/{tagName}")
    public ApiResponse<List<ArticlePreview>> getArticleListByTag(@RequestParam Integer page, @RequestParam Integer pageSize, @PathVariable String tagName) {
        List<ArticlePreview> articleList = articleService.getArticleListByTag(tagName, page, pageSize);
        return ApiResponse.success(articleList);
    }

    // 这个匹配的一定要放在最后，不然会被 /all 匹配
    @PermitAll
    @GetMapping("/{shortUrl}")
    public ApiResponse<ArticleView> viewOneArticleApi(@PathVariable String shortUrl) {
        ArticleView articleView = articleService.viewOneArticle(shortUrl);
        return ApiResponse.success(articleView);
    }


}
