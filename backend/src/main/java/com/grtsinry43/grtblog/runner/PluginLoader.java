//package com.grtsinry43.grtblog.runner;
//
//import com.grtsinry43.grtblog.plugin.BlogPlugin;
//import org.pf4j.PluginManager;
//import org.springframework.beans.factory.annotation.Autowired;
//import org.springframework.boot.CommandLineRunner;
//import org.springframework.stereotype.Component;
//
///**
// * @author grtsinry43
// * @date 2025/1/25 12:40
// * @description 热爱可抵岁月漫长
// */
//@Component
//public class PluginLoader implements CommandLineRunner {
//
//    @Autowired
//    private PluginManager pluginManager;
//
//    @Override
//    public void run(String... args) throws Exception {
//        pluginManager.loadPlugins();
//        pluginManager.startPlugins();
//
//        pluginManager.getExtensions(BlogPlugin.class).forEach(BlogPlugin::apply);
//    }
//}