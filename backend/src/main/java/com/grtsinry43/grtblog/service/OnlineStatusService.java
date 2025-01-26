package com.grtsinry43.grtblog.service;

import jakarta.annotation.PostConstruct;
import org.springframework.stereotype.Service;

import java.util.HashMap;
import java.util.Map;
import java.util.concurrent.Executors;
import java.util.concurrent.ScheduledExecutorService;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.locks.ReentrantLock;

/**
 * @author grtsinry43
 * @date 2025/1/26 23:25
 * @description 热爱可抵岁月漫长
 */
@Service
public class OnlineStatusService {
    private final Map<String, Object> cachedStatus = new HashMap<>();
    private final ReentrantLock lock = new ReentrantLock();
    private final SocketIOService socketIOService;
    private final ScheduledExecutorService scheduler = Executors.newScheduledThreadPool(1);
    private volatile long lastUpdateTime;

    public OnlineStatusService(SocketIOService socketIOService) {
        this.socketIOService = socketIOService;
    }

    @PostConstruct
    public void init() {
        // 初始数据
        cachedStatus.put("ok", 0); // 代表服务是否正常
        lastUpdateTime = System.currentTimeMillis();

        // 定时检查并重置状态（1 分钟未更新则认为服务不正常，也就是 ok = 0）
        scheduler.scheduleAtFixedRate(this::checkAndResetStatus, 1, 1, TimeUnit.MINUTES);
    }

    public void updateStatus(Map<String, Object> status) {
        lock.lock();
        try {
            cachedStatus.clear();
            cachedStatus.put("ok", 1);
            cachedStatus.putAll(status);
            lastUpdateTime = System.currentTimeMillis();
            broadcastStatus();
        } finally {
            lock.unlock();
        }
    }

    public Map<String, Object> getStatus() {
        lock.lock();
        try {
            return new HashMap<>(cachedStatus);
        } finally {
            lock.unlock();
        }
    }

    private void broadcastStatus() {
        // WebSocket 广播逻辑
        socketIOService.broadcastMessage("authorStatus", cachedStatus);
    }

    private void checkAndResetStatus() {
        lock.lock();
        try {
            if (System.currentTimeMillis() - lastUpdateTime >= TimeUnit.MINUTES.toMillis(5)) {
                cachedStatus.clear();
                cachedStatus.put("ok", 0);
                broadcastStatus();
            }
        } finally {
            lock.unlock();
        }
    }
}