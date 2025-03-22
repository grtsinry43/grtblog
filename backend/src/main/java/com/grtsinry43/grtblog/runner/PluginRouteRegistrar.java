package com.grtsinry43.grtblog.runner;

import com.grtblog.BlogPlugin;
import org.jetbrains.annotations.NotNull;
import org.pf4j.spring.SpringPluginManager;
import org.springframework.beans.BeansException;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.context.ApplicationContext;
import org.springframework.context.ApplicationContextAware;
import org.springframework.context.ApplicationListener;
import org.springframework.context.event.ContextRefreshedEvent;
import org.springframework.stereotype.Component;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestMethod;
import org.springframework.web.servlet.mvc.method.RequestMappingInfo;
import org.springframework.web.servlet.mvc.method.annotation.RequestMappingHandlerMapping;

import java.lang.reflect.Method;
import java.util.HashSet;
import java.util.Set;

/**
 * @author grtsinry43
 * @date 2025/1/25 12:57
 * @description 动态注册插件的 API
 */
@Component
public class PluginRouteRegistrar implements ApplicationContextAware, ApplicationListener<ContextRefreshedEvent> {

    private final SpringPluginManager pluginManager;
    private final RequestMappingHandlerMapping requestMappingHandlerMapping;
    private ApplicationContext applicationContext;
    private final Set<String> registeredEndpoints = new HashSet<>();

    @Autowired
    public PluginRouteRegistrar(SpringPluginManager pluginManager, RequestMappingHandlerMapping requestMappingHandlerMapping) {
        this.pluginManager = pluginManager;
        this.requestMappingHandlerMapping = requestMappingHandlerMapping;
    }

    @Override
    public void setApplicationContext(@NotNull ApplicationContext applicationContext) {
        this.applicationContext = applicationContext;
    }

    public void refreshPluginRoutes() {
        pluginManager.getExtensions(BlogPlugin.class).forEach(extension -> {
            try {
//                String endpoint = "/plugins" + extension.getEndpoint();
//                if (!registeredEndpoints.contains(endpoint)) {
//                    Method handleRequestMethod = extension.getClass().getMethod("handleRequest");
//                    requestMappingHandlerMapping.registerMapping(
//                            RequestMappingInfo.paths(endpoint).methods(RequestMethod.GET).build(),
//                            extension,
//                            handleRequestMethod
//                    );
//                    registeredEndpoints.add(endpoint);
//                }

                // 动态注册插件的 JavaScript 组件端点
                String jsEndpoint = "/plugins" + extension.getEndpoint() + "/js";
                if (!registeredEndpoints.contains(jsEndpoint)) {
                    Method getJavaScriptComponentMethod = extension.getClass().getMethod("getJavaScriptContent");
                    requestMappingHandlerMapping.registerMapping(
                            RequestMappingInfo.paths(jsEndpoint).methods(RequestMethod.GET).build(),
                            extension,
                            getJavaScriptComponentMethod
                    );
                    registeredEndpoints.add(jsEndpoint);
                }

                pluginManager.getExtensions(BlogPlugin.class).forEach(extension1 -> {
                    for (Method method : extension1.getClass().getMethods()) {
                        if (method.isAnnotationPresent(RequestMapping.class)) {
                            RequestMapping requestMapping = method.getAnnotation(RequestMapping.class);
                            String[] paths = requestMapping.value();
                            RequestMethod[] methods = requestMapping.method();
                            for (String path : paths) {
                                String fullPath = "/plugins" + extension1.getEndpoint() + path;
                                if (!registeredEndpoints.contains(fullPath)) {
                                    requestMappingHandlerMapping.registerMapping(
                                            RequestMappingInfo.paths(fullPath).methods(methods).build(),
                                            extension1,
                                            method
                                    );
                                    registeredEndpoints.add(fullPath);
                                }
                            }
                        }
                    }
                });
                System.out.println("Registered endpoints: " + registeredEndpoints);
            } catch (NoSuchMethodException e) {
                e.printStackTrace();
            }
        });
    }

    public void unregisterPluginRoutes() {
        registeredEndpoints.forEach(endpoint -> {
            requestMappingHandlerMapping.unregisterMapping(RequestMappingInfo.paths(endpoint).methods(RequestMethod.GET).build());
        });
        registeredEndpoints.clear();
    }

    @Override
    public void onApplicationEvent(ContextRefreshedEvent event) {
        refreshPluginRoutes();
    }
}