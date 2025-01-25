package com.grtsinry43.grtblog.controller;

import com.grtblog.BlogPlugin;
import org.pf4j.PluginManager;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.stream.Collectors;

/**
 * @author grtsinry43
 * @date 2025/1/25 12:52
 * @description 热爱可抵岁月漫长
 */
@RestController
@RequestMapping("/api/plugins")
public class PluginController {
    @Autowired
    private PluginManager pluginManager;

    @PostMapping("/load")
    public String loadPlugins() {
        pluginManager.loadPlugins();
        pluginManager.startPlugins();
        return "Plugins loaded and started.";
    }

    @GetMapping
    public List<String> getPlugins() {
        System.out.println(pluginManager.getExtensions(BlogPlugin.class));
        return pluginManager.getExtensions(BlogPlugin.class)
                .stream()
                .map(plugin -> plugin.getClass().getSimpleName())
                .collect(Collectors.toList());
    }

    @PostMapping("/apply/{pluginName}")
    public String applyPlugin(@PathVariable String pluginName) {
        pluginManager.getExtensions(BlogPlugin.class)
                .stream()
                .filter(plugin -> plugin.getClass().getSimpleName().equals(pluginName))
                .forEach(BlogPlugin::apply);
        return "Plugin " + pluginName + " applied.";
    }
}