package com.grtsinry43.grtblog.controller;

import com.grtsinry43.grtblog.dto.ApiResponse;
import com.grtsinry43.grtblog.service.OnlineStatusService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.security.access.prepost.PreAuthorize;
import org.springframework.web.bind.annotation.*;

import java.util.Map;

/**
 * @author grtsinry43
 * @date 2025/1/26 23:20
 * @description 热爱可抵岁月漫长
 */
@RestController
@RequestMapping("/onlineStatus")
public class OnlineStatusController {

    private final OnlineStatusService onlineStatusService;

    public OnlineStatusController(OnlineStatusService onlineStatusService) {
        this.onlineStatusService = onlineStatusService;
    }

    @PostMapping
    @PreAuthorize("hasAuthority('article:add')")
    public ApiResponse<String> updateStatus(@RequestBody Map<String, Object> status) {
        System.out.println(status);
        onlineStatusService.updateStatus(status);
        return ApiResponse.success();
    }

    @GetMapping
    public ApiResponse<Map<String, Object>> getStatus() {
        return ApiResponse.success(onlineStatusService.getStatus());
    }
}