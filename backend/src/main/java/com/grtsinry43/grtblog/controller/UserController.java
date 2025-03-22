package com.grtsinry43.grtblog.controller;

import com.grtsinry43.grtblog.dto.ApiResponse;
import com.grtsinry43.grtblog.dto.ArticleDTO;
import com.grtsinry43.grtblog.dto.NickNameUpdateRequest;
import com.grtsinry43.grtblog.dto.UserRegisterDTO;
import com.grtsinry43.grtblog.entity.Article;
import com.grtsinry43.grtblog.entity.PasswordResetToken;
import com.grtsinry43.grtblog.entity.User;
import com.grtsinry43.grtblog.security.LoginUserDetails;
import com.grtsinry43.grtblog.service.EmailService;
import com.grtsinry43.grtblog.service.PasswordResetTokenService;
import com.grtsinry43.grtblog.service.impl.UserServiceImpl;
import com.grtsinry43.grtblog.util.JwtUtil;
import com.grtsinry43.grtblog.vo.UserVO;
import jakarta.annotation.Resource;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import org.springframework.beans.BeanUtils;
import org.springframework.security.authentication.AuthenticationManager;
import org.springframework.security.authentication.UsernamePasswordAuthenticationToken;
import org.springframework.security.core.Authentication;
import org.springframework.security.core.annotation.AuthenticationPrincipal;
import org.springframework.security.core.context.SecurityContextHolder;
import org.springframework.security.crypto.bcrypt.BCryptPasswordEncoder;
import org.springframework.web.bind.annotation.*;

import java.time.LocalDateTime;
import java.util.HashMap;
import java.util.Map;
import java.util.Objects;

/**
 * 用户控制器
 *
 * @author grtsinry43
 * @since 2024-10-09
 */
@RestController
@RequestMapping("/user")
public class UserController {
    private final UserServiceImpl userService;
    private final AuthenticationManager authenticationManager;
    private final PasswordResetTokenService passwordResetTokenService;
    private final EmailService emailService;

    public UserController(UserServiceImpl userService, AuthenticationManager authenticationManager, PasswordResetTokenService passwordResetTokenService, EmailService emailService) {
        this.userService = userService;
        this.authenticationManager = authenticationManager;
        this.passwordResetTokenService = passwordResetTokenService;
        this.emailService = emailService;
    }

    @PostMapping("/login")
    public ApiResponse<UserVO> login(HttpServletRequest request, String userEmail, String password, HttpServletResponse response) {
        UsernamePasswordAuthenticationToken authenticationToken = new UsernamePasswordAuthenticationToken(userEmail, password);
        try {
            String sessionCaptcha = (String) request.getSession().getAttribute("captcha");
            if (sessionCaptcha == null || !sessionCaptcha.equals(request.getParameter("captcha"))) {
                return ApiResponse.error(401, "验证码错误");
            }
            Authentication authenticate = authenticationManager.authenticate(authenticationToken);
            if (authenticate.isAuthenticated()) {
                LoginUserDetails principal = (LoginUserDetails) authenticate.getPrincipal();
                if (Objects.isNull(principal)) {
                    return ApiResponse.error(400, "登录失败，请检查用户名和密码");
                }
                // 这里说明登录成功，终于能获取到 User 对象了
                User user = principal.getUser();
                // 转成 VO 对象
                UserVO userVO = new UserVO();
                BeanUtils.copyProperties(user, userVO);
                userVO.setId(user.getId().toString());
                // 生成 token
                String token = JwtUtil.generateToken(user.getEmail());
                response.setHeader("Authorization", token);
                return ApiResponse.success(userVO);
            }
        } catch (Exception e) {
            return ApiResponse.error(401, "登录失败，请检查用户名和密码");
        }
        return ApiResponse.error(401, "登录失败，请检查用户名和密码");
    }

    // TODO test
    @PostMapping("/register")
    public ApiResponse<UserVO> registerApi(HttpServletRequest request, @RequestBody UserRegisterDTO user) {
        String sessionCaptcha = (String) request.getSession().getAttribute("captcha");
        System.out.println("sessionCaptcha = " + sessionCaptcha);
        System.out.println("request.getParameter(\"captcha\") = " + request.getParameter("captcha"));
        if (sessionCaptcha == null || !sessionCaptcha.equals(request.getParameter("captcha"))) {
            return ApiResponse.error(401, "验证码错误");
        }
        if (userService.getUserByEmail(user.getUserEmail()) != null) {
            return ApiResponse.error(400, "邮箱已被注册");
        }
        User user1 = new User();
        BeanUtils.copyProperties(user, user1);
        user1.setNickname(user.getNickname());
        user1.setEmail(user.getUserEmail());
        // 这里要加密密码！！！！！！！！
        BCryptPasswordEncoder bCryptPasswordEncoder = new BCryptPasswordEncoder();
        user1.setPassword(bCryptPasswordEncoder.encode(user.getPassword()));
        userService.save(user1);
        UserVO userVO = new UserVO();
        BeanUtils.copyProperties(user1, userVO);
        userVO.setId(user1.getId().toString());
        return ApiResponse.success(userVO);
    }

    @PostMapping("/request-password-reset")
    public ApiResponse<String> requestPasswordReset(@RequestParam String email, HttpServletRequest request) {
        String sessionCaptcha = (String) request.getSession().getAttribute("captcha");
        if (sessionCaptcha == null || !sessionCaptcha.equals(request.getParameter("captcha"))) {
            return ApiResponse.error(401, "验证码错误");
        }
        PasswordResetToken token = passwordResetTokenService.createToken(email);
        emailService.sendPasswordResetEmail(email, token.getToken());
        return ApiResponse.success("重置密码邮件已发送");
    }

    @PostMapping("/reset-password")
    public ApiResponse<String> resetPassword(@RequestParam String token, @RequestParam String newPassword) {
        PasswordResetToken resetToken = passwordResetTokenService.getToken(token);

        if (resetToken == null || resetToken.getExpiryDate().isBefore(LocalDateTime.now())) {
            return ApiResponse.error(400, "无效的重置密码令牌");
        }

         userService.updatePassword(resetToken.getEmail(), newPassword);

        return ApiResponse.success("密码重置成功");
    }

    @GetMapping("/info")
    public ApiResponse<UserVO> getCurrentUser() {
        Authentication authentication = SecurityContextHolder.getContext().getAuthentication();
        LoginUserDetails principal = (LoginUserDetails) authentication.getPrincipal();
        if (Objects.isNull(principal)) {
            return ApiResponse.error(400, "登录失败，请检查用户名和密码");
        }
        // 这里说明登录成功，终于能获取到 User 对象了
        User user = principal.getUser();
        UserVO userVO = new UserVO();
        BeanUtils.copyProperties(user, userVO);
        userVO.setId(user.getId().toString());
        return ApiResponse.success(userVO);
    }

    @PatchMapping("/update/nickname")
    public ApiResponse<UserVO> updateNickname(@RequestBody NickNameUpdateRequest request) {
        Authentication authentication = SecurityContextHolder.getContext().getAuthentication();
        LoginUserDetails principal = (LoginUserDetails) authentication.getPrincipal();
        if (Objects.isNull(principal)) {
            return ApiResponse.error(400, "登录失败，请检查用户名和密码");
        }
        // 这里说明登录成功，终于能获取到 User 对象了
        User user1 = principal.getUser();
        user1.setNickname(request.getNickName());
        userService.updateById(user1);
        UserVO userVO = new UserVO();
        BeanUtils.copyProperties(user1, userVO);
        userVO.setId(user1.getId().toString());
        return ApiResponse.success(userVO);
    }
}
