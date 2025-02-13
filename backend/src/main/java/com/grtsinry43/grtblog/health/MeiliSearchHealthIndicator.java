package com.grtsinry43.grtblog.health;

import com.meilisearch.sdk.Client;
import com.meilisearch.sdk.model.Stats;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.actuate.health.Health;
import org.springframework.boot.actuate.health.HealthIndicator;
import org.springframework.stereotype.Component;

/**
 * Author: grtsinry43
 * Date: 2025/2/12 14:02
 * Description: 热爱可抵岁月漫长
 */
@Component
public class MeiliSearchHealthIndicator implements HealthIndicator {

    private final Client meiliSearchClient;

    @Autowired
    public MeiliSearchHealthIndicator(Client meiliSearchClient) {
        this.meiliSearchClient = meiliSearchClient;
    }

    @Override
    public Health health() {
        // 执行自定义的健康检查逻辑
        boolean isHealthy = checkHealth();
        Stats stats = checkStatus();
        if (isHealthy) {
            return Health.up()
                    .withDetail("MeiliSearch Health", "Service is healthy")
                    .withDetail("Stats", stats)
                    .build();
        } else {
            return Health.down()
                    .withDetail("MeiliSearch Health", "Service is down")
                    .withDetail("Stats", stats)
                    .build();
        }
    }

    private boolean checkHealth() {
        return meiliSearchClient.isHealthy();
    }

    private Stats checkStatus() {
        return meiliSearchClient.getStats();
    }
}