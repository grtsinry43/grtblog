package com.grtsinry43.grtblog.service;

import com.baomidou.mybatisplus.core.conditions.query.QueryWrapper;
import com.baomidou.mybatisplus.extension.service.impl.ServiceImpl;
import com.grtsinry43.grtblog.entity.Thinking;
import com.grtsinry43.grtblog.mapper.ThinkingMapper;
import org.springframework.stereotype.Service;

import java.util.List;

/**
 * @author grtsinry43
 * @date 2025/1/31 21:49
 * @description 热爱可抵岁月漫长
 */
@Service
public class ThinkingService extends ServiceImpl<ThinkingMapper, Thinking> {
    public Thinking getLatest() {
        QueryWrapper<Thinking> queryWrapper = new QueryWrapper<>();
        queryWrapper.orderByDesc("created_at").last("LIMIT 1");
        return this.baseMapper.selectOne(queryWrapper);
    }

    public List<Thinking> listAll() {
        return this.baseMapper.selectList(null);
    }
}
