package com.grtsinry43.grtblog.service.impl;

import com.baomidou.mybatisplus.core.conditions.query.QueryWrapper;
import com.grtsinry43.grtblog.common.ErrorCode;
import com.grtsinry43.grtblog.dto.CommentLoginForm;
import com.grtsinry43.grtblog.dto.CommentNotLoginForm;
import com.grtsinry43.grtblog.entity.*;
import com.grtsinry43.grtblog.exception.BusinessException;
import com.grtsinry43.grtblog.mapper.CommentMapper;
import com.grtsinry43.grtblog.mapper.UserRoleMapper;
import com.grtsinry43.grtblog.service.*;
import com.baomidou.mybatisplus.extension.service.impl.ServiceImpl;
import com.grtsinry43.grtblog.util.MD5Util;
import com.grtsinry43.grtblog.util.MarkdownConverter;
import com.grtsinry43.grtblog.util.UserAgentUtil;
import com.grtsinry43.grtblog.vo.CommentVO;
import com.grtsinry43.grtblog.vo.CommentView;
import org.springframework.beans.BeanUtils;
import org.springframework.context.annotation.Lazy;
import org.springframework.stereotype.Service;

import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;
import java.util.*;
import java.util.stream.Collectors;

/**
 * <p>
 * 服务实现类
 * </p>
 *
 * @author grtsinry43
 * @since 2024-10-09
 */
@Service
public class CommentServiceImpl extends ServiceImpl<CommentMapper, Comment> implements ICommentService {
    private final ArticleServiceImpl articleService;
    private final CommentAreaService commentAreaService;
    private final UserServiceImpl userService;
    private final EmailService emailService;
    private final StatusUpdateServiceImpl statusUpdateService;
    private final PageService pageService;
    private final WebsiteInfoServiceImpl websiteInfoServiceImpl;
    private final FriendLinkServiceImpl friendLinkServiceImpl;
    private final UserRoleMapper userRoleMapper;
    private final SocketIOService socketIOService;

    public CommentServiceImpl(ArticleServiceImpl articleService, CommentAreaService commentAreaService, UserServiceImpl userService, EmailService emailService, StatusUpdateServiceImpl statusUpdateService, PageService pageService, WebsiteInfoServiceImpl websiteInfoServiceImpl, FriendLinkServiceImpl friendLinkServiceImpl, UserRoleMapper userRoleMapper, @Lazy SocketIOService socketIOService) {
        this.articleService = articleService;
        this.commentAreaService = commentAreaService;
        this.userService = userService;
        this.emailService = emailService;
        this.statusUpdateService = statusUpdateService;
        this.pageService = pageService;
        this.websiteInfoServiceImpl = websiteInfoServiceImpl;
        this.friendLinkServiceImpl = friendLinkServiceImpl;
        this.userRoleMapper = userRoleMapper;
        this.socketIOService = socketIOService;
    }

    /**
     * 组装评论树
     *
     * @param comments 评论列表
     * @return 评论树
     */
    public List<CommentView> getCommentTree(List<Comment> comments) {
        // 首先流式将 Comment 转换为 CommentView
        List<CommentView> commentViewList = comments.stream().map(comment -> {
            CommentView commentView = new CommentView();
            BeanUtils.copyProperties(comment, commentView);
            if (comment.getDeletedAt() != null) {
                commentView.setContent("***该评论已被删除***");
                commentView.setIsDeleted(true);
            } else {
                commentView.setIsDeleted(false);
            }
            commentView.setId(comment.getId().toString());
            commentView.setAreaId(comment.getAreaId().toString());
            commentView.setUserName(comment.getNickName());
            commentView.setAvatarUrl(comment.getAuthorId() == null ? "https://cravatar.cn/avatar/" + MD5Util.getMD5(comment.getEmail()).toLowerCase() + "?d=retro"
                    : userService.getById(comment.getAuthorId()).getAvatar());
            commentView.setParentId(comment.getParentId() == null ? null : comment.getParentId().toString());
            commentView.setParentUserName(comment.getParentId() == null ? null : getById(comment.getParentId()).getNickName());
            return commentView;
        }).toList();

        // 构建评论映射
        Map<String, CommentView> commentVOMap = commentViewList.stream()
                .collect(Collectors.toMap(CommentView::getId, commentView -> commentView));

        // 组装评论树
        List<CommentView> commentTree = new ArrayList<>();
        for (CommentView commentView : commentViewList) {
            String parentId = commentView.getParentId();
            if (parentId == null) {
                // 顶级评论
                commentTree.add(commentView);
            } else {
                // 子评论，添加到父评论的子评论列表
                CommentView parentComment = commentVOMap.get(parentId);
                if (parentComment != null) {
                    parentComment.getChildren().add(commentView);
                }
            }
        }
        return commentTree;
    }

    @Override
    public CommentView addNewComment(CommentNotLoginForm form, String ip, String location, String ua) {
        System.out.println(form.toString());

        String os = UserAgentUtil.extractOS(ua);

        String browser = UserAgentUtil.extractBrowser(ua);

        // 查询评论区是否存在
        if (!commentAreaService.isExist(form.getAreaId())) {
            throw new BusinessException(ErrorCode.PARAMS_ERROR);
        }
        long areaLong = Long.parseLong(form.getAreaId());
        Comment comment = new Comment();
        BeanUtils.copyProperties(form, comment);
        comment.setNickName(form.getUserName());
        comment.setAreaId(areaLong);
        comment.setParentId(Objects.equals(form.getParentId(), "") ? null : Long.parseLong(form.getParentId()));
        comment.setIp(ip);
        comment.setLocation(location);
        comment.setPlatform(os);
        comment.setBrowser(browser);
        save(comment);
        // 这里处理一下评论内容，如果有父评论，就发送邮件通知
        commentEmailNotificationHandle(comment);
        CommentView vo = new CommentView();
        BeanUtils.copyProperties(comment, vo);
        vo.setCreatedAt(LocalDateTime.now());
        vo.setUpdatedAt(LocalDateTime.now());
        socketIOService.broadcastNotification("有新的评论发布，作者：" + comment.getNickName() + "，内容：" + comment.getContent() + "，评论区：" + comment.getAreaId() + "，快去围观吧！");
        return vo;
    }

    @Override
    public CommentView addNewCommentLogin(User user, CommentLoginForm form, String ip, String location, String ua) {
        String os = UserAgentUtil.extractOS(ua);

        String browser = UserAgentUtil.extractBrowser(ua);

        // 查询评论区是否存在
        if (!commentAreaService.isExist(form.getAreaId())) {
            throw new BusinessException(ErrorCode.PARAMS_ERROR);
        }
        long areaLong = Long.parseLong(form.getAreaId());
        Comment comment = new Comment();
        BeanUtils.copyProperties(form, comment);
        comment.setNickName(user.getNickname());
        comment.setAreaId(areaLong);
        comment.setAuthorId(user.getId());
        comment.setParentId(Objects.equals(form.getParentId(), "") ? null : Long.parseLong(form.getParentId()));
        comment.setIp(ip);
        comment.setIsFriend(friendLinkServiceImpl.isMyFriend(user.getId()));
        comment.setIsOwner(userRoleMapper.isUserAdmin(user.getId()));
        comment.setLocation(location);
        comment.setPlatform(os);
        comment.setBrowser(browser);
        save(comment);
        // 这里处理一下评论内容，如果有父评论，就发送邮件通知
        commentEmailNotificationHandle(comment);
        CommentView vo = new CommentView();
        BeanUtils.copyProperties(comment, vo);
        vo.setCreatedAt(LocalDateTime.now());
        vo.setUpdatedAt(LocalDateTime.now());
        socketIOService.broadcastNotification("有新的评论发布，作者：" + comment.getNickName() + "，内容：" + comment.getContent() + "，评论区：" + comment.getAreaId() + "，快去围观吧！");
        return vo;
    }

    @Override
    public List<CommentView> listCommentByArticleId(String shortUrl) {
        Article article = articleService.getBaseMapper().getArticleByShortUrl(shortUrl);
        Long commentId = article.getCommentId();
        List<Comment> comments = list(new QueryWrapper<Comment>().lambda().eq(Comment::getAreaId, commentId));
        return getCommentTree(comments);
    }

    @Override
    public List<CommentView> getListById(Long id, int page, int pageSize) {
        // 查询所有评论
        List<Comment> allComments = list(new QueryWrapper<Comment>()
                .lambda()
                .eq(Comment::getAreaId, id)
                .orderByDesc(Comment::getIsTop) // 先按置顶字段降序排列
                .orderByAsc(Comment::getParentId)
                .orderByDesc(Comment::getCreatedAt));

        // 构建评论树
        List<CommentView> commentTree = getCommentTree(allComments);

        // 进行分页处理
        int fromIndex = (page - 1) * pageSize;
        int toIndex = Math.min(fromIndex + pageSize, commentTree.size());
        if (fromIndex > commentTree.size()) {
            return Collections.emptyList();
        }
        return commentTree.subList(fromIndex, toIndex);
    }

    public void commentEmailNotificationHandle(Comment comment) {
        // 这里处理一下评论内容，如果有父评论，就发送邮件通知
        if (comment.getParentId() != null) {
            Comment parentComment = getById(comment.getParentId());
            if (parentComment != null) {
                String content = comment.getContent();
                // 这里将 md 渲染为 html
                String contentHtml = MarkdownConverter.convertMarkdownToHtml(content);
                // 根据评论区 id 获取原内容，可能是文章、动态、页面
                String title = "";
                String type = "";
                String url = websiteInfoServiceImpl.getWebsiteInfo("WEBSITE_URL");
                if (articleService.lambdaQuery().eq(Article::getCommentId, parentComment.getAreaId()).count() > 0) {
                    Article article = articleService.lambdaQuery().eq(Article::getCommentId, parentComment.getAreaId()).one();
                    title = article.getTitle();
                    type = "文章";
                    url += "/posts/" + article.getShortUrl();
                } else if (statusUpdateService.lambdaQuery().eq(StatusUpdate::getCommentId, parentComment.getAreaId()).count() > 0) {
                    StatusUpdate statusUpdate = statusUpdateService.lambdaQuery().eq(StatusUpdate::getCommentId, parentComment.getAreaId()).one();
                    title = statusUpdate.getTitle();
                    type = "动态";
                    url += "/moments/" + statusUpdate.getCreatedAt().format(DateTimeFormatter.ofPattern("yyyy/MM/dd")) + "/" + statusUpdate.getShortUrl();
                } else if (pageService.lambdaQuery().eq(Page::getCommentId, parentComment.getAreaId()).count() > 0) {
                    Page page = pageService.lambdaQuery().eq(Page::getCommentId, parentComment.getAreaId()).one();
                    title = page.getTitle();
                    type = "页面";
                    url += "/" + page.getRefPath();
                }
                HashMap<String, String> info = new HashMap<>();
                info.put("name", parentComment.getNickName());
                info.put("type", type);
                info.put("title", title);
                info.put("avatarUrl", "");
                info.put("replyName", comment.getNickName());
                info.put("replyTime", LocalDateTime.now().format(DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm:ss")));
                info.put("replyContent", contentHtml);
                info.put("commentContent", MarkdownConverter.convertMarkdownToHtml(parentComment.getContent()));
                info.put("url", url);
                try {
                    emailService.sendEmail(parentComment.getAuthorId() != null ? userService.getById(parentComment.getAuthorId()).getEmail() : parentComment.getEmail()
                            , "[" + websiteInfoServiceImpl.getWebsiteInfo("HOME_TITLE") + "] 有人回复了您的评论", info);
                } catch (Exception e) {
                    e.printStackTrace();
                }
            }
        }
    }

    public List<CommentVO> listAllAdmin(int page, int pageSize) {
        List<Comment> comments = list(new QueryWrapper<Comment>()
                .orderByDesc("created_at")
                .isNull("deleted_at")
                .last("LIMIT " + (page - 1) * pageSize + "," + pageSize));
        return comments.stream().map(this::convertToCommentVO).collect(Collectors.toList());
    }

    public List<CommentVO> listByAreaIdAdmin(Long areaId, int page, int pageSize) {
        List<Comment> comments = list(new QueryWrapper<Comment>()
                .lambda()
                .eq(Comment::getAreaId, areaId)
                .orderByDesc(Comment::getCreatedAt)
                .last("LIMIT " + (page - 1) * pageSize + "," + pageSize));
        return comments.stream().map(this::convertToCommentVO).collect(Collectors.toList());
    }

    public List<CommentVO> listAllNotReadAdmin(int page, int pageSize) {
        List<Comment> comments = list(new QueryWrapper<Comment>()
                .lambda()
                .eq(Comment::getIsViewed, false)
                .orderByDesc(Comment::getCreatedAt)
                .last("LIMIT " + (page - 1) * pageSize + "," + pageSize));
        return comments.stream().map(this::convertToCommentVO).collect(Collectors.toList());
    }

    public List<CommentVO> listByAreaIdNotReadAdmin(Long areaId, int page, int pageSize) {
        List<Comment> comments = list(new QueryWrapper<Comment>()
                .lambda()
                .eq(Comment::getAreaId, areaId)
                .eq(Comment::getIsViewed, false)
                .orderByDesc(Comment::getCreatedAt)
                .last("LIMIT " + (page - 1) * pageSize + "," + pageSize));
        return comments.stream().map(this::convertToCommentVO).collect(Collectors.toList());
    }

    public long countAllAdmin() {
        return count(new QueryWrapper<Comment>());
    }

    public long countAllNotReadAdmin() {
        return count(new QueryWrapper<Comment>().lambda().eq(Comment::getIsViewed, false));
    }

    public long countByAreaIdAdmin(Long areaId) {
        return count(new QueryWrapper<Comment>().lambda().eq(Comment::getAreaId, areaId));
    }

    public long countByAreaIdNotReadAdmin(Long areaId) {
        return count(new QueryWrapper<Comment>().lambda().eq(Comment::getAreaId, areaId).eq(Comment::getIsViewed, false));
    }

    private CommentVO convertToCommentVO(Comment comment) {
        CommentVO commentVO = new CommentVO();
        BeanUtils.copyProperties(comment, commentVO);
        commentVO.setId(comment.getId().toString());
        commentVO.setAuthorId(commentVO.getAuthorId());
        commentVO.setAreaId(comment.getAreaId().toString());
        commentVO.setParentId(comment.getParentId() == null ? null : comment.getParentId().toString());
        return commentVO;
    }
}
