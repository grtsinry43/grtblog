package com.grtsinry43.grtblog.service;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.grtsinry43.grtblog.entity.Article;
import com.grtsinry43.grtblog.entity.Page;
import com.grtsinry43.grtblog.entity.StatusUpdate;
import org.elasticsearch.action.delete.DeleteRequest;
import org.elasticsearch.action.index.IndexRequest;
import org.elasticsearch.action.update.UpdateRequest;
import org.elasticsearch.client.RequestOptions;
import org.elasticsearch.client.RestHighLevelClient;
import org.elasticsearch.common.xcontent.XContentType;
import org.elasticsearch.index.query.QueryBuilders;
import org.elasticsearch.search.builder.SearchSourceBuilder;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.io.IOException;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;

@Service
public class ElasticsearchService {

    private final RestHighLevelClient client;
    private final ObjectMapper objectMapper;

    @Autowired
    public ElasticsearchService(RestHighLevelClient client, ObjectMapper objectMapper) {
        this.client = client;
        this.objectMapper = objectMapper;
    }

    public void indexArticle(Article article) throws IOException {
        IndexRequest indexRequest = new IndexRequest("articles")
                .id(article.getId().toString())
                .source(objectMapper.writeValueAsString(article), XContentType.JSON);
        client.index(indexRequest, RequestOptions.DEFAULT);
    }

    public void updateArticle(Article article) throws IOException {
        UpdateRequest updateRequest = new UpdateRequest("articles", article.getId().toString())
                .doc(objectMapper.writeValueAsString(article), XContentType.JSON);
        client.update(updateRequest, RequestOptions.DEFAULT);
    }

    public void deleteArticle(String articleId) throws IOException {
        client.delete(new DeleteRequest("articles", articleId), RequestOptions.DEFAULT);
    }

    public void indexPage(Page page) throws IOException {
        IndexRequest indexRequest = new IndexRequest("pages")
                .id(page.getId().toString())
                .source(objectMapper.writeValueAsString(page), XContentType.JSON);
        client.index(indexRequest, RequestOptions.DEFAULT);
    }

    public void updatePage(Page page) throws IOException {
        UpdateRequest updateRequest = new UpdateRequest("pages", page.getId().toString())
                .doc(objectMapper.writeValueAsString(page), XContentType.JSON);
        client.update(updateRequest, RequestOptions.DEFAULT);
    }

    public void deletePage(String pageId) throws IOException {
        client.delete(new DeleteRequest("pages", pageId), RequestOptions.DEFAULT);
    }

    public void indexStatusUpdate(StatusUpdate statusUpdate) throws IOException {
        IndexRequest indexRequest = new IndexRequest("status_updates")
                .id(statusUpdate.getId().toString())
                .source(objectMapper.writeValueAsString(statusUpdate), XContentType.JSON);
        client.index(indexRequest, RequestOptions.DEFAULT);
    }

    public void updateStatusUpdate(StatusUpdate statusUpdate) throws IOException {
        UpdateRequest updateRequest = new UpdateRequest("status_updates", statusUpdate.getId().toString())
                .doc(objectMapper.writeValueAsString(statusUpdate), XContentType.JSON);
        client.update(updateRequest, RequestOptions.DEFAULT);
    }

    public void deleteStatusUpdate(String statusUpdateId) throws IOException {
        client.delete(new DeleteRequest("status_updates", statusUpdateId), RequestOptions.DEFAULT);
    }

    public List<Map<String, Object>> searchArticles(SearchSourceBuilder searchSourceBuilder) throws IOException {
        return search("articles", searchSourceBuilder);
    }

    public List<Map<String, Object>> searchPages(SearchSourceBuilder searchSourceBuilder) throws IOException {
        return search("pages", searchSourceBuilder);
    }

    public List<Map<String, Object>> searchStatusUpdates(SearchSourceBuilder searchSourceBuilder) throws IOException {
        return search("status_updates", searchSourceBuilder);
    }

    private List<Map<String, Object>> search(String index, SearchSourceBuilder searchSourceBuilder) throws IOException {
        var searchRequest = new org.elasticsearch.action.search.SearchRequest(index);
        searchRequest.source(searchSourceBuilder);
        var searchResponse = client.search(searchRequest, RequestOptions.DEFAULT);
        var searchHits = searchResponse.getHits().getHits();
        List<Map<String, Object>> results = new ArrayList<>();
        for (var hit : searchHits) {
            results.add(hit.getSourceAsMap());
        }
        return results;
    }

    public List<Map<String, Object>> searchAll(String query) throws IOException {
        SearchSourceBuilder searchSourceBuilder = new SearchSourceBuilder()
                .query(QueryBuilders.multiMatchQuery(query, "title", "content"));
        List<Map<String, Object>> articles = searchArticles(searchSourceBuilder);
        List<Map<String, Object>> pages = searchPages(searchSourceBuilder);
        List<Map<String, Object>> statusUpdates = searchStatusUpdates(searchSourceBuilder);
        List<Map<String, Object>> allResults = new ArrayList<>();
        allResults.addAll(articles);
        allResults.addAll(pages);
        allResults.addAll(statusUpdates);
        return allResults;
    }
}
