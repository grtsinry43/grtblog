package com.grtsinry43.grtblog.controller;

import com.grtsinry43.grtblog.entity.User;
import com.grtsinry43.grtblog.entity.UserBehavior;
import com.grtsinry43.grtblog.service.UserBehaviorService;
import com.grtsinry43.grtblog.util.SecurityUtils;
import jakarta.servlet.http.HttpServletRequest;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.time.LocalDateTime;
import java.time.temporal.ChronoUnit;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.Executors;
import java.util.concurrent.ScheduledExecutorService;
import java.util.concurrent.TimeUnit;

/**
 * @author grtsinry43
 * @date 2024/11/26 14:39
 * @description 热爱可抵岁月漫长
 */
@RestController
@RequestMapping("/user-behavior")
public class UserBehaviorController {
    private final UserBehaviorService userBehaviorService;
    private final ConcurrentHashMap<String, LocalDateTime> userArticleTimestamps = new ConcurrentHashMap<>();
    private final ConcurrentHashMap<String, Integer> userArticleReadTime = new ConcurrentHashMap<>();
    private final ScheduledExecutorService scheduler = Executors.newScheduledThreadPool(1);

    public UserBehaviorController(UserBehaviorService userBehaviorService) {
        this.userBehaviorService = userBehaviorService;
    }

    @GetMapping("/read/{articleId}")
    public void addReadBehavior(@PathVariable String articleId, HttpServletRequest request) {
        User currentUser = SecurityUtils.getCurrentUser();
        String userId = currentUser == null ? request.getSession().getId() : currentUser.getId().toString();
        String key = userId + ":" + articleId;
        LocalDateTime now = LocalDateTime.now();

        userArticleTimestamps.compute(key, (k, lastReadTime) -> {
            if (lastReadTime == null || ChronoUnit.MINUTES.between(lastReadTime, now) >= 5) {
                UserBehavior userBehavior = new UserBehavior();
                userBehavior.setUserId(userId);
                userBehavior.setArticleId(articleId);
                userBehavior.setType("1");
                userBehavior.setDate(now);
                userBehaviorService.save(userBehavior);
                return now;
            }
            return lastReadTime;
        });

        userArticleReadTime.merge(key, 5, Integer::sum);

        scheduler.schedule(() -> {
            Integer readTime = userArticleReadTime.remove(key);
            if (readTime != null) {
                UserBehavior userBehavior = new UserBehavior();
                userBehavior.setUserId(userId);
                userBehavior.setArticleId(articleId);
                userBehavior.setType("4");
                userBehavior.setValue(readTime);
                userBehavior.setDate(LocalDateTime.now());
                userBehaviorService.save(userBehavior);
            }
        }, 10, TimeUnit.SECONDS);
    }
}