package com.grtsinry43.grtblog.controller;

// import com.grtsinry43.grtblog.config.MyBatisPlusConfig;

import com.grtsinry43.grtblog.dto.ApiResponse;
import com.grtsinry43.grtblog.entity.Tag;
import com.grtsinry43.grtblog.service.impl.TagServiceImpl;
import com.grtsinry43.grtblog.vo.TagVO;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;
import lombok.extern.slf4j.Slf4j;
import org.springframework.validation.annotation.Validated;
import org.springframework.web.bind.annotation.*;

import java.util.List;

/**
 * 标签控制器
 *
 * @author grtsinry43
 * @since 2024-10-09
 */
@Slf4j
@Validated
@RestController
@RequestMapping("/tag")
public class TagController {
    private final TagServiceImpl tagService;

    public TagController(TagServiceImpl tagService) {
        this.tagService = tagService;
    }

    //    @AuthCheck(requiredRole = UserRole.ADMIN)
//    @PostMapping("/add")
//    public ApiResponse<Tag> addNewTagApi(@NotNull @NotBlank @RequestParam String tagName) {
//        Tag tag = tagService.addNewTag(tagName);
//        return ApiResponse.success(tag);
//    }

    @GetMapping
    public ApiResponse<List<TagVO>> listTagsApi() {
        return ApiResponse.success(tagService.getTagInfoList());
    }

    @GetMapping("/names")
    public ApiResponse<List<String>> listTagNamesApi() {
        return ApiResponse.success(tagService.getAllTagNames());
    }
}
