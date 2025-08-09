package com.grtsinry43.grtblog.service;

import com.fasterxml.jackson.annotation.JsonProperty;
import com.grtsinry43.grtblog.entity.Article;
import com.grtsinry43.grtblog.entity.StatusUpdate;
import com.grtsinry43.grtblog.service.impl.ArticleServiceImpl;
import com.grtsinry43.grtblog.service.impl.CategoryServiceImpl;
import com.grtsinry43.grtblog.service.impl.StatusUpdateServiceImpl;
import lombok.Data;
import lombok.Getter;
import lombok.Setter;
import org.springframework.stereotype.Service;

import java.time.LocalDateTime;
import java.time.Year;
import java.time.format.DateTimeFormatter;
import java.util.*;
import java.util.stream.Collectors;

/**
 * @author grtsinry43
 * @date 2024/11/17 09:57
 * @description 热爱可抵岁月漫长
 */
@Service
public class ArchiveService {
    private final ArticleServiceImpl articleService;
    private final StatusUpdateServiceImpl statusUpdateService;
    private final CategoryServiceImpl categoryService;

    public ArchiveService(ArticleServiceImpl articleService, StatusUpdateServiceImpl statusUpdateService, CategoryServiceImpl categoryService) {
        this.articleService = articleService;
        this.statusUpdateService = statusUpdateService;
        this.categoryService = categoryService;
    }

    /**
     * 获取总体的归档，按年份，综合所有的文章和说说，按时间倒序，加上文章和说说的数量，还有其对应类别
     */
    public Map<Year, ArchiveSummary> getArchive() {
        List<Archive<Article>> articles = getArticleArchives();
        List<Archive<StatusUpdate>> statusUpdates = getStatusUpdateArchives();

        Map<Year, ArchiveSummary> ArchiveMap = new HashMap<>();

        articles.forEach(article -> {
            Year year = Year.of(article.getCreatedAt().getYear());
            ArchiveMap.computeIfAbsent(year, k -> new ArchiveSummary()).addArticle(article);
        });

        statusUpdates.forEach(statusUpdate -> {
            Year year = Year.of(statusUpdate.getCreatedAt().getYear());
            ArchiveMap.computeIfAbsent(year, k -> new ArchiveSummary()).addStatusUpdate(statusUpdate);
        });

        return ArchiveMap.entrySet().stream()
                .sorted(Map.Entry.<Year, ArchiveSummary>comparingByKey().reversed())
                .collect(Collectors.toMap(
                        Map.Entry::getKey,
                        Map.Entry::getValue,
                        (e1, e2) -> e1,
                        LinkedHashMap::new
                ));
    }

    public List<Archive<Article>> getArticleArchives() {
        return articleService.list().stream()
                .filter(article -> article.getIsPublished() && Objects.isNull(article.getDeletedAt()))
                .map(this::convertToArticleArchive)
                .collect(Collectors.toList());
    }

    public List<Archive<StatusUpdate>> getStatusUpdateArchives() {
        return statusUpdateService.list().stream()
                .filter(statusUpdate -> statusUpdate.getIsPublished() && Objects.isNull(statusUpdate.getDeletedAt()))
                .map(this::convertToStatusUpdateArchive)
                .collect(Collectors.toList());
    }

    private <T> Archive<T> convertToArchive(T entity, String title, String shortUrl, Long categoryId, LocalDateTime createdAt) {
        Archive<T> archive = new Archive<>();
        archive.setTitle(title);
        if (entity instanceof StatusUpdate) {
            archive.setShortUrl(createdAt.format(DateTimeFormatter.ofPattern("yyyy/MM/dd/")) + shortUrl);
        } else {
            archive.setShortUrl(shortUrl);
        }
        archive.setCategory(categoryService.getById(categoryId) != null ? categoryService.getById(categoryId).getName() : "未分类");
        archive.setCreatedAt(createdAt);
        return archive;
    }

    private Archive<Article> convertToArticleArchive(Article article) {
        return convertToArchive(article, article.getTitle(), article.getShortUrl(), article.getCategoryId(), article.getCreatedAt());
    }

    private Archive<StatusUpdate> convertToStatusUpdateArchive(StatusUpdate statusUpdate) {
        return convertToArchive(statusUpdate, statusUpdate.getTitle(), statusUpdate.getShortUrl(), statusUpdate.getCategoryId(), statusUpdate.getCreatedAt());
    }

    @Getter
    @Setter
    public static class ArchiveSummary {
        private int articleCount;
        private int statusUpdateCount;
        private List<Archive<Article>> articles = new ArrayList<>();
        private List<Archive<StatusUpdate>> statusUpdates = new ArrayList<>();

        public void addArticle(Archive<Article> article) {
            articles.add(article);
            articleCount++;
        }

        public void addStatusUpdate(Archive<StatusUpdate> statusUpdate) {
            statusUpdates.add(statusUpdate);
            statusUpdateCount++;
        }
    }

    @Data
    public static class Archive<T> {
        private String title;
        private String shortUrl;
        private String category;
        private transient LocalDateTime createdAt;

        @JsonProperty("createdAt")
        public String getFormattedCreatedAt() {
            return createdAt.format(DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm:ss"));
        }
    }
}