package com.grtsinry43.grtblog.health;

import org.springframework.boot.actuate.health.Health;
import org.springframework.boot.actuate.health.HealthIndicator;
import org.springframework.stereotype.Component;

/**
 * @author grtsinry43
 * @date 2025/2/12 14:02
 * @description 热爱可抵岁月漫长
 */
@Component
public class CustomHealthIndicator implements HealthIndicator {

    @Override
    public Health health() {
        // 执行自定义的健康检查逻辑
        boolean isHealthy = checkHealth();
        if (isHealthy) {
            return Health.up().withDetail("Custom Health", "Service is healthy").build();
        } else {
            return Health.down().withDetail("Custom Health", "Service is down").build();
        }
    }

    private boolean checkHealth() {
        // 实现您的健康检查逻辑
        // 返回 true 表示健康，false 表示不健康
        return true;
    }
}
