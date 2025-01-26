package com.grtsinry43.grtblog.service;

import com.grtsinry43.grtblog.entity.*;
import com.grtsinry43.grtblog.mapper.*;
import com.grtsinry43.grtblog.security.LoginUserDetails;
import jakarta.annotation.Resource;
import org.springframework.security.core.userdetails.UserDetails;
import org.springframework.security.core.userdetails.UserDetailsService;
import org.springframework.security.core.userdetails.UsernameNotFoundException;
import org.springframework.stereotype.Service;
import org.springframework.util.CollectionUtils;

import java.util.ArrayList;
import java.util.List;
import java.util.Objects;
import java.util.stream.Collectors;

/**
 * @author grtsinry43
 * @date 2024/11/9 19:30
 * @description 为 Spring Security 提供用户信息的服务类
 */
@Service
public class UserDetailsServiceImpl implements UserDetailsService {
    @Resource
    private UserMapper userMapper;
    @Resource
    private UserRoleMapper userRoleMapper;
    @Resource
    private RoleMapper roleMapper;
    @Resource
    private RolePermissionMapper rolePermissionMapper;
    @Resource
    private PermissionMapper permissionMapper;

    @Override
    public UserDetails loadUserByUsername(String userEmail) throws UsernameNotFoundException {
        // 通过邮箱信息从数据库中查询用户信息
        User user = userMapper.getUserByUserEmail(userEmail);
        // 判断当前账号是否存在
        if (Objects.isNull(user)) {
            // 如果为空，直接抛出异常
            throw new UsernameNotFoundException(userEmail);
        }

        // 获取角色名称列表
        List<String> roleNames = getRoleNamesByUserId(user.getId());
        // 获取权限名称列表
        List<String> permissionNames = getPermissionNamesByRoleNames(roleNames);

        // 返回用户信息
        return new LoginUserDetails(user, roleNames, permissionNames);
    }

    public UserDetails loadUserByUserId(String userId) throws UsernameNotFoundException {
        // 通过邮箱信息从数据库中查询用户信息
        User user = userMapper.selectById(userId);
        // 判断当前账号是否存在
        if (Objects.isNull(user)) {
            // 如果为空，直接抛出异常
            throw new UsernameNotFoundException(userId);
        }

        // 获取角色名称列表
        List<String> roleNames = getRoleNamesByUserId(user.getId());
        // 获取权限名称列表
        List<String> permissionNames = getPermissionNamesByRoleNames(roleNames);

        // 返回用户信息
        return new LoginUserDetails(user, roleNames, permissionNames);
    }

    private List<String> getRoleNamesByUserId(Long userId) {
        List<UserRole> userRoles = userRoleMapper.getUserRolesByUserId(userId);
        if (CollectionUtils.isEmpty(userRoles)) {
            return List.of();
        }

        List<Long> roleIds = userRoles.stream().map(UserRole::getRoleId).collect(Collectors.toList());
        List<Role> roles = roleMapper.batchGetRolesByRoleIds(roleIds);
        return roles.stream().map(Role::getRoleName).collect(Collectors.toList());
    }

    private List<String> getPermissionNamesByRoleNames(List<String> roleNames) {
        if (CollectionUtils.isEmpty(roleNames)) {
            return List.of();
        }

        List<Long> roleIds = roleMapper.getRoleIdsByRoleNames(roleNames);
        List<RolePermission> rolePermissions = rolePermissionMapper.getRolePermissionsByRoleIds(roleIds);
        if (CollectionUtils.isEmpty(rolePermissions)) {
            return List.of();
        }

        List<Long> permissionIds = rolePermissions.stream().map(RolePermission::getPermissionId).collect(Collectors.toList());
        List<Permission> permissions = permissionMapper.batchGetPermissionsByPermissionIds(permissionIds);
        return permissions.stream().map(Permission::getPermissionName).collect(Collectors.toList());
    }
}

