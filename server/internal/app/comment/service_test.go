package comment

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	domaincomment "github.com/grtsinry43/grtblog-v2/server/internal/domain/comment"
	"github.com/grtsinry43/grtblog-v2/server/internal/domain/identity"
)

type commentRepositoryStub struct {
	domaincomment.CommentRepository
	getAreaByIDFn                func(context.Context, int64) (*domaincomment.CommentArea, error)
	findByIDFn                   func(context.Context, int64) (*domaincomment.Comment, error)
	listPublicRootsByAreaIDFn    func(context.Context, domaincomment.PublicListOptions, int, int) ([]*domaincomment.Comment, int64, error)
	listPublicRepliesByRootIDsFn func(context.Context, domaincomment.PublicListOptions, []int64) ([]*domaincomment.Comment, error)
	createFn                     func(context.Context, *domaincomment.Comment) error
	existsBlockedIdentityFn      func(context.Context, *int64, *string) (bool, error)
}

func (s *commentRepositoryStub) GetAreaByID(ctx context.Context, id int64) (*domaincomment.CommentArea, error) {
	return s.getAreaByIDFn(ctx, id)
}

func (s *commentRepositoryStub) FindByID(ctx context.Context, id int64) (*domaincomment.Comment, error) {
	return s.findByIDFn(ctx, id)
}

func (s *commentRepositoryStub) ListPublicRootsByAreaID(ctx context.Context, options domaincomment.PublicListOptions, page, pageSize int) ([]*domaincomment.Comment, int64, error) {
	return s.listPublicRootsByAreaIDFn(ctx, options, page, pageSize)
}

func (s *commentRepositoryStub) ListPublicRepliesByRootIDs(ctx context.Context, options domaincomment.PublicListOptions, rootIDs []int64) ([]*domaincomment.Comment, error) {
	return s.listPublicRepliesByRootIDsFn(ctx, options, rootIDs)
}

func (s *commentRepositoryStub) Create(ctx context.Context, item *domaincomment.Comment) error {
	return s.createFn(ctx, item)
}

func (s *commentRepositoryStub) ExistsBlockedIdentity(ctx context.Context, authorID *int64, email *string) (bool, error) {
	if s.existsBlockedIdentityFn == nil {
		return false, nil
	}
	return s.existsBlockedIdentityFn(ctx, authorID, email)
}

type identityRepositoryStub struct {
	identity.Repository
	findByIDFn func(context.Context, int64) (*identity.User, error)
}

func (s *identityRepositoryStub) FindByID(ctx context.Context, id int64) (*identity.User, error) {
	return s.findByIDFn(ctx, id)
}

func TestBuildCommentThreadsFlattensArbitraryReplyDepth(t *testing.T) {
	t.Parallel()

	base := time.Date(2026, time.July, 14, 10, 0, 0, 0, time.UTC)
	rootOne := testComment(1, 1, nil, 1, 1, "root-one", base)
	rootTwo := testComment(2, 1, nil, 2, 1, "root-two", base.Add(time.Minute))
	rootTwo.IsTop = true
	rootOne.Floor = 1
	rootTwo.Floor = 2

	parentOne := int64(1)
	parentThree := int64(3)
	replyToRoot := testComment(3, 1, &parentOne, 1, 2, "reply-one", base.Add(2*time.Minute))
	replyToReply := testComment(4, 1, &parentThree, 1, 3, "reply-two", base.Add(3*time.Minute))

	threads := buildCommentThreads(
		[]*domaincomment.Comment{rootTwo, rootOne},
		[]*domaincomment.Comment{replyToReply, replyToRoot},
	)
	for _, thread := range threads {
		assignChildFloors(thread)
	}

	if len(threads) != 2 {
		t.Fatalf("expected 2 threads, got %d", len(threads))
	}
	if threads[0].Comment.ID != rootTwo.ID || threads[0].Floor != "2" {
		t.Fatalf("expected pinned newer root to display first with chronological floor 2, got id=%d floor=%q", threads[0].Comment.ID, threads[0].Floor)
	}

	thread := threads[1]
	if thread.Comment.ID != rootOne.ID || thread.Floor != "1" {
		t.Fatalf("expected first chronological root to retain floor 1, got id=%d floor=%q", thread.Comment.ID, thread.Floor)
	}
	if len(thread.Children) != 2 {
		t.Fatalf("expected both replies directly under the root, got %d", len(thread.Children))
	}
	if thread.Children[0].Comment.ID != replyToRoot.ID || thread.Children[0].Floor != "1-1" {
		t.Fatalf("unexpected first reply: id=%d floor=%q", thread.Children[0].Comment.ID, thread.Children[0].Floor)
	}
	if thread.Children[1].Comment.ID != replyToReply.ID || thread.Children[1].Floor != "1-2" {
		t.Fatalf("unexpected second reply: id=%d floor=%q", thread.Children[1].Comment.ID, thread.Children[1].Floor)
	}
	if thread.Children[1].Comment.ParentID == nil || *thread.Children[1].Comment.ParentID != replyToRoot.ID {
		t.Fatalf("exact reply target was not preserved: parent=%v", thread.Children[1].Comment.ParentID)
	}
	if got := valueOrEmpty(thread.Children[1].Comment.ReplyToNickName); got != "reply-one" {
		t.Fatalf("expected reply target nickname %q, got %q", "reply-one", got)
	}
	if len(thread.Children[0].Children) != 0 || len(thread.Children[1].Children) != 0 {
		t.Fatal("reply nodes must never contain recursive children")
	}
}

func TestEnsureParentValidUsesPersistedDepthBoundary(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		parent  *domaincomment.Comment
		areaID  int64
		wantErr error
	}{
		{
			name:   "ninth level can receive the tenth reply",
			parent: &domaincomment.Comment{ID: 9, AreaID: 1, RootID: 1, Depth: 9, CanReply: true},
			areaID: 1,
		},
		{
			name:    "tenth level rejects another reply",
			parent:  &domaincomment.Comment{ID: 10, AreaID: 1, RootID: 1, Depth: 10, CanReply: true},
			areaID:  1,
			wantErr: domaincomment.ErrCommentTooDeep,
		},
		{
			name:    "cross-area parent is rejected",
			parent:  &domaincomment.Comment{ID: 3, AreaID: 2, RootID: 3, Depth: 1, CanReply: true},
			areaID:  1,
			wantErr: domaincomment.ErrCommentParentNotFound,
		},
		{
			name:    "reply-disabled parent is rejected",
			parent:  &domaincomment.Comment{ID: 3, AreaID: 1, RootID: 3, Depth: 1, CanReply: false},
			areaID:  1,
			wantErr: domaincomment.ErrCommentReplyDisabled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			findCalls := 0
			repo := &commentRepositoryStub{
				findByIDFn: func(context.Context, int64) (*domaincomment.Comment, error) {
					findCalls++
					return tt.parent, nil
				},
			}
			svc := &Service{repo: repo}

			got, err := svc.ensureParentValid(context.Background(), tt.areaID, tt.parent.ID)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}
			if tt.wantErr == nil && got != tt.parent {
				t.Fatal("expected validated parent to be returned")
			}
			if findCalls != 1 {
				t.Fatalf("depth validation must be O(1); expected one lookup, got %d", findCalls)
			}
		})
	}
}

func TestCreateCommentLoginDerivesRootAndDepthWithoutChangingReplyTarget(t *testing.T) {
	t.Parallel()

	parentID := int64(90)
	parent := &domaincomment.Comment{
		ID:       parentID,
		AreaID:   7,
		RootID:   11,
		Depth:    9,
		CanReply: true,
	}
	var created *domaincomment.Comment
	repo := &commentRepositoryStub{
		getAreaByIDFn: func(context.Context, int64) (*domaincomment.CommentArea, error) {
			return &domaincomment.CommentArea{ID: 7}, nil
		},
		findByIDFn: func(context.Context, int64) (*domaincomment.Comment, error) {
			return parent, nil
		},
		createFn: func(_ context.Context, item *domaincomment.Comment) error {
			item.ID = 100
			created = item
			return nil
		},
	}
	users := &identityRepositoryStub{
		findByIDFn: func(context.Context, int64) (*identity.User, error) {
			return &identity.User{ID: 5, Username: "owner", Email: "owner@example.com", IsAdmin: true}, nil
		},
	}
	svc := NewService(repo, users, nil, nil, nil, nil, nil)

	got, err := svc.CreateCommentLogin(context.Background(), 5, CreateCommentLoginCmd{
		AreaID:   7,
		Content:  "level ten",
		ParentID: &parentID,
	}, RequestMeta{})
	if err != nil {
		t.Fatalf("create reply: %v", err)
	}
	if got != created {
		t.Fatal("service did not return the created comment")
	}
	if got.ParentID == nil || *got.ParentID != parentID {
		t.Fatalf("exact reply target changed: %v", got.ParentID)
	}
	if got.RootID != parent.RootID {
		t.Fatalf("expected root %d, got %d", parent.RootID, got.RootID)
	}
	if got.Depth != 10 {
		t.Fatalf("expected depth 10, got %d", got.Depth)
	}
}

func TestListPublicCommentsPaginatesRootsBeforeLoadingReplies(t *testing.T) {
	t.Parallel()

	base := time.Date(2026, time.July, 14, 10, 0, 0, 0, time.UTC)
	rootThree := testComment(3, 8, nil, 3, 1, "three", base.Add(2*time.Minute))
	rootThree.IsTop = true
	rootThree.Floor = 3
	parentThree := int64(3)
	reply := testComment(30, 8, &parentThree, 3, 2, "reply", base.Add(3*time.Minute))

	var requestedRootIDs []int64
	repo := &commentRepositoryStub{
		getAreaByIDFn: func(context.Context, int64) (*domaincomment.CommentArea, error) {
			return &domaincomment.CommentArea{ID: 8}, nil
		},
		listPublicRootsByAreaIDFn: func(_ context.Context, _ domaincomment.PublicListOptions, page, pageSize int) ([]*domaincomment.Comment, int64, error) {
			if page != 1 || pageSize != 1 {
				t.Fatalf("unexpected root pagination: page=%d size=%d", page, pageSize)
			}
			return []*domaincomment.Comment{rootThree}, 3, nil
		},
		listPublicRepliesByRootIDsFn: func(_ context.Context, _ domaincomment.PublicListOptions, rootIDs []int64) ([]*domaincomment.Comment, error) {
			requestedRootIDs = append([]int64(nil), rootIDs...)
			return []*domaincomment.Comment{reply}, nil
		},
	}
	svc := &Service{repo: repo}

	page, err := svc.ListPublicComments(context.Background(), ListPublicCommentsCmd{
		AreaID:   8,
		Page:     1,
		PageSize: 1,
	})
	if err != nil {
		t.Fatalf("list comments: %v", err)
	}
	if page.Total != 3 || len(page.Items) != 1 {
		t.Fatalf("unexpected page: total=%d items=%d", page.Total, len(page.Items))
	}
	if !reflect.DeepEqual(requestedRootIDs, []int64{3}) {
		t.Fatalf("replies must be loaded only for paged roots; got %v", requestedRootIDs)
	}
	if len(page.Items[0].Children) != 1 || page.Items[0].Children[0].Comment.ID != reply.ID {
		t.Fatalf("reply was not attached to selected root: %#v", page.Items[0].Children)
	}
	if page.Items[0].Floor != "3" || page.Items[0].Children[0].Floor != "3-1" {
		t.Fatalf("database root floor was not preserved: root=%q reply=%q", page.Items[0].Floor, page.Items[0].Children[0].Floor)
	}
}

func testComment(id, areaID int64, parentID *int64, rootID int64, depth int16, nickname string, createdAt time.Time) *domaincomment.Comment {
	return &domaincomment.Comment{
		ID:        id,
		AreaID:    areaID,
		ParentID:  parentID,
		RootID:    rootID,
		Depth:     depth,
		NickName:  &nickname,
		CreatedAt: createdAt,
	}
}

func valueOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
