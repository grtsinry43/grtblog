package com.grtsinry43.grtblog.mapper;

import com.grtsinry43.grtblog.entity.StatusUpdate;
import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import org.apache.ibatis.annotations.Select;

import java.util.List;

/**
 * <p>
 * Mapper 接口
 * </p>
 *
 * @author grtsinry43
 * @since 2024-10-09
 */
public interface StatusUpdateMapper extends BaseMapper<StatusUpdate> {
    /**
     * 获取最近的四条说说
     */
    @Select("SELECT * FROM status_update ORDER BY created_at DESC LIMIT 4")
    public List<StatusUpdate> selectLastFourStatusUpdates();

    /**
     * 获取最近的一条
     */
    @Select("SELECT * FROM status_update ORDER BY created_at DESC LIMIT 1")
    public StatusUpdate selectLastStatusUpdate();
}
