package com.grtsinry43.grtblog.security;


import com.auth0.jwt.JWT;
import com.auth0.jwt.algorithms.Algorithm;
import com.auth0.jwt.exceptions.JWTVerificationException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.grtsinry43.grtblog.common.ErrorCode;
import com.grtsinry43.grtblog.dto.ApiResponse;
import com.grtsinry43.grtblog.entity.AdminToken;
import com.grtsinry43.grtblog.exception.BusinessException;
import com.grtsinry43.grtblog.service.AdminTokenService;
import com.grtsinry43.grtblog.service.UserDetailsServiceImpl;
import com.grtsinry43.grtblog.util.TokenGenerator;
import jakarta.annotation.Resource;
import jakarta.servlet.FilterChain;
import jakarta.servlet.ServletException;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.security.authentication.UsernamePasswordAuthenticationToken;
import org.springframework.security.core.context.SecurityContextHolder;
import org.springframework.security.core.userdetails.UserDetails;
import org.springframework.security.web.authentication.WebAuthenticationDetailsSource;
import org.springframework.stereotype.Component;
import org.springframework.web.filter.OncePerRequestFilter;

import java.io.IOException;
import java.time.LocalDateTime;

/**
 * @author grtsinry43
 * @date 2024/10/10 16:39
 * @description 热爱可抵岁月漫长
 */

@Component
public class JwtAuthenticationFilter extends OncePerRequestFilter {

    @Value("${com.grtsinry43.secret_key}")
    private String SECRET_KEY;

    @Resource
    private UserDetailsServiceImpl userDetailsService;

    @Resource
    private AdminTokenService adminTokenService;

    @Override
    protected void doFilterInternal(HttpServletRequest request, HttpServletResponse response, FilterChain filterChain)
            throws ServletException, IOException {
        String token = request.getHeader("Authorization");
        if (token != null && token.startsWith("Bearer ")) {
            token = token.substring(7);
            String username = null;
            try {
                username = JWT.require(Algorithm.HMAC512(SECRET_KEY)).build().verify(token).getSubject();
            } catch (JWTVerificationException e) {
                System.out.println("Token 验证失败");
            }
            if (username != null) {
                UserDetails userDetails = userDetailsService.loadUserByUsername(username);
                UsernamePasswordAuthenticationToken authentication = new UsernamePasswordAuthenticationToken(userDetails, null, userDetails.getAuthorities());
                authentication.setDetails(new WebAuthenticationDetailsSource().buildDetails(request));
                SecurityContextHolder.getContext().setAuthentication(authentication);
            }
        } else if (token != null && token.startsWith("gb_tk_")) {
            token = token.substring(6);
            String userId = null;
            String hashedToken = TokenGenerator.calculateHashedToken(token);
            AdminToken tokenMatched = adminTokenService.lambdaQuery().eq(AdminToken::getToken, hashedToken)
                    .gt(AdminToken::getExpireAt, LocalDateTime.now())
                    .one();
            if (tokenMatched != null) {
                userId = tokenMatched.getUserId().toString();
            }
            if (userId != null) {
                UserDetails userDetails = userDetailsService.loadUserByUserId(userId);
                UsernamePasswordAuthenticationToken authentication = new UsernamePasswordAuthenticationToken(userDetails, null, userDetails.getAuthorities());
                authentication.setDetails(new WebAuthenticationDetailsSource().buildDetails(request));
                SecurityContextHolder.getContext().setAuthentication(authentication);
            }

        }

        // 放行, 后面交给 Spring Security 框架
        // 交给你啦！.jpg（）
        filterChain.doFilter(request, response);
    }
}
