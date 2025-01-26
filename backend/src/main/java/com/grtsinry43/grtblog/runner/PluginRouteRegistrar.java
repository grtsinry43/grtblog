package com.grtsinry43.grtblog.runner;

import com.grtblog.BlogPlugin;
import org.pf4j.spring.SpringPluginManager;
import org.springframework.beans.BeansException;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.context.ApplicationContext;
import org.springframework.context.ApplicationContextAware;
import org.springframework.context.ApplicationListener;
import org.springframework.context.event.ContextRefreshedEvent;
import org.springframework.stereotype.Component;
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
    public void setApplicationContext(ApplicationContext applicationContext) throws BeansException {
        this.applicationContext = applicationContext;
    }

    public void refreshPluginRoutes() {
        System.out.println("===Context refreshed event received.===");
        pluginManager.getPlugins().forEach(plugin -> {
            System.out.println("Plugin loaded: " + plugin.getPluginId() + " - " + plugin.getPluginState());
        });

        System.out.println(pluginManager.getExtensions(BlogPlugin.class));

        pluginManager.getExtensions(BlogPlugin.class).forEach(extension -> {
            try {
                String endpoint = "/plugins" + extension.getEndpoint();
                if (!registeredEndpoints.contains(endpoint)) { // Check if already registered
                    System.out.println("Registering endpoint: " + endpoint);
                    Method handleRequestMethod = extension.getClass().getMethod("handleRequest");
                    requestMappingHandlerMapping.registerMapping(
                            RequestMappingInfo.paths(endpoint).methods(RequestMethod.GET).build(),
                            extension,
                            handleRequestMethod
                    );
                    registeredEndpoints.add(endpoint); // Add to registered set
                } else {
                    System.out.println("Endpoint " + endpoint + " already registered. Skipping.");
                }
            } catch (NoSuchMethodException e) {
                System.err.println("Method handleRequest not found for extension: " + extension.getClass().getName());
                e.printStackTrace(); // Important for debugging
            } catch (Exception e) { // Catch other potential exceptions during mapping
                System.err.println("Error registering endpoint: " + e.getMessage());
                e.printStackTrace();
            }
        });
    }

    public void unregisterPluginRoutes() {
        registeredEndpoints.forEach(endpoint -> {
            System.out.println("Unregistering endpoint: " + endpoint);
            requestMappingHandlerMapping.unregisterMapping(RequestMappingInfo.paths(endpoint).methods(RequestMethod.GET).build());
        });
        registeredEndpoints.clear();
    }

    @Override
    public void onApplicationEvent(ContextRefreshedEvent event) {
        refreshPluginRoutes();
    }
}