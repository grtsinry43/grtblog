package com.grtsinry43.grtblog.runner;

import com.grtsinry43.grtblog.plugin.BlogPlugin;
import org.pf4j.PluginManager;
import org.pf4j.PluginState;
import org.pf4j.PluginStateEvent;
import org.pf4j.PluginStateListener;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;
import org.springframework.web.servlet.mvc.method.RequestMappingInfo;
import org.springframework.web.servlet.mvc.method.annotation.RequestMappingHandlerMapping;

import java.lang.reflect.Method;

/**
 * @author grtsinry43
 * @date 2025/1/25 12:57
 * @description 动态注册插件的 API
 */
@Component
public class PluginApiRegistrar implements PluginStateListener {

    @Autowired
    private PluginManager pluginManager;

    @Autowired
    private RequestMappingHandlerMapping handlerMapping;

    @Override
    public void pluginStateChanged(PluginStateEvent event) {
        System.out.println(event);
        if (event.getPluginState() == PluginState.STARTED) {
            pluginManager.getExtensions(BlogPlugin.class).forEach(extension -> {
                try {
                    // 获取插件的 API 路径和处理方法
                    System.out.println("===Registering API...===");
                    String endpoint = "/plugins" + extension.getEndpoint();
                    System.out.println("Registering API: " + endpoint);
                    System.out.println("Extension: " + extension.getClass().getSimpleName());
                    Method handleRequestMethod = extension.getClass().getMethod("handleRequest");
                    System.out.println("Method: " + handleRequestMethod.getName());

                    // 注册 API
                    handlerMapping.registerMapping(
                            RequestMappingInfo.paths(endpoint).methods(org.springframework.web.bind.annotation.RequestMethod.GET).build(),
                            extension,
                            handleRequestMethod
                    );
                    System.out.println("API registered.");
                } catch (NoSuchMethodException e) {
                    e.printStackTrace();
                }
            });
        }
    }
}
