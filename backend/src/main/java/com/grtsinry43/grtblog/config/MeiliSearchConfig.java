package com.grtsinry43.grtblog.config;

import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import com.meilisearch.sdk.Client;
import com.meilisearch.sdk.Config;

/**
 * @author grtsinry43
 * @date 2025/2/11 14:20
 * @description 热爱可抵岁月漫长
 */
@Configuration
public class MeiliSearchConfig {
    @Value("${com.grtsinry43.meilisearch.url}")
    private String url;

    @Value("${com.grtsinry43.meilisearch.apiKey}")
    private String apiKey;

    @Bean
    public Client meiliSearchClient() {
        return new Client(new Config(url, apiKey));
    }
}
