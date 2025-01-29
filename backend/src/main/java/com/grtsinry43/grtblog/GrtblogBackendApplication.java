package com.grtsinry43.grtblog;

import org.mybatis.spring.annotation.MapperScan;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.cloud.openfeign.EnableFeignClients;
import org.springframework.scheduling.annotation.EnableAsync;
import org.springframework.scheduling.annotation.EnableScheduling;
import org.springframework.security.crypto.bcrypt.BCryptPasswordEncoder;
import lombok.extern.slf4j.Slf4j;
import org.springframework.web.bind.annotation.CrossOrigin;

@Slf4j
@SpringBootApplication
@EnableScheduling
@EnableFeignClients
@EnableAsync
@MapperScan("com.grtsinry43.grtblog.mapper")
public class GrtblogBackendApplication {

    public static void main(String[] args) {
        System.out.println(System.getProperty("user.dir"));
        SpringApplication.run(GrtblogBackendApplication.class, args);
    }

}
