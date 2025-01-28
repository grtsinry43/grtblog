package com.grtsinry43.grtblog.controller;

import com.grtsinry43.grtblog.dto.ApiResponse;
import com.grtsinry43.grtblog.entity.Article;
import com.grtsinry43.grtblog.entity.Page;
import com.grtsinry43.grtblog.entity.StatusUpdate;
import com.grtsinry43.grtblog.service.CommentAreaService;
import com.grtsinry43.grtblog.service.PageService;
import com.grtsinry43.grtblog.service.impl.ArticleServiceImpl;
import com.grtsinry43.grtblog.service.impl.CommentServiceImpl;
import com.grtsinry43.grtblog.service.impl.StatusUpdateServiceImpl;
import com.grtsinry43.grtblog.vo.OverviewVO;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

/**
 * @author grtsinry43
 * @date 2025/1/28 08:20
 * @description 热爱可抵岁月漫长
 */
@RestController
@RequestMapping("/overview")
public class OverviewController {
    private final ArticleServiceImpl articleService;
    private final StatusUpdateServiceImpl statusUpdateService;
    private final CommentServiceImpl commentService;
    private final CommentAreaService commentAreaService;
    private final PageService pageService;

    public OverviewController(ArticleServiceImpl articleService, StatusUpdateServiceImpl statusUpdateService, CommentServiceImpl commentService, CommentAreaService commentAreaService, PageService pageService) {
        this.articleService = articleService;
        this.statusUpdateService = statusUpdateService;
        this.commentService = commentService;
        this.commentAreaService = commentAreaService;
        this.pageService = pageService;
    }

    @GetMapping
    public ApiResponse<OverviewVO> getOverview() {
        OverviewVO overview = new OverviewVO();
        overview.setArticleCount(articleService.getAllArticleShortLinks().size());
        overview.setMomentCount(statusUpdateService.getAllStatusUpdateShortLinks().size());
        int commentCount = 0;
        int allReadingCount = 0;
        for (Article article : articleService.lambdaQuery().isNull(Article::getDeletedAt).eq(Article::getIsPublished, true).list()) {
            commentCount += article.getComments();
            allReadingCount += article.getViews();
        }
        for (StatusUpdate statusUpdate : statusUpdateService.lambdaQuery().isNull(StatusUpdate::getDeletedAt).eq(StatusUpdate::getIsPublished, true).list()) {
            commentCount += statusUpdate.getComments();
            allReadingCount += statusUpdate.getViews();
        }
        for (Page page : pageService.lambdaQuery().isNull(Page::getDeletedAt).eq(Page::getCanDelete, true).eq(Page::getEnable, true).list()) {
            commentCount += page.getComments();
            allReadingCount += page.getViews();
        }
        overview.setCommentCount(commentCount);
        overview.setAllReadingCount(allReadingCount);
        overview.setPageCount(pageService.getAllPageRefPath().length);
        return ApiResponse.success(overview);
    }
}
