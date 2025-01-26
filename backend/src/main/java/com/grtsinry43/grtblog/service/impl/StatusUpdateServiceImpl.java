package com.grtsinry43.grtblog.service.impl;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.grtsinry43.grtblog.common.ErrorCode;
import com.grtsinry43.grtblog.dto.PostStatusToggle;
import com.grtsinry43.grtblog.dto.StatusUpdateDTO;
import com.grtsinry43.grtblog.entity.CommentArea;
import com.grtsinry43.grtblog.entity.StatusUpdate;
import com.grtsinry43.grtblog.exception.BusinessException;
import com.grtsinry43.grtblog.mapper.StatusUpdateMapper;
import com.grtsinry43.grtblog.mapper.UserMapper;
import com.grtsinry43.grtblog.security.LoginUserDetails;
import com.grtsinry43.grtblog.service.CommentAreaService;
import com.grtsinry43.grtblog.service.IStatusUpdateService;
import com.baomidou.mybatisplus.extension.service.impl.ServiceImpl;
import com.grtsinry43.grtblog.service.SocketIOService;
import com.grtsinry43.grtblog.util.ArticleParser;
import com.grtsinry43.grtblog.vo.StatusUpdatePreview;
import com.grtsinry43.grtblog.vo.StatusUpdateVO;
import com.grtsinry43.grtblog.vo.StatusUpdateView;
import org.springframework.beans.BeanUtils;
import org.springframework.context.annotation.Lazy;
import org.springframework.stereotype.Service;

import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;
import java.util.List;
import java.util.Objects;
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
public class StatusUpdateServiceImpl extends ServiceImpl<StatusUpdateMapper, StatusUpdate> implements IStatusUpdateService {
    private final UserMapper userMapper;
    private final CategoryServiceImpl categoryService;
    private final CommentAreaService commentAreaService;
    private final SocketIOService socketIOService;

    public StatusUpdateServiceImpl(UserMapper userMapper, CategoryServiceImpl categoryService, CommentAreaService commentAreaService, @Lazy SocketIOService socketIOService) {
        this.userMapper = userMapper;
        this.categoryService = categoryService;
        this.commentAreaService = commentAreaService;
        this.socketIOService = socketIOService;
    }

    @Override
    public List<StatusUpdatePreview> getLastFourStatusUpdates() {
        List<StatusUpdate> statusUpdates = baseMapper.selectLastFourStatusUpdates();
        return statusUpdates.stream()
                .filter(statusUpdate -> statusUpdate.getIsPublished() && Objects.isNull(statusUpdate.getDeletedAt()))
                .map(statusUpdate -> {
                    StatusUpdatePreview preview = new StatusUpdatePreview();
                    BeanUtils.copyProperties(statusUpdate, preview);
                    preview.setAuthorName(this.userMapper.selectById(statusUpdate.getAuthorId()).getNickname());
                    preview.setAuthorAvatar(this.userMapper.selectById(statusUpdate.getAuthorId()).getAvatar());
                    preview.setImages(statusUpdate.getImg() != null ? statusUpdate.getImg().split(",") : new String[0]);
                    preview.setCommentId(statusUpdate.getCommentId() != null ? statusUpdate.getCommentId().toString() : null);
                    preview.setShortUrl(statusUpdate.getCreatedAt().format(DateTimeFormatter.ofPattern("yyyy/MM/dd/")) + statusUpdate.getShortUrl());
                    return preview;
                })
                .collect(Collectors.toList());
    }

    @Override
    public StatusUpdatePreview getLastStatusUpdate() {
        StatusUpdate statusUpdate = baseMapper.selectLastStatusUpdate();
        StatusUpdatePreview preview = new StatusUpdatePreview();
        BeanUtils.copyProperties(statusUpdate, preview);
        preview.setImages(statusUpdate.getImg() != null ? statusUpdate.getImg().split(",") : new String[0]);
        preview.setAuthorName(this.userMapper.selectById(statusUpdate.getAuthorId()).getNickname());
        preview.setCommentId(statusUpdate.getCommentId() != null ? statusUpdate.getCommentId().toString() : null);
        preview.setAuthorAvatar(this.userMapper.selectById(statusUpdate.getAuthorId()).getAvatar());
        preview.setShortUrl(statusUpdate.getCreatedAt().format(DateTimeFormatter.ofPattern("yyyy/MM/dd/")) + statusUpdate.getShortUrl());
        return preview;
    }

    @Override
    public List<StatusUpdatePreview> listStatusUpdatesByPage(int page, int pageSize) {
        int offset = (page - 1) * pageSize;
        List<StatusUpdate> statusUpdates = baseMapper.listStatusUpdatesByPage(offset, pageSize);
        return statusUpdates.stream()
                .map(statusUpdate -> {
                    StatusUpdatePreview preview = new StatusUpdatePreview();
                    BeanUtils.copyProperties(statusUpdate, preview);
                    preview.setAuthorName(this.userMapper.selectById(statusUpdate.getAuthorId()).getNickname());
                    preview.setAuthorAvatar(this.userMapper.selectById(statusUpdate.getAuthorId()).getAvatar());
                    preview.setCommentId(statusUpdate.getCommentId() != null ? statusUpdate.getCommentId().toString() : null);
                    preview.setImages(statusUpdate.getImg() != null ? statusUpdate.getImg().split(",") : new String[0]);
                    preview.setShortUrl(statusUpdate.getCreatedAt().format(DateTimeFormatter.ofPattern("yyyy/MM/dd/")) + statusUpdate.getShortUrl());
                    preview.setSummary(statusUpdate.getSummary() != null ? statusUpdate.getSummary() : statusUpdate.getContent().length() > 200 ? statusUpdate.getContent().substring(0, 200) : statusUpdate.getContent());
                    return preview;
                })
                .collect(Collectors.toList());
    }

    @Override
    public List<StatusUpdatePreview> getStatusUpdatesByCategory(int page, int pageSize, String shortUrl) {
        int offset = (page - 1) * pageSize;
        List<StatusUpdate> statusUpdates = baseMapper.getStatusUpdatesByCategory(offset, pageSize, shortUrl);
        return statusUpdates.stream()
                .map(statusUpdate -> {
                    StatusUpdatePreview preview = new StatusUpdatePreview();
                    BeanUtils.copyProperties(statusUpdate, preview);
                    preview.setAuthorName(this.userMapper.selectById(statusUpdate.getAuthorId()).getNickname());
                    preview.setCommentId(statusUpdate.getCommentId() != null ? statusUpdate.getCommentId().toString() : null);
                    preview.setAuthorAvatar(this.userMapper.selectById(statusUpdate.getAuthorId()).getAvatar());
                    preview.setShortUrl(statusUpdate.getCreatedAt().format(DateTimeFormatter.ofPattern("yyyy/MM/dd/")) + statusUpdate.getShortUrl());
                    preview.setImages(statusUpdate.getImg() != null ? statusUpdate.getImg().split(",") : new String[0]);
                    preview.setSummary(statusUpdate.getSummary() != null ? statusUpdate.getSummary() : statusUpdate.getContent().length() > 200 ? statusUpdate.getContent().substring(0, 200) : statusUpdate.getContent());
                    return preview;
                })
                .collect(Collectors.toList());
    }

    @Override
    public StatusUpdateView getStatusUpdateByShortUrl(String shortUrl) {
        StatusUpdate statusUpdate = baseMapper.getStatusUpdateByShortUrl(shortUrl);
        StatusUpdateView preview = new StatusUpdateView();
        BeanUtils.copyProperties(statusUpdate, preview);
        preview.setImages(statusUpdate.getImg() != null ? statusUpdate.getImg().split(",") : new String[0]);
        preview.setAuthorName(this.userMapper.selectById(statusUpdate.getAuthorId()).getNickname());
        preview.setAuthorAvatar(this.userMapper.selectById(statusUpdate.getAuthorId()).getAvatar());
        preview.setCommentId(statusUpdate.getCommentId() != null ? statusUpdate.getCommentId().toString() : null);
        preview.setCommentId(statusUpdate.getCommentId() != null ? statusUpdate.getCommentId().toString() : null);
        preview.setCategoryName(statusUpdate.getCategoryId() != null ? categoryService.getById(statusUpdate.getCategoryId()).getName() : null);
        return preview;
    }

    @Override
    public List<String> getAllStatusUpdateShortLinks() {
        return this.baseMapper.getAllShortLinks();
    }

    @Override
    public List<StatusUpdateVO> getStatusUpdateListAdmin(int page, int pageSize) {
        int offset = (page - 1) * pageSize;
        List<StatusUpdate> statusUpdates = baseMapper.getStatusUpdateListAdmin(offset, pageSize);
        return statusUpdates.stream()
                .map(statusUpdate -> {
                    StatusUpdateVO statusUpdateVO = new StatusUpdateVO();
                    BeanUtils.copyProperties(statusUpdate, statusUpdateVO);
                    statusUpdateVO.setId(statusUpdate.getId().toString());
                    statusUpdateVO.setAuthorName(this.userMapper.selectById(statusUpdate.getAuthorId()).getNickname());
                    statusUpdateVO.setCategoryId(statusUpdate.getCategoryId() != null ? statusUpdate.getCategoryId().toString() : null);
                    return statusUpdateVO;
                })
                .collect(Collectors.toList());
    }

    @Override
    public StatusUpdateVO addStatusUpdate(StatusUpdateDTO statusUpdateDTO, Long userId) throws JsonProcessingException {
        StatusUpdate statusUpdate = new StatusUpdate();
        BeanUtils.copyProperties(statusUpdateDTO, statusUpdate);
        statusUpdate.setAuthorId(userId);
        statusUpdate.setToc(ArticleParser.generateToc(statusUpdateDTO.getContent()));
        if (!categoryService.isCategoryExist(Long.valueOf(statusUpdateDTO.getCategoryId()))) {
            throw new BusinessException(ErrorCode.PARAMS_ERROR);
        }
        // 创建评论区
        CommentArea commentArea = commentAreaService.createCommentArea("分享", statusUpdate.getTitle());
        statusUpdate.setCommentId(commentArea.getId());
        statusUpdate.setSummary(statusUpdateDTO.getSummary() != null ? statusUpdateDTO.getSummary() : statusUpdateDTO.getContent().length() > 200 ? statusUpdateDTO.getContent().substring(0, 200) : statusUpdateDTO.getContent());
        statusUpdate.setCategoryId(Long.valueOf(statusUpdateDTO.getCategoryId()));
        statusUpdate.setImg(String.join(",", ArticleParser.extractImages(statusUpdateDTO.getContent())));
        this.save(statusUpdate);
        StatusUpdateVO statusUpdateVO = new StatusUpdateVO();
        BeanUtils.copyProperties(statusUpdate, statusUpdateVO);
        statusUpdateVO.setAuthorName(userMapper.selectById(userId).getNickname());
        if (statusUpdate.getIsPublished()) {
            socketIOService.broadcastNotification("分享：" + statusUpdate.getTitle() + " 已发布，新鲜事儿来啦");
        }
        return statusUpdateVO;
    }

    @Override
    public void deleteStatusUpdate(Long id, LoginUserDetails principal) {
        StatusUpdate statusUpdate = this.getById(id);
        if (Objects.equals(statusUpdate.getAuthorId(), principal.getUser().getId())) {
            statusUpdate.setDeletedAt(LocalDateTime.now());
            // 删除评论区
            commentAreaService.deleteCommentArea(statusUpdate.getCommentId());
            this.updateById(statusUpdate);
        } else {
            throw new BusinessException(ErrorCode.UNAUTHORIZED);
        }
    }

    @Override
    public StatusUpdateVO updateStatusUpdate(Long id, StatusUpdateDTO statusUpdateDTO, Long userId) throws JsonProcessingException {
        StatusUpdate statusUpdate = this.getById(id);
        if (Objects.equals(statusUpdate.getAuthorId(), userId)) {
            BeanUtils.copyProperties(statusUpdateDTO, statusUpdate);
            statusUpdate.setToc(ArticleParser.generateToc(statusUpdateDTO.getContent()));
            if (!categoryService.isCategoryExist(Long.valueOf(statusUpdateDTO.getCategoryId()))) {
                throw new BusinessException(ErrorCode.PARAMS_ERROR);
            }
            statusUpdate.setCategoryId(Long.valueOf(statusUpdateDTO.getCategoryId()));
            statusUpdate.setUpdatedAt(LocalDateTime.now());
            statusUpdate.setImg(String.join(",", ArticleParser.extractImages(statusUpdateDTO.getContent())));
            statusUpdate.setSummary(statusUpdateDTO.getSummary() != null ? statusUpdateDTO.getSummary() : statusUpdateDTO.getContent().length() > 200 ? statusUpdateDTO.getContent().substring(0, 200) : statusUpdateDTO.getContent());
            this.updateById(statusUpdate);
            StatusUpdateVO statusUpdateVO = new StatusUpdateVO();
            BeanUtils.copyProperties(statusUpdate, statusUpdateVO);
            statusUpdateVO.setAuthorName(userMapper.selectById(userId).getNickname());
            if (statusUpdate.getIsPublished()) {
                socketIOService.broadcastNotification("分享：" + statusUpdate.getTitle() + " 已更新");
            }
            return statusUpdateVO;
        } else {
            throw new BusinessException(ErrorCode.UNAUTHORIZED);
        }
    }

    @Override
    public StatusUpdateVO toggleStatusUpdate(Long id, PostStatusToggle postStatusToggle) {
        StatusUpdate statusUpdate = this.baseMapper.selectById(id);
        if (statusUpdate == null) {
            throw new BusinessException(ErrorCode.NOT_FOUND);
        }
        statusUpdate.setIsPublished(postStatusToggle.getIsPublished() == null ? statusUpdate.getIsPublished() : postStatusToggle.getIsPublished());
        statusUpdate.setIsTop(postStatusToggle.getIsTop() == null ? statusUpdate.getIsTop() : postStatusToggle.getIsTop());
        statusUpdate.setIsHot(postStatusToggle.getIsHot() == null ? statusUpdate.getIsHot() : postStatusToggle.getIsHot());
        statusUpdate.setIsOriginal(postStatusToggle.getIsOriginal() == null ? statusUpdate.getIsOriginal() : postStatusToggle.getIsOriginal());

        this.baseMapper.updateById(statusUpdate);

        StatusUpdateVO statusUpdateVO = new StatusUpdateVO();
        BeanUtils.copyProperties(statusUpdate, statusUpdateVO);
        statusUpdateVO.setAuthorName(userMapper.selectById(statusUpdate.getAuthorId()).getNickname());
        statusUpdateVO.setId(statusUpdate.getId().toString());
        statusUpdateVO.setCategoryId(statusUpdate.getCategoryId() != null ? statusUpdate.getCategoryId().toString() : null);
        if (statusUpdate.getIsPublished()) {
            socketIOService.broadcastNotification("分享" + statusUpdate.getTitle() + " 状态已更新");
        }
        return statusUpdateVO;
    }

    @Override
    public StatusUpdateVO getStatusUpdateById(Long id) {
        StatusUpdate statusUpdate = this.getById(id);
        StatusUpdateVO statusUpdateVO = new StatusUpdateVO();
        BeanUtils.copyProperties(statusUpdate, statusUpdateVO);
        statusUpdateVO.setAuthorName(userMapper.selectById(statusUpdate.getAuthorId()).getNickname());
        statusUpdateVO.setId(statusUpdate.getId().toString());
        statusUpdateVO.setCategoryId(statusUpdate.getCategoryId() != null ? statusUpdate.getCategoryId().toString() : null);
        return statusUpdateVO;
    }
}
