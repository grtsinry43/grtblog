package com.grtsinry43.grtblog.vo;

import lombok.Data;

/**
 * @author grtsinry43
 * @date 2025/1/28 08:21
 * @description 热爱可抵岁月漫长
 */
@Data
public class OverviewVO {
    private int articleCount;
    private int momentCount;
    private int commentCount;
    private int pageCount;
    private int allReadingCount;
}
