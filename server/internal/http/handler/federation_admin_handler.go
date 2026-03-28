package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	appEvent "github.com/grtsinry43/grtblog-v2/server/internal/app/event"
	appfed "github.com/grtsinry43/grtblog-v2/server/internal/app/federation"
	"github.com/grtsinry43/grtblog-v2/server/internal/app/sysconfig"
	"github.com/grtsinry43/grtblog-v2/server/internal/domain/content"
	domainfed "github.com/grtsinry43/grtblog-v2/server/internal/domain/federation"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/contract"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/response"
	fedinfra "github.com/grtsinry43/grtblog-v2/server/internal/infra/federation"
)

type FederationAdminHandler struct {
	cfgSvc        *sysconfig.Service
	contentRepo   content.Repository
	deliverySvc   *appfed.DeliveryService
	instanceRepo  domainfed.FederationInstanceRepository
	postCacheRepo domainfed.FederatedPostCacheRepository
	resolver      *fedinfra.Resolver
	events        appEvent.Bus
}

func NewFederationAdminHandler(cfgSvc *sysconfig.Service, contentRepo content.Repository, deliverySvc *appfed.DeliveryService, instanceRepo domainfed.FederationInstanceRepository, postCacheRepo domainfed.FederatedPostCacheRepository, resolver *fedinfra.Resolver, events appEvent.Bus) *FederationAdminHandler {
	if events == nil {
		events = appEvent.NopBus{}
	}
	return &FederationAdminHandler{
		cfgSvc:        cfgSvc,
		contentRepo:   contentRepo,
		deliverySvc:   deliverySvc,
		instanceRepo:  instanceRepo,
		postCacheRepo: postCacheRepo,
		resolver:      resolver,
		events:        events,
	}
}

// RequestFriendLink 由后台发起对外友链申请。
// @Summary 后台发起友链申请
// @Tags FederationAdmin
// @Accept json
// @Produce json
// @Param request body contract.FederationAdminFriendLinkRequestReq true "友链申请参数"
// @Success 200 {object} contract.FederationAdminProxyResp
// @Security BearerAuth
// @Router /admin/friend-links/federation/request [post]
// @Security JWTAuth
func (h *FederationAdminHandler) RequestFriendLink(c *fiber.Ctx) error {
	var req contract.FederationAdminFriendLinkRequestReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}
	target := strings.TrimSpace(req.TargetURL)
	if target == "" {
		return response.NewBizErrorWithMsg(response.ParamsError, "target_url 不能为空")
	}
	if h.deliverySvc == nil {
		return response.NewBizErrorWithMsg(response.ServerError, "联邦服务未初始化")
	}
	delivery, err := h.deliverySvc.DispatchFriendLink(c.Context(), target, req.Message, req.RSSURL, nil)
	if err != nil {
		return response.NewBizErrorWithCause(response.ServerError, "请求失败", err)
	}
	_ = h.events.Publish(c.Context(), appEvent.Generic{
		EventName: "federation.friendlink.requested",
		At:        time.Now(),
		Payload: map[string]any{
			"TargetURL":   target,
			"StatusCode":  intPtrValue(delivery.HTTPStatus),
			"ResponseRaw": stringPtrValue(delivery.ResponseBody),
		},
	})
	return response.Success(c, contract.FederationAdminProxyResp{
		RequestID:  delivery.RequestID,
		DeliveryID: delivery.ID,
		StatusCode: intPtrValue(delivery.HTTPStatus),
		Body:       stringPtrValue(delivery.ResponseBody),
	})
}

// SendCitation 由后台发起对外引用请求。
// @Summary 后台发起引用请求
// @Tags FederationAdmin
// @Accept json
// @Produce json
// @Param request body contract.FederationAdminCitationReq true "引用请求参数"
// @Success 200 {object} contract.FederationAdminProxyResp
// @Security BearerAuth
// @Router /admin/federation/citations/request [post]
// @Security JWTAuth
func (h *FederationAdminHandler) SendCitation(c *fiber.Ctx) error {
	var req contract.FederationAdminCitationReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}
	target := strings.TrimSpace(req.TargetInstanceURL)
	if target == "" || strings.TrimSpace(req.TargetPostID) == "" {
		return response.NewBizErrorWithMsg(response.ParamsError, "target_instance_url/target_post_id 不能为空")
	}
	if h.deliverySvc == nil {
		return response.NewBizErrorWithMsg(response.ServerError, "联邦服务未初始化")
	}
	article, err := h.resolveArticle(c, req.SourceArticleID, req.SourceShortURL)
	if err != nil {
		if errors.Is(err, content.ErrArticleNotFound) {
			return response.NewBizError(response.NotFound)
		}
		return response.NewBizErrorWithCause(response.ServerError, "文章获取失败", err)
	}
	context := strings.TrimSpace(req.CitationContext)
	if context == "" {
		context = article.Summary
	}
	citationType := strings.TrimSpace(req.CitationType)
	ev := appfed.CitationDetected{
		ArticleID:      article.ID,
		AuthorID:       article.AuthorID,
		Title:          article.Title,
		ShortURL:       article.ShortURL,
		TargetInstance: target,
		TargetPostID:   strings.TrimSpace(req.TargetPostID),
		Context:        context,
		CitationType:   citationType,
	}
	delivery, err := h.deliverySvc.DispatchCitation(c.Context(), ev, nil)
	if err != nil {
		return mapFederationDispatchError(err)
	}
	_ = h.events.Publish(c.Context(), appEvent.Generic{
		EventName: "federation.citation.requested",
		At:        time.Now(),
		Payload: map[string]any{
			"TargetInstanceURL": target,
			"TargetPostID":      ev.TargetPostID,
			"StatusCode":        intPtrValue(delivery.HTTPStatus),
		},
	})
	return response.Success(c, contract.FederationAdminProxyResp{
		RequestID:  delivery.RequestID,
		DeliveryID: delivery.ID,
		StatusCode: intPtrValue(delivery.HTTPStatus),
		Body:       stringPtrValue(delivery.ResponseBody),
	})
}

// SendMention 由后台发起对外提及通知。
// @Summary 后台发起提及通知
// @Tags FederationAdmin
// @Accept json
// @Produce json
// @Param request body contract.FederationAdminMentionReq true "提及通知参数"
// @Success 200 {object} contract.FederationAdminProxyResp
// @Security BearerAuth
// @Router /admin/federation/mentions/notify [post]
// @Security JWTAuth
func (h *FederationAdminHandler) SendMention(c *fiber.Ctx) error {
	var req contract.FederationAdminMentionReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}
	target := strings.TrimSpace(req.TargetInstanceURL)
	if target == "" || strings.TrimSpace(req.MentionedUser) == "" {
		return response.NewBizErrorWithMsg(response.ParamsError, "target_instance_url/mentioned_user 不能为空")
	}
	if h.deliverySvc == nil {
		return response.NewBizErrorWithMsg(response.ServerError, "联邦服务未初始化")
	}
	article, err := h.resolveArticle(c, req.SourceArticleID, req.SourceShortURL)
	if err != nil {
		if errors.Is(err, content.ErrArticleNotFound) {
			return response.NewBizError(response.NotFound)
		}
		return response.NewBizErrorWithCause(response.ServerError, "文章获取失败", err)
	}
	context := strings.TrimSpace(req.MentionContext)
	if context == "" {
		context = article.Summary
	}
	mentionType := strings.TrimSpace(req.MentionType)
	ev := appfed.MentionDetected{
		ArticleID:      article.ID,
		AuthorID:       article.AuthorID,
		Title:          article.Title,
		ShortURL:       article.ShortURL,
		TargetUser:     strings.TrimSpace(req.MentionedUser),
		TargetInstance: target,
		Context:        context,
		MentionType:    mentionType,
	}
	delivery, err := h.deliverySvc.DispatchMention(c.Context(), ev, nil)
	if err != nil {
		return mapFederationDispatchError(err)
	}
	_ = h.events.Publish(c.Context(), appEvent.Generic{
		EventName: "federation.mention.requested",
		At:        time.Now(),
		Payload: map[string]any{
			"TargetInstanceURL": target,
			"MentionedUser":     ev.TargetUser,
			"StatusCode":        intPtrValue(delivery.HTTPStatus),
		},
	})
	return response.Success(c, contract.FederationAdminProxyResp{
		RequestID:  delivery.RequestID,
		DeliveryID: delivery.ID,
		StatusCode: intPtrValue(delivery.HTTPStatus),
		Body:       stringPtrValue(delivery.ResponseBody),
	})
}

// ListOutbound 查询出站投递状态。
// @Summary 查询联合出站状态
// @Tags FederationAdmin
// @Produce json
// @Param type query string false "类型: friendlink|citation|mention"
// @Param request_id query string false "投递单据 request_id（精确匹配）"
// @Param status query string false "状态"
// @Param target query string false "目标实例模糊匹配"
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(20)
// @Success 200 {object} contract.FederationOutboundDeliveryListResp
// @Security BearerAuth
// @Router /admin/federation/outbound [get]
// @Security JWTAuth
func (h *FederationAdminHandler) ListOutbound(c *fiber.Ctx) error {
	if h.deliverySvc == nil {
		return response.NewBizErrorWithMsg(response.ServerError, "联邦服务未初始化")
	}
	page := parseIntQuery(c, "page", 1)
	size := parseIntQuery(c, "pageSize", 20)
	items, total, err := h.deliverySvc.List(c.Context(), domainfed.OutboundDeliveryListOptions{
		RequestID: strings.TrimSpace(c.Query("request_id")),
		Type:      strings.TrimSpace(c.Query("type")),
		Status:    strings.TrimSpace(c.Query("status")),
		Target:    strings.TrimSpace(c.Query("target")),
		Page:      page,
		PageSize:  size,
	})
	if err != nil {
		return err
	}
	respItems := make([]contract.FederationOutboundDeliveryResp, len(items))
	for i := range items {
		respItems[i] = mapOutboundDelivery(items[i])
	}
	return response.Success(c, contract.FederationOutboundDeliveryListResp{
		Items: respItems,
		Total: total,
		Page:  page,
		Size:  size,
	})
}

// GetOutbound 查询单条出站投递。
// @Summary 查询联合出站详情
// @Tags FederationAdmin
// @Produce json
// @Param id path int true "投递ID"
// @Success 200 {object} contract.FederationOutboundDeliveryResp
// @Security BearerAuth
// @Router /admin/federation/outbound/{id} [get]
// @Security JWTAuth
func (h *FederationAdminHandler) GetOutbound(c *fiber.Ctx) error {
	if h.deliverySvc == nil {
		return response.NewBizErrorWithMsg(response.ServerError, "联邦服务未初始化")
	}
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的投递ID")
	}
	item, err := h.deliverySvc.Get(c.Context(), id)
	if err != nil {
		return err
	}
	return response.Success(c, mapOutboundDelivery(*item))
}

// GetOutboundByRequestID 根据 request_id 查询单据。
// @Summary 通过单据号查询联合出站详情
// @Tags FederationAdmin
// @Produce json
// @Param requestId path string true "request_id"
// @Success 200 {object} contract.FederationOutboundDeliveryResp
// @Security BearerAuth
// @Router /admin/federation/outbound/request/{requestId} [get]
// @Security JWTAuth
func (h *FederationAdminHandler) GetOutboundByRequestID(c *fiber.Ctx) error {
	if h.deliverySvc == nil {
		return response.NewBizErrorWithMsg(response.ServerError, "联邦服务未初始化")
	}
	requestID := strings.TrimSpace(c.Params("requestId"))
	if requestID == "" {
		return response.NewBizErrorWithMsg(response.ParamsError, "request_id 不能为空")
	}
	items, total, err := h.deliverySvc.List(c.Context(), domainfed.OutboundDeliveryListOptions{
		RequestID: requestID,
		Page:      1,
		PageSize:  1,
	})
	if err != nil {
		return err
	}
	if total == 0 || len(items) == 0 {
		return response.NewBizError(response.NotFound)
	}
	return response.Success(c, mapOutboundDelivery(items[0]))
}

// RetryOutbound 手动重试出站投递。
// @Summary 重试联合出站投递
// @Tags FederationAdmin
// @Produce json
// @Param id path int true "投递ID"
// @Success 200 {object} contract.FederationOutboundDeliveryResp
// @Security BearerAuth
// @Router /admin/federation/outbound/{id}/retry [post]
// @Security JWTAuth
func (h *FederationAdminHandler) RetryOutbound(c *fiber.Ctx) error {
	if h.deliverySvc == nil {
		return response.NewBizErrorWithMsg(response.ServerError, "联邦服务未初始化")
	}
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的投递ID")
	}
	item, err := h.deliverySvc.Retry(c.Context(), id)
	if err != nil {
		return response.NewBizErrorWithCause(response.ServerError, "重试失败", err)
	}
	return response.Success(c, mapOutboundDelivery(*item))
}

// CheckRemote 校验远端连通性（manifest/public-key/endpoints）。
// @Summary 远端联通性检查
// @Tags FederationAdmin
// @Accept json
// @Produce json
// @Param target_url query string true "远端实例地址"
// @Success 200 {object} contract.FederationAdminRemoteCheckResp
// @Security BearerAuth
// @Router /admin/federation/remote/check [get]
// @Security JWTAuth
func (h *FederationAdminHandler) CheckRemote(c *fiber.Ctx) error {
	target := strings.TrimSpace(c.Query("target_url"))
	if target == "" {
		var req contract.FederationAdminRemoteCheckReq
		if err := c.BodyParser(&req); err == nil {
			target = strings.TrimSpace(req.TargetURL)
		}
	}
	if target == "" {
		return response.NewBizErrorWithMsg(response.ParamsError, "target_url 不能为空")
	}
	if h.resolver == nil {
		return response.NewBizErrorWithMsg(response.ServerError, "resolver 未初始化")
	}
	baseURL := normalizeInstanceURL(target)
	if baseURL == "" {
		return response.NewBizErrorWithMsg(response.ParamsError, "target_url 不能为空")
	}
	manifest, err := h.resolver.FetchManifest(c.Context(), baseURL)
	if err != nil {
		return response.NewBizErrorWithCause(response.ServerError, "拉取 manifest 失败", err)
	}
	publicKey, err := h.resolver.FetchPublicKey(c.Context(), baseURL)
	if err != nil {
		return response.NewBizErrorWithCause(response.ServerError, "拉取公钥失败", err)
	}
	endpoints, err := h.resolver.FetchEndpoints(c.Context(), baseURL)
	if err != nil {
		return response.NewBizErrorWithCause(response.ServerError, "拉取 endpoints 失败", err)
	}
	return response.Success(c, contract.FederationAdminRemoteCheckResp{
		Manifest:  mapManifestResp(manifest),
		PublicKey: mapPublicKeyResp(publicKey),
		Endpoints: mapEndpointsResp(endpoints),
	})
}

// FetchRemotePosts 实时拉取任意远端实例的文章列表，用于引用选择器。
// @Summary 拉取远端文章列表
// @Tags FederationAdmin
// @Produce json
// @Param url query string true "远端实例地址"
// @Param query query string false "搜索关键字"
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(20)
// @Success 200 {object} contract.FederationTimelineResp
// @Security BearerAuth
// @Router /admin/federation/remote/posts [get]
// @Security JWTAuth
func (h *FederationAdminHandler) FetchRemotePosts(c *fiber.Ctx) error {
	target := strings.TrimSpace(c.Query("url"))
	if target == "" {
		return response.NewBizErrorWithMsg(response.ParamsError, "url 不能为空")
	}
	if h.resolver == nil {
		return response.NewBizErrorWithMsg(response.ServerError, "resolver 未初始化")
	}
	baseURL := normalizeInstanceURL(target)
	if baseURL == "" {
		return response.NewBizErrorWithMsg(response.ParamsError, "url 格式无效")
	}

	keyword := strings.TrimSpace(c.Query("query"))
	page, _ := strconv.Atoi(c.Query("page", "1"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(c.Query("pageSize", "20"))
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 50 {
		pageSize = 50
	}

	// 尝试拉取远端 endpoints
	endpoints, err := h.resolver.FetchEndpoints(c.Context(), baseURL)
	if err != nil {
		return response.NewBizErrorWithCause(response.ServerError, "拉取远端 endpoints 失败", err)
	}
	if endpoints == nil || endpoints.Endpoints == nil {
		return response.NewBizErrorWithMsg(response.ServerError, "远端 endpoints 不可用")
	}

	// 构建时间线 URL
	path := strings.TrimSpace(endpoints.Endpoints["timeline"])
	if path == "" {
		return response.NewBizErrorWithMsg(response.ServerError, "远端未提供 timeline 端点")
	}
	endpointBaseURL := strings.TrimSpace(endpoints.BaseURL)
	if endpointBaseURL == "" {
		endpointBaseURL = baseURL
	}
	timelineURL, err := resolveEndpointURL(endpointBaseURL, path)
	if err != nil {
		return response.NewBizErrorWithCause(response.ServerError, "解析 timeline 端点失败", err)
	}
	q := timelineURL.Query()
	q.Set("page", strconv.Itoa(page))
	q.Set("per_page", strconv.Itoa(pageSize))
	timelineURL.RawQuery = q.Encode()

	// 拉取远端时间线
	httpClient := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequestWithContext(c.Context(), http.MethodGet, timelineURL.String(), nil)
	if err != nil {
		return response.NewBizErrorWithCause(response.ServerError, "构建请求失败", err)
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return response.NewBizErrorWithCause(response.ServerError, "拉取远端时间线失败", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return response.NewBizErrorWithMsg(response.ServerError, fmt.Sprintf("远端返回 %d: %s", resp.StatusCode, string(body)))
	}

	var envelope struct {
		Data struct {
			Items []struct {
				ID             string     `json:"id"`
				URL            string     `json:"url"`
				Title          string     `json:"title"`
				Summary        string     `json:"summary"`
				ContentPreview *string    `json:"content_preview"`
				Author         any        `json:"author"`
				PublishedAt    time.Time  `json:"published_at"`
				UpdatedAt      *time.Time `json:"updated_at"`
				CoverImage     *string    `json:"cover_image"`
				Language       *string    `json:"language"`
				AllowCitation  bool       `json:"allow_citation"`
				AllowComment   bool       `json:"allow_comment"`
			} `json:"items"`
			Total int64 `json:"total"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return response.NewBizErrorWithCause(response.ServerError, "解析远端响应失败", err)
	}

	// 获取远端实例名
	instanceName := baseURL
	if manifest, err := h.resolver.FetchManifest(c.Context(), baseURL); err == nil && manifest != nil && manifest.Instance.Name != "" {
		instanceName = manifest.Instance.Name
	}

	items := make([]contract.FederationPostResp, 0, len(envelope.Data.Items))
	keywordLower := strings.ToLower(keyword)
	for _, item := range envelope.Data.Items {
		if strings.TrimSpace(item.URL) == "" || strings.TrimSpace(item.Title) == "" {
			continue
		}
		// 客户端关键字过滤（远端可能不支持搜索参数）
		if keywordLower != "" {
			if !strings.Contains(strings.ToLower(item.Title), keywordLower) &&
				!strings.Contains(strings.ToLower(item.Summary), keywordLower) {
				continue
			}
		}
		id := strings.TrimSpace(item.ID)
		if id == "" {
			id = item.URL
		}
		var authorName string
		var authorAvatar *string
		var authorURL *string
		if raw, err := json.Marshal(item.Author); err == nil {
			var a struct {
				Name   string  `json:"name"`
				Avatar *string `json:"avatar"`
				URL    *string `json:"url"`
			}
			if json.Unmarshal(raw, &a) == nil {
				authorName = a.Name
				authorAvatar = a.Avatar
				authorURL = a.URL
			}
		}
		items = append(items, contract.FederationPostResp{
			ID:             id,
			URL:            item.URL,
			Title:          item.Title,
			Summary:        item.Summary,
			ContentPreview: item.ContentPreview,
			Author: contract.FederationPostAuthorResp{
				Name:   authorName,
				Avatar: authorAvatar,
				URL:    authorURL,
			},
			InstanceName:  instanceName,
			InstanceURL:   baseURL,
			PublishedAt:   item.PublishedAt,
			UpdatedAt:     item.UpdatedAt,
			CoverImage:    item.CoverImage,
			Language:      item.Language,
			AllowCitation: item.AllowCitation,
			AllowComment:  item.AllowComment,
		})
	}

	total := envelope.Data.Total
	if total < int64(len(items)) {
		total = int64(len(items))
	}
	return response.Success(c, contract.FederationTimelineResp{
		Items: items,
		Total: total,
		Page:  page,
		Size:  pageSize,
	})
}

func resolveEndpointURL(base, path string) (*url.URL, error) {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return url.Parse(path)
	}
	parsed, err := url.Parse(strings.TrimSpace(base))
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	parsed.Path = strings.TrimRight(parsed.Path, "/") + path
	return parsed, nil
}

// ListInstances 查询联合实例列表。
// @Summary 查询联合实例列表
// @Tags FederationAdmin
// @Produce json
// @Param status query string false "状态: pending|active|blocked"
// @Param keyword query string false "关键字（base_url/name）"
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(20)
// @Success 200 {object} contract.FederationInstanceListResp
// @Security BearerAuth
// @Router /admin/federation/instances [get]
// @Security JWTAuth
func (h *FederationAdminHandler) ListInstances(c *fiber.Ctx) error {
	if h.instanceRepo == nil {
		return response.NewBizErrorWithMsg(response.ServerError, "联邦服务未初始化")
	}
	page := parseIntQuery(c, "page", 1)
	size := parseIntQuery(c, "pageSize", 20)
	status := strings.TrimSpace(c.Query("status"))
	keyword := strings.TrimSpace(c.Query("keyword"))
	items, total, err := h.instanceRepo.List(c.Context(), status, keyword, page, size)
	if err != nil {
		return err
	}
	respItems := make([]contract.FederationInstanceResp, len(items))
	for i := range items {
		respItems[i] = mapFederationInstance(items[i])
	}
	return response.Success(c, contract.FederationInstanceListResp{
		Items: respItems,
		Total: total,
		Page:  page,
		Size:  size,
	})
}

// GetInstance 查询联合实例详情。
// @Summary 查询联合实例详情
// @Tags FederationAdmin
// @Produce json
// @Param id path int true "实例ID"
// @Param refresh query bool false "是否实时拉取远端文档并刷新本地快照（默认 true）"
// @Success 200 {object} contract.FederationInstanceDetailResp
// @Security BearerAuth
// @Router /admin/federation/instances/{id} [get]
// @Security JWTAuth
func (h *FederationAdminHandler) GetInstance(c *fiber.Ctx) error {
	if h.instanceRepo == nil {
		return response.NewBizErrorWithMsg(response.ServerError, "联邦服务未初始化")
	}
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的实例ID")
	}
	instance, err := h.instanceRepo.GetByID(c.Context(), id)
	if err != nil {
		return err
	}
	refresh := true
	if raw := strings.TrimSpace(c.Query("refresh")); raw != "" {
		if parsed, parseErr := strconv.ParseBool(raw); parseErr == nil {
			refresh = parsed
		}
	}
	var (
		manifest  any
		publicKey any
		endpoints any
		remoteErr *string
	)
	if refresh && h.resolver != nil {
		baseURL := normalizeInstanceURL(instance.BaseURL)
		if baseURL != "" {
			m, errM := h.resolver.FetchManifest(c.Context(), baseURL)
			p, errP := h.resolver.FetchPublicKey(c.Context(), baseURL)
			e, errE := h.resolver.FetchEndpoints(c.Context(), baseURL)
			if errM == nil && errP == nil && errE == nil {
				manifest = mapManifestResp(m)
				publicKey = mapPublicKeyResp(p)
				endpoints = mapEndpointsResp(e)
				if updated, err := ensureInstanceFromDocs(c.Context(), baseURL, m, e, p, h.instanceRepo); err == nil && updated != nil {
					instance = updated
				}
			} else {
				msg := "拉取远端实例详情失败"
				if errM != nil {
					msg += ": manifest"
				} else if errP != nil {
					msg += ": public_key"
				} else if errE != nil {
					msg += ": endpoints"
				}
				remoteErr = &msg
			}
		}
	}
	return response.Success(c, mapFederationInstanceDetail(*instance, manifest, publicKey, endpoints, remoteErr))
}

// UpdateInstanceStatus 更新联合实例状态。
// @Summary 更新联合实例状态
// @Tags FederationAdmin
// @Accept json
// @Produce json
// @Param id path int true "实例ID"
// @Param request body contract.FederationInstanceStatusUpdateReq true "状态更新"
// @Success 200 {object} contract.FederationInstanceResp
// @Security BearerAuth
// @Router /admin/federation/instances/{id}/status [put]
// @Security JWTAuth
func (h *FederationAdminHandler) UpdateInstanceStatus(c *fiber.Ctx) error {
	if h.instanceRepo == nil {
		return response.NewBizErrorWithMsg(response.ServerError, "联邦服务未初始化")
	}
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的实例ID")
	}
	var req contract.FederationInstanceStatusUpdateReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}
	status := normalizeInstanceStatus(req.Status)
	if status == "" {
		return response.NewBizErrorWithMsg(response.ParamsError, "status 仅支持 pending|active|blocked")
	}
	instance, err := h.instanceRepo.GetByID(c.Context(), id)
	if err != nil {
		return err
	}
	instance.Status = status
	if err := h.instanceRepo.Update(c.Context(), instance); err != nil {
		return err
	}
	return response.Success(c, mapFederationInstance(*instance))
}

// ListInstancePosts 按实例浏览/搜索缓存文章。
// @Summary 按实例搜索缓存文章
// @Tags FederationAdmin
// @Produce json
// @Param id path int true "实例ID"
// @Param q query string false "搜索关键词"
// @Param limit query int false "返回数量" default(20)
// @Success 200 {object} contract.FederationCachedPostListResp
// @Security BearerAuth
// @Router /admin/federation/instances/{id}/posts [get]
// @Security JWTAuth
func (h *FederationAdminHandler) ListInstancePosts(c *fiber.Ctx) error {
	if h.postCacheRepo == nil {
		return response.NewBizErrorWithMsg(response.ServerError, "联邦服务未初始化")
	}
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的实例ID")
	}
	keyword := strings.TrimSpace(c.Query("q"))
	limit := parseIntQuery(c, "limit", 20)
	posts, err := h.postCacheRepo.SearchPostsByInstance(c.Context(), id, keyword, limit)
	if err != nil {
		return err
	}
	items := make([]contract.FederationCachedPostResp, len(posts))
	for i, p := range posts {
		authorName := ""
		if len(p.Author) > 0 {
			var author struct {
				Name string `json:"name"`
			}
			if jsonErr := json.Unmarshal(p.Author, &author); jsonErr == nil {
				authorName = author.Name
			}
		}
		items[i] = contract.FederationCachedPostResp{
			ID:            p.ID,
			RemotePostID:  p.RemotePostID,
			InstanceID:    int64PtrValue(p.InstanceID),
			URL:           p.URL,
			Title:         p.Title,
			Summary:       p.Summary,
			CoverImage:    p.CoverImage,
			AuthorName:    authorName,
			PublishedAt:   p.PublishedAt.UTC().Format(time.RFC3339),
			AllowCitation: p.AllowCitation,
		}
	}
	return response.Success(c, contract.FederationCachedPostListResp{Items: items})
}

// SearchAuthors 搜索缓存作者（mention 用）。
// @Summary 搜索缓存作者
// @Tags FederationAdmin
// @Produce json
// @Param q query string false "搜索关键词"
// @Param limit query int false "返回数量" default(20)
// @Success 200 {object} contract.FederationAuthorListResp
// @Security BearerAuth
// @Router /admin/federation/authors/search [get]
// @Security JWTAuth
func (h *FederationAdminHandler) SearchAuthors(c *fiber.Ctx) error {
	if h.postCacheRepo == nil {
		return response.NewBizErrorWithMsg(response.ServerError, "联邦服务未初始化")
	}
	keyword := strings.TrimSpace(c.Query("q"))
	limit := parseIntQuery(c, "limit", 20)
	authors, err := h.postCacheRepo.SearchAuthors(c.Context(), keyword, limit)
	if err != nil {
		return err
	}
	items := make([]contract.FederationAuthorResp, len(authors))
	for i, a := range authors {
		items[i] = contract.FederationAuthorResp{
			Name:         a.Name,
			InstanceURL:  a.InstanceURL,
			InstanceName: a.InstanceName,
		}
	}
	return response.Success(c, contract.FederationAuthorListResp{Items: items})
}

func (h *FederationAdminHandler) resolveArticle(c *fiber.Ctx, id *int64, shortURL *string) (*content.Article, error) {
	if id != nil && *id > 0 {
		return h.contentRepo.GetArticleByID(c.Context(), *id)
	}
	if shortURL != nil && strings.TrimSpace(*shortURL) != "" {
		return h.contentRepo.GetArticleByShortURL(c.Context(), strings.TrimSpace(*shortURL))
	}
	return nil, content.ErrArticleNotFound
}

func mapFederationDispatchError(err error) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, appfed.ErrTargetFriendLinkNotFederation):
		return response.NewBizErrorWithMsg(response.ParamsError, appfed.ErrTargetFriendLinkNotFederation.Error())
	case errors.Is(err, appfed.ErrFederationInstanceNotBound):
		return response.NewBizErrorWithMsg(response.ParamsError, appfed.ErrFederationInstanceNotBound.Error())
	case errors.Is(err, appfed.ErrFederationInstanceNotActive):
		return response.NewBizErrorWithMsg(response.ParamsError, appfed.ErrFederationInstanceNotActive.Error())
	case errors.Is(err, appfed.ErrTargetInstanceEmpty):
		return response.NewBizErrorWithMsg(response.ParamsError, appfed.ErrTargetInstanceEmpty.Error())
	default:
		return response.NewBizErrorWithCause(response.ServerError, "请求失败", err)
	}
}

func mapManifestResp(manifest *fedinfra.Manifest) map[string]any {
	if manifest == nil {
		return nil
	}
	return map[string]any{
		"protocol_version": manifest.ProtocolVersion,
		"instance": map[string]any{
			"name":        manifest.Instance.Name,
			"url":         manifest.Instance.URL,
			"description": manifest.Instance.Description,
			"language":    manifest.Instance.Language,
			"timezone":    manifest.Instance.Timezone,
		},
		"software": map[string]any{
			"name":    manifest.Software.Name,
			"version": manifest.Software.Version,
		},
		"features": manifest.Features,
		"policies": map[string]any{
			"allow_citation":                   manifest.Policies.AllowCitation,
			"allow_mention":                    manifest.Policies.AllowMention,
			"auto_approve_friendlink_citation": manifest.Policies.AutoApproveFriendlinkCitation,
			"require_https":                    manifest.Policies.RequireHTTPS,
			"max_cache_age":                    manifest.Policies.MaxCacheAge,
		},
		"created_at": manifest.CreatedAt.Format(time.RFC3339),
		"updated_at": manifest.UpdatedAt.Format(time.RFC3339),
	}
}

func mapPublicKeyResp(doc *fedinfra.PublicKeyDoc) map[string]any {
	if doc == nil {
		return nil
	}
	return map[string]any{
		"key_id":     doc.KeyID,
		"algorithm":  doc.Algorithm,
		"public_key": doc.PublicKey,
	}
}

func mapEndpointsResp(doc *fedinfra.EndpointsDoc) map[string]any {
	if doc == nil {
		return nil
	}
	return map[string]any{
		"base_url":  doc.BaseURL,
		"endpoints": doc.Endpoints,
	}
}

func mapOutboundDelivery(item domainfed.OutboundDelivery) contract.FederationOutboundDeliveryResp {
	return contract.FederationOutboundDeliveryResp{
		ID:                item.ID,
		RequestID:         item.RequestID,
		Type:              item.DeliveryType,
		SourceArticleID:   item.SourceArticleID,
		TargetInstanceURL: item.TargetInstanceURL,
		TargetEndpoint:    item.TargetEndpoint,
		Status:            item.Status,
		AttemptCount:      item.AttemptCount,
		MaxAttempts:       item.MaxAttempts,
		NextRetryAt:       timePtrToRFC3339(item.NextRetryAt),
		HTTPStatus:        item.HTTPStatus,
		ResponseBody:      item.ResponseBody,
		ErrorMessage:      item.ErrorMessage,
		RemoteTicketID:    item.RemoteTicketID,
		TraceID:           item.TraceID,
		LastCallbackAt:    timePtrToRFC3339(item.LastCallbackAt),
		CreatedAt:         item.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:         item.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func mapFederationInstance(item domainfed.FederationInstance) contract.FederationInstanceResp {
	return contract.FederationInstanceResp{
		ID:              item.ID,
		BaseURL:         item.BaseURL,
		Name:            item.Name,
		Description:     item.Description,
		ProtocolVersion: item.ProtocolVersion,
		KeyID:           item.KeyID,
		Status:          item.Status,
		LastSeenAt:      timePtrToRFC3339(item.LastSeenAt),
		CreatedAt:       item.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:       item.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func mapFederationInstanceDetail(item domainfed.FederationInstance, manifest any, publicKey any, endpoints any, remoteErr *string) contract.FederationInstanceDetailResp {
	return contract.FederationInstanceDetailResp{
		ID:              item.ID,
		BaseURL:         item.BaseURL,
		Name:            item.Name,
		Description:     item.Description,
		ProtocolVersion: item.ProtocolVersion,
		KeyID:           item.KeyID,
		PublicKey:       item.PublicKey,
		Status:          item.Status,
		Features:        jsonRawToAny(item.Features),
		Policies:        jsonRawToAny(item.Policies),
		Endpoints:       jsonRawToAny(item.Endpoints),
		Manifest:        manifest,
		PublicKeyDoc:    publicKey,
		EndpointsDoc:    endpoints,
		RemoteError:     remoteErr,
		LastSeenAt:      timePtrToRFC3339(item.LastSeenAt),
		CreatedAt:       item.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:       item.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func jsonRawToAny(raw json.RawMessage) any {
	if len(raw) == 0 {
		return nil
	}
	var payload any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil
	}
	return payload
}

func normalizeInstanceStatus(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "pending":
		return "pending"
	case "active":
		return "active"
	case "blocked":
		return "blocked"
	default:
		return ""
	}
}

func timePtrToRFC3339(t *time.Time) *string {
	if t == nil {
		return nil
	}
	val := t.UTC().Format(time.RFC3339)
	return &val
}

func intPtrValue(v *int) int {
	if v == nil {
		return 0
	}
	return *v
}

func stringPtrValue(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func int64PtrValue(v *int64) int64 {
	if v == nil {
		return 0
	}
	return *v
}
