package com.grtsinry43.grtblog.config;

import com.grtsinry43.grtblog.runner.PluginApiRegistrar;
import org.pf4j.spring.SpringPluginManager;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

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

    @Bean
    public PluginApiRegistrar pluginApiRegistrar(SpringPluginManager pluginManager) {
        PluginApiRegistrar registrar = new PluginApiRegistrar();
        pluginManager.addPluginStateListener(registrar);
        return registrar;
    }
}
