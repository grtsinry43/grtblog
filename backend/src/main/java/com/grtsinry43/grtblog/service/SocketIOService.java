package com.grtsinry43.grtblog.service;

import com.corundumstudio.socketio.SocketIOClient;
import com.corundumstudio.socketio.SocketIOServer;
import com.corundumstudio.socketio.annotation.OnConnect;
import com.corundumstudio.socketio.annotation.OnDisconnect;
import com.corundumstudio.socketio.annotation.OnEvent;
import com.grtsinry43.grtblog.util.PageMatcher;
import jakarta.annotation.PostConstruct;
import jakarta.annotation.PreDestroy;
import lombok.Getter;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.net.SocketAddress;
import java.util.Set;
import java.util.UUID;
import java.util.concurrent.*;

/**
 * @author grtsinry43
 * @date 2024/11/19 17:08
 * @description 热爱可抵岁月漫长
 */
@Service
public class SocketIOService {
    private final SocketIOServer socketIOServer;
    @Getter
    private final ConcurrentHashMap<String, Set<UUID>> pageUserMap = new ConcurrentHashMap<>();
    private final ConcurrentHashMap<UUID, String> clientPageMap = new ConcurrentHashMap<>();
    private final ScheduledExecutorService scheduler = Executors.newScheduledThreadPool(1);
    private final ConcurrentHashMap<String, ScheduledFuture<?>> debounceTasks = new ConcurrentHashMap<>();
    private final PageMatcher pageMatcher;

    public void broadcastNotification(String message) {
        socketIOServer.getBroadcastOperations().sendEvent("updateNotification", message);
    }

    public void broadcastMessage(String eventName, Object message) {
        socketIOServer.getBroadcastOperations().sendEvent(eventName, message);
    }

    @Autowired
    public SocketIOService(SocketIOServer socketIOServer, PageMatcher pageMatcher) {
        this.socketIOServer = socketIOServer;
        this.pageMatcher = pageMatcher;
    }

    @PostConstruct
    public void startServer() {
        try {
            socketIOServer.start();
            System.out.println("Socket.IO server started");
        } catch (Exception e) {
            System.err.println("Failed to start Socket.IO server: " + e.getMessage());
        }
    }

    @OnConnect
    public void onConnect(SocketIOClient client) {
        debounceUpdateTotalOnlineCount();
    }

    @OnDisconnect
    public void onDisconnect(SocketIOClient client) {
        UUID clientId = client.getSessionId();
        String page = clientPageMap.remove(clientId);

        if (page != null) {
            Set<UUID> users = pageUserMap.getOrDefault(page, ConcurrentHashMap.newKeySet());
            users.remove(clientId);
            if (users.isEmpty()) {
                pageUserMap.remove(page);
            } else {
                pageUserMap.put(page, users);
            }
            debounceUpdatePageViewCount(page);
        }

        debounceUpdateTotalOnlineCount();
    }

    @OnEvent("enterPage")
    public void onEnterPage(SocketIOClient client, String page) {
        UUID clientId = client.getSessionId();
        SocketAddress remoteAddress = client.getRemoteAddress();
        String pageName = pageMatcher.matchPath(page, remoteAddress);
        String previousPage = clientPageMap.put(clientId, pageName);

        if (previousPage != null && !previousPage.equals(pageName)) {
            Set<UUID> previousUsers = pageUserMap.getOrDefault(previousPage, ConcurrentHashMap.newKeySet());
            previousUsers.remove(clientId);
            if (previousUsers.isEmpty()) {
                pageUserMap.remove(previousPage);
            } else {
                pageUserMap.put(previousPage, previousUsers);
            }
            debounceUpdatePageViewCount(previousPage);
        }

        pageUserMap.computeIfAbsent(pageName, k -> ConcurrentHashMap.newKeySet()).add(clientId);
        debounceUpdatePageViewCount(pageName);
        debounceUpdateTotalOnlineCount();
    }

    @PreDestroy
    public void stopServer() {
        try {
            socketIOServer.stop();
            scheduler.shutdown();
            System.out.println("Socket.IO server stopped");
        } catch (Exception e) {
            System.err.println("Failed to stop Socket.IO server: " + e.getMessage());
        }
    }

    private void debounceUpdateTotalOnlineCount() {
        String key = "totalOnlineCount";
        if (debounceTasks.containsKey(key)) {
            debounceTasks.get(key).cancel(false);
        }
        ScheduledFuture<?> future = scheduler.schedule(this::updateTotalOnlineCount, 2000, TimeUnit.MILLISECONDS);
        debounceTasks.put(key, future);
    }

    private void updateTotalOnlineCount() {
        int totalOnlineCount = pageUserMap.values().stream()
                .mapToInt(Set::size)
                .sum();
        socketIOServer.getBroadcastOperations().sendEvent("totalOnlineCount", totalOnlineCount);
    }

    private void debounceUpdatePageViewCount(String page) {
        if (debounceTasks.containsKey(page)) {
            debounceTasks.get(page).cancel(false);
        }
        ScheduledFuture<?> future = scheduler.schedule(() -> updatePageViewCount(page), 2000, TimeUnit.MILLISECONDS);
        debounceTasks.put(page, future);
    }

    private void updatePageViewCount(String page) {
        int count = pageUserMap.getOrDefault(page, ConcurrentHashMap.newKeySet()).size();
        socketIOServer.getBroadcastOperations().sendEvent("pageViewCount", page, count);
    }
}