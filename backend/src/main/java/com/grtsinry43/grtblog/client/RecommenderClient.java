package com.grtsinry43.grtblog.client;

import com.grtsinry43.grtblog.dto.ApiResponse;
import com.grtsinry43.grtblog.dto.ArticleRecommendUpdate;
import com.grtsinry43.grtblog.vo.ArticleRecommendation;
import org.springframework.cloud.openfeign.FeignClient;
import org.springframework.web.bind.annotation.*;

/**
 * @author grtsinry43
 * @date 2024/11/8 00:24
 * @description 热爱可抵岁月漫长
 */

@FeignClient(name = "recommender-service", url = "http://localhost:8001")
public interface RecommenderClient {

    @GetMapping("/article/{article_id}")
    ApiResponse<ArticleRecommendation> getRecommendations(@PathVariable("article_id") String articleId, @RequestParam("count") Integer size);

    @PostMapping("/article/{article_id}")
    ApiResponse<Object> updateArticleStatus(@PathVariable("article_id") String articleId, @RequestBody ArticleRecommendUpdate articleRecommendUpdate);

    @DeleteMapping("/article/{article_id}")
    ApiResponse<Object> deleteArticle(@PathVariable("article_id") String articleId);

    @GetMapping("/user/{user_id}")
    ApiResponse<ArticleRecommendation> getUserRecommendations(@PathVariable("user_id") String userId, @RequestParam("count") Integer size);
}