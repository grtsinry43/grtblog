package com.grtsinry43.grtblog.config;

import org.pf4j.PluginManager;
import org.pf4j.spring.SpringPluginManager;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.web.servlet.mvc.method.RequestMappingInfo;
import org.springframework.web.servlet.mvc.method.annotation.RequestMappingHandlerMapping;

import java.lang.reflect.Method;

/**
 * @author grtsinry43
 * @date 2025/1/25 12:39
 * @description 热爱可抵岁月漫长
 */
@Configuration
public class PluginConfig {

    @Bean
    public SpringPluginManager pluginManager() {
        return new SpringPluginManager();
    }

}
