package com.grtsinry43.grtblog.service;

import com.baomidou.mybatisplus.core.conditions.query.QueryWrapper;
import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.grtsinry43.grtblog.entity.Article;
import com.grtsinry43.grtblog.entity.Page;
import com.grtsinry43.grtblog.entity.StatusUpdate;
import com.grtsinry43.grtblog.esdao.ArticleDocument;
import com.grtsinry43.grtblog.esdao.MomentDocument;
import com.grtsinry43.grtblog.esdao.PageDocument;
import com.grtsinry43.grtblog.mapper.ArticleMapper;
import com.grtsinry43.grtblog.mapper.PageMapper;
import com.grtsinry43.grtblog.mapper.StatusUpdateMapper;
import com.meilisearch.sdk.Client;
import com.meilisearch.sdk.Index;
import org.springframework.stereotype.Service;

import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;
import java.util.List;

@Service
public class MeiliDataSyncService {
    private final ArticleMapper articleMapper;
    private final StatusUpdateMapper statusUpdateMapper;
    private final PageMapper pageMapper;
    private final Client meiliSearchClient;

    public MeiliDataSyncService(ArticleMapper articleMapper, StatusUpdateMapper statusUpdateMapper, PageMapper pageMapper, Client meiliSearchClient) {
        this.articleMapper = articleMapper;
        this.statusUpdateMapper = statusUpdateMapper;
        this.pageMapper = pageMapper;
        this.meiliSearchClient = meiliSearchClient;
    }

    public void syncAllRecent() {
        List<Article> articles = articleMapper.selectList(new QueryWrapper<Article>().lambda().ge(Article::getUpdatedAt, LocalDateTime.now().minusMinutes(5)));
        List<ArticleDocument> articleDocuments = articles.stream()
                .filter(article -> article.getIsPublished() && article.getDeletedAt() == null)
                .map(article -> new ArticleDocument(article.getId(), article.getTitle(), article.getSummary(), article.getContent(), article.getShortUrl()))
                .toList();
        syncArticlesToMeiliSearch(articleDocuments);

        List<StatusUpdate> statusUpdates = statusUpdateMapper.selectList(new QueryWrapper<StatusUpdate>().lambda().ge(StatusUpdate::getUpdatedAt, LocalDateTime.now().minusMinutes(5)));
        List<MomentDocument> statusUpdateDocuments = statusUpdates.stream()
                .filter(statusUpdate -> statusUpdate.getIsPublished() && statusUpdate.getDeletedAt() == null)
                .map(statusUpdate -> new MomentDocument(statusUpdate.getId(), statusUpdate.getTitle(), statusUpdate.getSummary(), statusUpdate.getContent(), statusUpdate.getCreatedAt().format(DateTimeFormatter.ofPattern("yyyy/MM/dd/")) + statusUpdate.getShortUrl()))
                .toList();
        syncMomentsToMeiliSearch(statusUpdateDocuments);

        List<Page> pages = pageMapper.selectList(new QueryWrapper<Page>().lambda().ge(Page::getUpdatedAt, LocalDateTime.now().minusMinutes(5)));
        List<PageDocument> pageDocuments = pages.stream()
                .filter(page -> page.getDeletedAt() == null)
                .map(page -> new PageDocument(page.getId(), page.getTitle(), page.getDescription(), page.getContent(), page.getRefPath()))
                .toList();
        syncPagesToMeiliSearch(pageDocuments);
    }

    public void syncAll() {
        List<Article> articles = articleMapper.selectList(null);
        List<ArticleDocument> articleDocuments = articles.stream()
                .filter(article -> article.getIsPublished() && article.getDeletedAt() == null)
                .map(article -> new ArticleDocument(article.getId(), article.getTitle(), article.getSummary(), article.getContent(), article.getShortUrl()))
                .toList();
        syncArticlesToMeiliSearch(articleDocuments);

        List<StatusUpdate> statusUpdates = statusUpdateMapper.selectList(null);
        List<MomentDocument> statusUpdateDocuments = statusUpdates.stream()
                .filter(statusUpdate -> statusUpdate.getIsPublished() && statusUpdate.getDeletedAt() == null)
                .map(statusUpdate -> new MomentDocument(statusUpdate.getId(), statusUpdate.getTitle(), statusUpdate.getSummary(), statusUpdate.getContent(), statusUpdate.getCreatedAt().format(DateTimeFormatter.ofPattern("yyyy/MM/dd/")) + statusUpdate.getShortUrl()))
                .toList();
        syncMomentsToMeiliSearch(statusUpdateDocuments);

        List<Page> pages = pageMapper.selectList(null);
        List<PageDocument> pageDocuments = pages.stream()
                .filter(page -> page.getCanDelete() && page.getDeletedAt() == null && page.getEnable())
                .map(page -> new PageDocument(page.getId(), page.getTitle(), page.getDescription(), page.getContent(), page.getRefPath()))
                .toList();
        syncPagesToMeiliSearch(pageDocuments);
    }

    public void deleteContent(Long id, String type) {
        Index index = null;
        if ("article".equals(type)) {
            index = meiliSearchClient.index("articles");
            index.deleteDocumentsByFilter("id=" + id);
        } else if ("moment".equals(type)) {
            index = meiliSearchClient.index("moments");
            index.deleteDocumentsByFilter("id=" + id);
        } else if ("page".equals(type)) {
            index = meiliSearchClient.index("pages");
            index.deleteDocumentsByFilter("id=" + id);
        }
    }

    public void deleteAllContent() {
        Index articleIndex = meiliSearchClient.index("articles");
        articleIndex.deleteAllDocuments();
        Index momentIndex = meiliSearchClient.index("moments");
        momentIndex.deleteAllDocuments();
        Index pageIndex = meiliSearchClient.index("pages");
        pageIndex.deleteAllDocuments();
    }

    private void syncArticlesToMeiliSearch(List<ArticleDocument> articleDocuments) {
        Index index = meiliSearchClient.index("articles");
        ObjectMapper objectMapper = new ObjectMapper();
        try {
            String json = objectMapper.writeValueAsString(articleDocuments);
            index.addDocuments(json);
        } catch (JsonProcessingException e) {
            e.printStackTrace();
        }
    }

    private void syncMomentsToMeiliSearch(List<MomentDocument> momentDocuments) {
        Index index = meiliSearchClient.index("moments");
        ObjectMapper objectMapper = new ObjectMapper();
        try {
            String json = objectMapper.writeValueAsString(momentDocuments);
            index.addDocuments(json);
        } catch (JsonProcessingException e) {
            e.printStackTrace();
        }
    }

    private void syncPagesToMeiliSearch(List<PageDocument> pageDocuments) {
        Index index = meiliSearchClient.index("pages");
        ObjectMapper objectMapper = new ObjectMapper();
        try {
            String json = objectMapper.writeValueAsString(pageDocuments);
            index.addDocuments(json);
        } catch (JsonProcessingException e) {
            e.printStackTrace();
        }
    }
}