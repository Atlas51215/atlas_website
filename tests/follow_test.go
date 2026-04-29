package tests

import (
	"testing"

	"github.com/claude/blog/internal/model"
)

// --- Follow model tests ---

func TestFollow_CreateAndCheck(t *testing.T) {
	db := openTestDB(t)
	follower := makeUser(t, db, "follower1")
	curator := makeUser(t, db, "curator1")

	if err := model.Follow(db, follower.ID, curator.ID); err != nil {
		t.Fatalf("Follow: %v", err)
	}

	ok, err := model.IsFollowing(db, follower.ID, curator.ID)
	if err != nil {
		t.Fatalf("IsFollowing: %v", err)
	}
	if !ok {
		t.Error("expected IsFollowing=true after Follow")
	}
}

func TestFollow_Idempotent(t *testing.T) {
	db := openTestDB(t)
	follower := makeUser(t, db, "follower2")
	curator := makeUser(t, db, "curator2")

	if err := model.Follow(db, follower.ID, curator.ID); err != nil {
		t.Fatalf("first Follow: %v", err)
	}
	if err := model.Follow(db, follower.ID, curator.ID); err != nil {
		t.Fatalf("second Follow (should be idempotent): %v", err)
	}
}

func TestUnfollow(t *testing.T) {
	db := openTestDB(t)
	follower := makeUser(t, db, "follower3")
	curator := makeUser(t, db, "curator3")

	if err := model.Follow(db, follower.ID, curator.ID); err != nil {
		t.Fatalf("Follow: %v", err)
	}
	if err := model.Unfollow(db, follower.ID, curator.ID); err != nil {
		t.Fatalf("Unfollow: %v", err)
	}

	ok, err := model.IsFollowing(db, follower.ID, curator.ID)
	if err != nil {
		t.Fatalf("IsFollowing: %v", err)
	}
	if ok {
		t.Error("expected IsFollowing=false after Unfollow")
	}
}

func TestUnfollow_NotFollowing(t *testing.T) {
	db := openTestDB(t)
	follower := makeUser(t, db, "follower4")
	curator := makeUser(t, db, "curator4")

	// Unfollow someone we never followed — should not error
	if err := model.Unfollow(db, follower.ID, curator.ID); err != nil {
		t.Fatalf("Unfollow non-existent: %v", err)
	}
}

func TestIsFollowing_False(t *testing.T) {
	db := openTestDB(t)
	follower := makeUser(t, db, "follower5")
	curator := makeUser(t, db, "curator5")

	ok, err := model.IsFollowing(db, follower.ID, curator.ID)
	if err != nil {
		t.Fatalf("IsFollowing: %v", err)
	}
	if ok {
		t.Error("expected IsFollowing=false when no follow exists")
	}
}

func TestFollowedCurators_Empty(t *testing.T) {
	db := openTestDB(t)
	follower := makeUser(t, db, "follower6")

	curators, err := model.FollowedCurators(db, follower.ID)
	if err != nil {
		t.Fatalf("FollowedCurators: %v", err)
	}
	if len(curators) != 0 {
		t.Errorf("expected 0 curators, got %d", len(curators))
	}
}

func TestFollowedCurators_ReturnsList(t *testing.T) {
	db := openTestDB(t)
	follower := makeUser(t, db, "follower7")
	c1 := makeUser(t, db, "curator7a")
	c2 := makeUser(t, db, "curator7b")

	if err := model.Follow(db, follower.ID, c1.ID); err != nil {
		t.Fatalf("Follow c1: %v", err)
	}
	if err := model.Follow(db, follower.ID, c2.ID); err != nil {
		t.Fatalf("Follow c2: %v", err)
	}

	curators, err := model.FollowedCurators(db, follower.ID)
	if err != nil {
		t.Fatalf("FollowedCurators: %v", err)
	}
	if len(curators) != 2 {
		t.Errorf("expected 2 curators, got %d", len(curators))
	}
}

func TestFollowedCurators_OnlyOwnFollows(t *testing.T) {
	db := openTestDB(t)
	f1 := makeUser(t, db, "follower8a")
	f2 := makeUser(t, db, "follower8b")
	curator := makeUser(t, db, "curator8")

	if err := model.Follow(db, f1.ID, curator.ID); err != nil {
		t.Fatalf("Follow: %v", err)
	}

	curators, err := model.FollowedCurators(db, f2.ID)
	if err != nil {
		t.Fatalf("FollowedCurators: %v", err)
	}
	if len(curators) != 0 {
		t.Errorf("f2 should have 0 followed curators, got %d", len(curators))
	}
}

// --- PostsByFollowedCurators model tests ---

func TestPostsByFollowedCurators_Empty(t *testing.T) {
	db := openTestDB(t)
	follower := makeUser(t, db, "follower9")

	posts, total, err := model.PostsByFollowedCurators(db, follower.ID, 1, 10)
	if err != nil {
		t.Fatalf("PostsByFollowedCurators: %v", err)
	}
	if total != 0 || len(posts) != 0 {
		t.Errorf("expected 0 posts, got total=%d len=%d", total, len(posts))
	}
}

func TestPostsByFollowedCurators_ReturnsFollowedPosts(t *testing.T) {
	db := openTestDB(t)
	follower := makeUser(t, db, "follower10")
	curator := makeUser(t, db, "curator10")
	other := makeUser(t, db, "other10")
	catID := blogCatID(t, db)

	if err := model.Follow(db, follower.ID, curator.ID); err != nil {
		t.Fatalf("Follow: %v", err)
	}
	makePost(t, db, catID, curator.ID, "curator-post", "published")
	makePost(t, db, catID, other.ID, "other-post", "published")

	posts, total, err := model.PostsByFollowedCurators(db, follower.ID, 1, 10)
	if err != nil {
		t.Fatalf("PostsByFollowedCurators: %v", err)
	}
	if total != 1 {
		t.Errorf("expected total=1, got %d", total)
	}
	if len(posts) != 1 {
		t.Fatalf("expected 1 post, got %d", len(posts))
	}
	if posts[0].Slug != "curator-post" {
		t.Errorf("expected curator-post, got %q", posts[0].Slug)
	}
}

func TestPostsByFollowedCurators_ExcludesDrafts(t *testing.T) {
	db := openTestDB(t)
	follower := makeUser(t, db, "follower11")
	curator := makeUser(t, db, "curator11")
	catID := blogCatID(t, db)

	if err := model.Follow(db, follower.ID, curator.ID); err != nil {
		t.Fatalf("Follow: %v", err)
	}
	makePost(t, db, catID, curator.ID, "draft-post", "draft")
	makePost(t, db, catID, curator.ID, "pub-post", "published")

	posts, total, err := model.PostsByFollowedCurators(db, follower.ID, 1, 10)
	if err != nil {
		t.Fatalf("PostsByFollowedCurators: %v", err)
	}
	if total != 1 {
		t.Errorf("expected total=1 (only published), got %d", total)
	}
	if len(posts) != 1 || posts[0].Slug != "pub-post" {
		t.Errorf("expected only pub-post, got %+v", posts)
	}
}

func TestPostsByFollowedCurators_Pagination(t *testing.T) {
	db := openTestDB(t)
	follower := makeUser(t, db, "follower12")
	curator := makeUser(t, db, "curator12")
	catID := blogCatID(t, db)

	if err := model.Follow(db, follower.ID, curator.ID); err != nil {
		t.Fatalf("Follow: %v", err)
	}
	for i := range 5 {
		makePost(t, db, catID, curator.ID, "fp-post-"+string(rune('a'+i)), "published")
	}

	posts, total, err := model.PostsByFollowedCurators(db, follower.ID, 1, 3)
	if err != nil {
		t.Fatalf("page 1: %v", err)
	}
	if total != 5 {
		t.Errorf("expected total=5, got %d", total)
	}
	if len(posts) != 3 {
		t.Errorf("expected 3 on page 1, got %d", len(posts))
	}

	posts2, _, err := model.PostsByFollowedCurators(db, follower.ID, 2, 3)
	if err != nil {
		t.Fatalf("page 2: %v", err)
	}
	if len(posts2) != 2 {
		t.Errorf("expected 2 on page 2, got %d", len(posts2))
	}
}
