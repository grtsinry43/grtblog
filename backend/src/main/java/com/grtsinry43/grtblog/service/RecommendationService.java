package com.grtsinry43.grtblog.service;

/**
 * @author grtsinry43
 * @date 2024/11/8 00:26
 * @description 热爱可抵岁月漫长
 */

import com.grtsinry43.grtblog.client.RecommenderClient;
import com.grtsinry43.grtblog.dto.ApiResponse;
import com.grtsinry43.grtblog.dto.ArticleRecommendUpdate;
import com.grtsinry43.grtblog.entity.Article;
import com.grtsinry43.grtblog.service.impl.ArticleServiceImpl;
import com.grtsinry43.grtblog.vo.ArticleRecommendation;
import org.springframework.stereotype.Service;

import java.util.List;

@Service
public class RecommendationService {

    private final RecommenderClient recommenderClient;

    public RecommendationService(RecommenderClient recommenderClient) {
        this.recommenderClient = recommenderClient;
    }

    public String[] getRecommendations(Long articleId, Integer count) {
        ApiResponse<ArticleRecommendation> res = recommenderClient.getRecommendations(articleId.toString(), count);
        if (res.getCode() == 0) {
            return res.getData().getRecommendation();
        } else {
            return new String[0];
        }
    }

    public void updateArticleStatus(Article article) {
        ArticleRecommendUpdate articleRecommendUpdate = new ArticleRecommendUpdate();
        articleRecommendUpdate.setId(article.getId().toString());
        articleRecommendUpdate.setTitle(article.getTitle());
        articleRecommendUpdate.setContent(article.getContent());
        ApiResponse<Object> objectApiResponse = recommenderClient.updateArticleStatus(article.getId().toString(), articleRecommendUpdate);
        if (objectApiResponse.getCode() != 0) {
            throw new RuntimeException("更新文章状态失败");
        }
    }

    public void deleteArticleStatus(Long articleId) {
        ApiResponse<Object> objectApiResponse = recommenderClient.deleteArticle(articleId.toString());
        if (objectApiResponse.getCode() != 0 && objectApiResponse.getCode() != 404) {
            throw new RuntimeException("删除文章失败");
        }
    }

    public String[] getUserRecommendations(String userId, Integer count) {
        ApiResponse<ArticleRecommendation> res = recommenderClient.getUserRecommendations(userId, count);
        if (res.getCode() == 0) {
            return res.getData().getRecommendation();
        } else {
            return new String[0];
        }
    }
}

