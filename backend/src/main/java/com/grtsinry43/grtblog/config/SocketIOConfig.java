package com.grtsinry43.grtblog.config;

import com.corundumstudio.socketio.SocketIOServer;
import com.corundumstudio.socketio.annotation.SpringAnnotationScanner;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

/**
 * @author grtsinry43
 * @date 2024/11/19 17:06
 * @description 热爱可抵岁月漫长
 */
@Configuration
public class SocketIOConfig {
    @Bean
    public SocketIOServer socketIOServer() {
        com.corundumstudio.socketio.Configuration config = new com.corundumstudio.socketio.Configuration();
        config.setHostname("0.0.0.0");  // 设置服务器主机名
        config.setPort(9091);              // 设置服务器端口

        // 启用 SSL(如果需要)：
        // config.setKeyStorePassword("password");
        // config.setKeyStore("classpath:keystore.jks");

        // 心跳配置
        config.setPingInterval(5000);  // 每 5 秒发送一次心跳
        config.setPingTimeout(10000);   // 超过 10 秒未收到心跳则断开连接

        return new SocketIOServer(config);
    }

    @Bean
    public SpringAnnotationScanner springAnnotationScanner(SocketIOServer socketIOServer) {
        return new SpringAnnotationScanner(socketIOServer);
    }
}
