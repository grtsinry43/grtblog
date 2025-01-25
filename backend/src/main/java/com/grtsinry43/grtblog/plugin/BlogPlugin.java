package com.grtsinry43.grtblog.plugin;

import org.pf4j.ExtensionPoint;
import org.springframework.http.ResponseEntity;

/**
 * @author grtsinry43
 * @date 2025/1/25 12:38
 * @description 热爱可抵岁月漫长
 */
public interface BlogPlugin extends ExtensionPoint {
    /**
     * 用于插件的初始化（加载插件）
     */
    void apply();
    /**
     * 返回插件的 REST API 路径
     */
    String getEndpoint();

    /**
     * 处理请求逻辑
     */
    ResponseEntity<?> handleRequest();
}
