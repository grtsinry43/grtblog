package com.grtsinry43.grtblog.config;

import io.swagger.v3.oas.annotations.info.License;
import io.swagger.v3.oas.models.OpenAPI;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.web.servlet.config.annotation.ResourceHandlerRegistry;
import org.springframework.web.servlet.config.annotation.WebMvcConfigurationSupport;

/**
 * @author grtsinry43
 * @date 2025/3/4 12:34
 * @description 热爱可抵岁月漫长
 */

@Configuration
public class Knife4jConfig extends WebMvcConfigurationSupport {

    /**
     * 设置静态资源映射
     *
     * @param registry
     */
    protected void addResourceHandlers (ResourceHandlerRegistry registry) {
        registry.addResourceHandler ("/doc.html").addResourceLocations ("classpath:/META-INF/resources/");
        registry.addResourceHandler ("/webjars/**").addResourceLocations ("classpath:/META-INF/resources/webjars/");
    }
}
