package com.grtsinry43.grtblog.dto;

import lombok.Data;

import java.io.Serializable;

/**
 * @Author： grtsinry43
 * @Date： 2024/7/13 上午12:27
 * 这里规定返回消息格式，所有的正确响应HTTP状态码都是200，只会根据code字段和msg来提供错误信息
 */
@Data
public class ApiResponse<T> implements Serializable {
    private Integer code;
    private String msg = "";
    private T data;

    /**
     * 含有数据的成功响应
     *
     * @param data 返回的数据
     * @param <T>  数据类型
     */
    public static <T> ApiResponse<T> success(T data) {
        ApiResponse<T> response = new ApiResponse<>();
        response.setCode(0);
        response.setData(data);
        return response;
    }

    /**
     * 不含数据的成功响应
     */
    public static <T> ApiResponse<T> success() {
        return success(null);
    }

    /**
     * 失败响应，包含错误码和错误信息
     * @param code 错误码
     * @param msg  错误信息
     */
    public static <T> ApiResponse<T> error(Integer code, String msg) {
        ApiResponse<T> response = new ApiResponse<>();
        response.setCode(code);
        response.setMsg(msg);
        return response;
    }
}
