package com.grtsinry43.grtblog.config;

import com.grtsinry43.grtblog.security.*;
import com.grtsinry43.grtblog.service.CustomOAuth2UserService;
import jakarta.annotation.Resource;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.security.authentication.AuthenticationManager;
import org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration;
import org.springframework.security.config.annotation.method.configuration.EnableMethodSecurity;
import org.springframework.security.config.annotation.web.builders.HttpSecurity;
import org.springframework.security.config.annotation.web.configuration.EnableWebSecurity;
import org.springframework.security.config.http.SessionCreationPolicy;
import org.springframework.security.crypto.bcrypt.BCryptPasswordEncoder;
import org.springframework.security.crypto.password.PasswordEncoder;
import org.springframework.security.config.annotation.web.configurers.AbstractHttpConfigurer;
import org.springframework.security.web.SecurityFilterChain;
import org.springframework.security.web.authentication.UsernamePasswordAuthenticationFilter;

/**
 * @author grtsinry43
 * @date 2024/11/9 12:51
 * @description 热爱可抵岁月漫长
 */
@Configuration
@EnableWebSecurity
@EnableMethodSecurity(securedEnabled = true)
public class SecurityConfig {
    // 自定义认证管理器
    @Bean
    public AuthenticationManager authenticationManager(AuthenticationConfiguration authenticationConfiguration) throws Exception {
        return authenticationConfiguration.getAuthenticationManager();
    }

    // 自定义 jwt 过滤器，用于解析每次请求的 token
    @Resource
    private JwtAuthenticationFilter jwtAuthenticationFilter;

    // 自定义 OAuth2 用户服务
    @Resource
    private CustomOAuth2UserService customOAuth2UserService;

    // 自定义登录失败处理器
    @Resource
    private LoginUnAuthenticationEntryPointHandler loginUnAuthenticationEntryPointHandler;

    // 自定义权限不足处理器
    @Resource
    private LoginUnAccessDeniedHandler loginUnAccessDeniedHandler;

    @Resource
    private LogoutStatusSuccessHandler logoutStatusSuccessHandler;

    @Resource
    private OAuth2LoginSuccessHandler oAuth2LoginSuccessHandler;

    // 自定义密码加密器
    @Bean
    public PasswordEncoder generalPasswordEncoder() {
        return new BCryptPasswordEncoder();
    }

    // 设置过滤器链
    @Bean
    public SecurityFilterChain filterChain(HttpSecurity http) throws Exception {
        http.csrf(AbstractHttpConfigurer::disable) // 防止跨站请求伪造
                .sessionManagement(sessionManagement -> sessionManagement.sessionCreationPolicy(SessionCreationPolicy.IF_REQUIRED)) // 取消 session
                .authorizeHttpRequests(authorizeRequests -> authorizeRequests
                        .requestMatchers("/user/login", "/user/logout", "/oauth2/authorization/**", "/login**", "/captcha", "/admin/login", "/page-view", "/page/**", "/search/**", "/websiteInfo/**", "/uploads/**", "/friendLink", "/feed", "/notification", "/plugins/**", "/onlineStatus", "/overview", "/recommend", "/user-behavior/**", "/thinking/**").permitAll() // 登录和未登录的人都可以访问
                        // 这里感觉还是把对外的接口放在统一的路径，而需要权限认证和登录的或者管理员接口每个都需要校验
                        .requestMatchers("/article/**", "/statusUpdate/**", "/comment/**", "/tag/**", "/nav/**", "/category/**", "/archive").permitAll()
                        // TODO : 仅开发使用
                        .requestMatchers("/doc.html", "/swagger-resources/**", "/webjars/**", "/swagger-ui*/**", "/swagger-ui*/*swagger-initializer.js", "/test/sendEmail").permitAll() // 允许访问 Knife4j 和 Swagger 相关的端点
                        .anyRequest().authenticated() // 除了上面设置的地址可以匿名访问, 其它所有的请求地址需要认证访问
                )
                .oauth2Login(oauth2Login -> oauth2Login
                        .userInfoEndpoint(userInfoEndpoint -> userInfoEndpoint
                                .userService(customOAuth2UserService)
                        )
                        .successHandler(oAuth2LoginSuccessHandler)
                );
        // 自定义每次请求的过滤器
        http.addFilterBefore(jwtAuthenticationFilter, UsernamePasswordAuthenticationFilter.class);
        // 自定义未认证处理器
        http.exceptionHandling(exceptionHandling -> exceptionHandling.authenticationEntryPoint(loginUnAuthenticationEntryPointHandler));
        // 注册自定义的处理器 (认证后的用户访问需要认证资源时因为权限不足走的处理器)
        http.exceptionHandling(exceptionHandling -> exceptionHandling.accessDeniedHandler(loginUnAccessDeniedHandler));
        // 自定义登出成功处理器
        http.logout(logout -> logout.logoutUrl("/user/logout").logoutSuccessHandler(logoutStatusSuccessHandler));
        return http.build();
    }
}
