package com.grtsinry43.grtblog.controller;

import com.grtsinry43.grtblog.dto.ApiResponse;
import com.grtsinry43.grtblog.exception.BusinessException;
import com.grtsinry43.grtblog.exception.TestException;
import jakarta.validation.ConstraintViolationException;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.HttpRequestMethodNotSupportedException;
import org.springframework.web.bind.MissingServletRequestParameterException;
import org.springframework.web.bind.annotation.ControllerAdvice;
import org.springframework.web.bind.annotation.ExceptionHandler;

import java.util.concurrent.atomic.AtomicInteger;

/**
 * @author grtsinry43
 * @date 2024/7/13 上午 12:29
 * @description 规定触发异常时的返回格式，依然是 HTTP 状态码 200，只会根据 code 字段和 msg 来提供错误信息
 */
@ControllerAdvice
public class GlobalExceptionHandler {
    private static final AtomicInteger businessExceptionCount = new AtomicInteger(0);
    private static final Logger logger = LoggerFactory.getLogger(GlobalExceptionHandler.class);

    /**
     * 处理缺少请求参数异常
     */
    @ExceptionHandler(MissingServletRequestParameterException.class)
    public ResponseEntity<ApiResponse<Object>> handleMissingServletRequestParameterException(Exception ex) {
        String errorMessage = "缺少请求参数：" + ex.getMessage();
        logger.error(errorMessage);
        ApiResponse<Object> apiResponse = ApiResponse.error(400, errorMessage);
        return new ResponseEntity<>(apiResponse, HttpStatus.OK);
    }

    /**
     * 处理参数校验异常（注释声明即可）
     */
    @ExceptionHandler(ConstraintViolationException.class)
    public ResponseEntity<ApiResponse<Object>> handleConstraintViolationException(ConstraintViolationException ex) {
        String errorMessage = ex.getMessage();
        logger.error(errorMessage);
        ApiResponse<Object> apiResponse = ApiResponse.error(400, errorMessage);
        return new ResponseEntity<>(apiResponse, HttpStatus.OK);
    }

    /**
     * 处理请求方法错误
     */
    @ExceptionHandler(HttpRequestMethodNotSupportedException.class)
    public ResponseEntity<ApiResponse<Object>> handleHttpRequestMethodNotSupportedException(HttpRequestMethodNotSupportedException ex) {
        String errorMessage = ex.getMessage();
        logger.error(errorMessage);
        ApiResponse<Object> apiResponse = ApiResponse.error(400, errorMessage);
        return new ResponseEntity<>(apiResponse, HttpStatus.OK);
    }

    /**
     * 处理业务逻辑异常（注释声明即可）
     */
    @ExceptionHandler(BusinessException.class)
    public ResponseEntity<ApiResponse<Object>> handleBusinessException(BusinessException ex) {
        businessExceptionCount.incrementAndGet();
        String errorMessage = ex.getMessage();
        logger.error(errorMessage);
        ApiResponse<Object> apiResponse = ApiResponse.error(ex.getErrorCode().getCode(), errorMessage);
        return new ResponseEntity<>(apiResponse, HttpStatus.OK);
    }

    /**
     * 处理其他异常（注释声明即可）
     */
    @ExceptionHandler(TestException.class)
    public ResponseEntity<ApiResponse<Object>> handleTestException(TestException ex) {
        String errorMessage = ex.getMessage();
        logger.error(errorMessage);
        ApiResponse<Object> apiResponse = ApiResponse.error(500, errorMessage);
        return new ResponseEntity<>(apiResponse, HttpStatus.OK);
    }
    /**
     * 处理其他异常（注释声明即可）
     */
    @ExceptionHandler(Exception.class)
    public ResponseEntity<ApiResponse<Object>> handleException(Exception ex) {
        String errorMessage = ex.getMessage();
        logger.error(errorMessage);
        ApiResponse<Object> apiResponse = ApiResponse.error(500, errorMessage);
        return new ResponseEntity<>(apiResponse, HttpStatus.OK);
    }

    public static int getBusinessExceptionCount() {
        return businessExceptionCount.get();
    }
}