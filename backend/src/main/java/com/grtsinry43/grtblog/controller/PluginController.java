package com.grtsinry43.grtblog.controller;

import com.grtblog.BlogPlugin;
import com.grtsinry43.grtblog.common.ErrorCode;
import com.grtsinry43.grtblog.dto.ApiResponse;
import com.grtsinry43.grtblog.exception.BusinessException;
import com.grtsinry43.grtblog.runner.PluginRouteRegistrar;
import com.grtsinry43.grtblog.vo.PluginItem;
import org.pf4j.PluginManager;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.multipart.MultipartFile;

import java.io.File;
import java.io.IOException;
import java.nio.file.Files;
import java.util.List;
import java.util.Objects;
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

    @Autowired
    private PluginRouteRegistrar pluginRouteRegistrar;


    private static final String PLUGINS_DIR = "plugins";
    private static final String DISABLED_FILE = "disabled.txt";

    @PostMapping("/load")
    public String loadPlugins() {
        pluginManager.loadPlugins();
        pluginManager.startPlugins();
        return "Plugins loaded and started.";
    }

    @PostMapping("/unload")
    public String unloadPlugins() {
        pluginManager.stopPlugins();
        pluginManager.unloadPlugins();
        return "Plugins stopped and unloaded.";
    }

    @PostMapping("/reload")
    public String reloadPlugins() {
        pluginManager.stopPlugins();
        pluginManager.unloadPlugins();
        pluginManager.loadPlugins();
        pluginManager.startPlugins();
        return "Plugins reloaded.";
    }

    @GetMapping
    public ApiResponse<List<PluginItem>> getPlugins() {
        List<PluginItem> plugins = pluginManager.getPlugins().stream().map(plugin -> {
            PluginItem item = new PluginItem();
            item.setPluginId(plugin.getDescriptor().getPluginId());
            item.setPluginDescription(plugin.getDescriptor().getPluginDescription());
            item.setPluginVersion(plugin.getDescriptor().getVersion());
            item.setPluginAuthor(plugin.getDescriptor().getProvider());
            item.setPluginClass(plugin.getClass().getName());
            item.setPluginState(plugin.getPluginState().toString());
            item.setPluginEndpoint(plugin.getDescriptor().getPluginId());
            return item;
        }).toList();
        return ApiResponse.success(plugins);
    }

    @PostMapping("/apply/{pluginId}")
    public ApiResponse<String> applyPlugin(@PathVariable String pluginId) {
        pluginManager.getPlugins().stream()
                .filter(plugin -> plugin.getDescriptor().getPluginId().equals(pluginId))
                .flatMap(plugin -> pluginManager.getExtensions(BlogPlugin.class).stream())
                .forEach(BlogPlugin::apply);
        return ApiResponse.success("Plugin " + pluginId + " initialization executed successfully.");
    }

    @PostMapping("/upload")
    public ApiResponse<String> uploadPlugin(@RequestParam("file") MultipartFile file) {
        if (file.isEmpty()) {
            throw new BusinessException(ErrorCode.PARAMS_ERROR);
        }

        File dest = new File(PLUGINS_DIR, Objects.requireNonNull(file.getOriginalFilename()));

        try {
            file.transferTo(dest);
            String pluginId = pluginManager.loadPlugin(dest.toPath());

            if (pluginManager.getPlugin(pluginId) != null) {
                pluginManager.stopPlugin(pluginId);
                pluginManager.unloadPlugin(pluginId);
            }

            pluginManager.loadPlugin(dest.toPath());
            pluginManager.startPlugin(pluginId);

            pluginRouteRegistrar.unregisterPluginRoutes();
            pluginRouteRegistrar.refreshPluginRoutes();

            return ApiResponse.success("Plugin " + file.getOriginalFilename() + " uploaded and replaced successfully.");
        } catch (IOException e) {
            e.printStackTrace();
            throw new BusinessException(ErrorCode.OPERATION_ERROR);
        }
    }

    @PostMapping("/disable/{pluginId}")
    public ApiResponse<String> disablePlugin(@PathVariable String pluginId) {
        try {
            File disabledFile = new File(PLUGINS_DIR, DISABLED_FILE);
            List<String> disabledPlugins = disabledFile.exists() ? Files.readAllLines(disabledFile.toPath()) : List.of();

            if (!disabledPlugins.contains(pluginId)) {
                Files.write(disabledFile.toPath(), (pluginId + System.lineSeparator()).getBytes(), java.nio.file.StandardOpenOption.CREATE, java.nio.file.StandardOpenOption.APPEND);
            }

            pluginManager.stopPlugin(pluginId);
            pluginRouteRegistrar.unregisterPluginRoutes();
            pluginRouteRegistrar.refreshPluginRoutes();
            return ApiResponse.success("Plugin " + pluginId + " disabled successfully.");
        } catch (IOException e) {
            e.printStackTrace();
            throw new BusinessException(ErrorCode.OPERATION_ERROR);
        }
    }

    @PostMapping("/enable/{pluginId}")
    public ApiResponse<String> enablePlugin(@PathVariable String pluginId) {
        try {
            File disabledFile = new File(PLUGINS_DIR, DISABLED_FILE);
            if (disabledFile.exists()) {
                List<String> disabledPlugins = Files.readAllLines(disabledFile.toPath());
                List<String> updatedPlugins = disabledPlugins.stream().filter(id -> !id.equals(pluginId)).collect(Collectors.toList());
                Files.write(disabledFile.toPath(), updatedPlugins);
            }

            pluginManager.startPlugin(pluginId);
            pluginRouteRegistrar.unregisterPluginRoutes();
            pluginRouteRegistrar.refreshPluginRoutes();
            return ApiResponse.success("Plugin " + pluginId + " enabled successfully.");
        } catch (IOException e) {
            e.printStackTrace();
            throw new BusinessException(ErrorCode.OPERATION_ERROR);
        }
    }

    @DeleteMapping("/delete/{pluginId}")
    public ApiResponse<String> deletePlugin(@PathVariable String pluginId) {
        try {
            // Find the plugin and get its path
            String pluginPath = pluginManager.getPlugins().stream()
                    .filter(plugin -> plugin.getDescriptor().getPluginId().equals(pluginId))
                    .findFirst()
                    .map(plugin -> plugin.getPluginPath().toString())
                    .orElseThrow(() -> new BusinessException(ErrorCode.PARAMS_ERROR));

            // Stop and unload the plugin
            pluginManager.stopPlugin(pluginId);
            pluginManager.unloadPlugin(pluginId);
            pluginRouteRegistrar.unregisterPluginRoutes();
            pluginRouteRegistrar.refreshPluginRoutes();

            // Delete the plugin file
            File pluginFile = new File(pluginPath);
            if (pluginFile.exists() && !pluginFile.delete()) {
                throw new BusinessException(ErrorCode.OPERATION_ERROR);
            }

            return ApiResponse.success("Plugin " + pluginId + " deleted successfully.");
        } catch (Exception e) {
            e.printStackTrace();
            throw new BusinessException(ErrorCode.OPERATION_ERROR);
        }
    }
}
