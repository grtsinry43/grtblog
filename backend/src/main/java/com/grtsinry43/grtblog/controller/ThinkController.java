package com.grtsinry43.grtblog.controller;

import com.grtsinry43.grtblog.dto.ApiResponse;
import com.grtsinry43.grtblog.entity.Thinking;
import com.grtsinry43.grtblog.service.ThinkingService;
import com.grtsinry43.grtblog.vo.ThinkingVO;
import org.springframework.beans.BeanUtils;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.List;

/**
 * @author grtsinry43
 * @date 2025/1/31 21:43
 * @description 热爱可抵岁月漫长
 */
@RestController
@RequestMapping("/thinking")
public class ThinkController {
    private final ThinkingService thinkingService;

    public ThinkController(ThinkingService thinkingService) {
        this.thinkingService = thinkingService;
    }

    @GetMapping("/latest")
    public ApiResponse<ThinkingVO> getLatestThinkingApi() {
        Thinking latest = thinkingService.getLatest();
        ThinkingVO thinkingVO = new ThinkingVO();
        BeanUtils.copyProperties(latest, thinkingVO);
        thinkingVO.setId(latest.getId().toString());
        return ApiResponse.success(thinkingVO);
    }

    @GetMapping("/all")
    public ApiResponse<List<ThinkingVO>> getAllThinkingApi() {
        List<Thinking> thinkingList = thinkingService.listAll();
        List<ThinkingVO> thinkingVOS = thinkingList.stream()
                .map(thinking -> {
                    ThinkingVO thinkingVO = new ThinkingVO();
                    BeanUtils.copyProperties(thinking, thinkingVO);
                    thinkingVO.setId(thinking.getId().toString());
                    return thinkingVO;
                })
                .toList();
        return ApiResponse.success(thinkingVOS);
    }
}
