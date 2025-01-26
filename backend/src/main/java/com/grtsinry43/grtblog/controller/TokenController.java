package com.grtsinry43.grtblog.controller;

import com.grtsinry43.grtblog.common.ErrorCode;
import com.grtsinry43.grtblog.dto.AdminTokenRequest;
import com.grtsinry43.grtblog.dto.ApiResponse;
import com.grtsinry43.grtblog.entity.AdminToken;
import com.grtsinry43.grtblog.exception.BusinessException;
import com.grtsinry43.grtblog.service.AdminTokenService;
import com.grtsinry43.grtblog.util.SecurityUtils;
import com.grtsinry43.grtblog.util.TokenGenerator;
import org.springframework.web.bind.annotation.*;

import java.sql.Timestamp;
import java.util.List;
import java.util.Objects;

/**
 * @author grtsinry43
 * @date 2025/1/26 20:31
 * @description 热爱可抵岁月漫长
 */
@RestController
@RequestMapping("/token")
public class TokenController {
    private final AdminTokenService adminTokenService;

    public TokenController(AdminTokenService adminTokenService) {
        this.adminTokenService = adminTokenService;
    }

    @GetMapping("/test")
    public String test() {
        return "test";
    }

    @GetMapping
    public ApiResponse<List<AdminToken>> getToken() {
        return ApiResponse.success(adminTokenService.list().stream()
                .filter(adminToken -> adminToken.getExpireAt().after(new Timestamp(System.currentTimeMillis())))
                .toList());
    }

    @PostMapping
    public ApiResponse<AdminToken> createToken(@RequestBody AdminTokenRequest tokenRequest) {
        Long userId = Objects.requireNonNull(SecurityUtils.getCurrentUser()).getId();
        if (userId == null) {
            throw new BusinessException(ErrorCode.NOT_LOGIN);
        }
        TokenGenerator.TokenPair token = TokenGenerator.generateToken();
        AdminToken adminToken = new AdminToken();
        adminToken.setDescription(tokenRequest.getDescription());
        adminToken.setExpireAt(tokenRequest.getExpireAt());
        adminToken.setUserId(userId);
        adminToken.setToken(token.getHashedToken());
        // 保存 hash 之后的 token
        adminTokenService.save(adminToken);
        // 返回原始 token 并添加前缀
        adminToken.setToken("gb_tk_" + token.getToken());
        return ApiResponse.success(adminToken);
    }

    @DeleteMapping("/{id}")
    public ApiResponse<String> deleteToken(@PathVariable Long id) {
        AdminToken tokenMatched = adminTokenService.getById(id);
        if (tokenMatched == null) {
            throw new BusinessException(ErrorCode.PARAMS_ERROR);
        }
        tokenMatched.setExpireAt(new Timestamp(System.currentTimeMillis()));
        adminTokenService.updateById(tokenMatched);
        return ApiResponse.success("Token has been marked as expired.");
    }
}
