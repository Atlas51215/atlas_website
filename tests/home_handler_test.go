package tests

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/claude/blog/internal/handler"
	"github.com/claude/blog/internal/model"
)

func TestHome_VisitorSeesRecentPosts(t *testing.T) {
	db := openTestDB(t)
	u := makeUser(t, db, "homeauthor1")
	catID := blogCatID(t, db)
	p := makePost(t, db, catID, u.ID, "visible-post", "published")

	srv := httptest.NewServer(handler.NewRouter(db, "static"))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	body := readBody(t, resp)
	if !strings.Contains(body, p.Title) {
		t.Errorf("visitor should see post title %q", p.Title)
	}
	if !strings.Contains(body, "Recent Posts") {
		t.Error("visitor should see 'Recent Posts' heading")
	}
}

func TestHome_VisitorDoesNotSeeFollowHint(t *testing.T) {
	db := openTestDB(t)
	srv := httptest.NewServer(handler.NewRouter(db, "static"))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	body := readBody(t, resp)
	if strings.Contains(body, "Follow curators") {
		t.Error("visitor should not see follow-curators hint")
	}
}

func TestHome_VisitorDoesNotSeeDrafts(t *testing.T) {
	db := openTestDB(t)
	u := makeUser(t, db, "homeauthor2")
	catID := blogCatID(t, db)
	makePost(t, db, catID, u.ID, "draft-only", "draft")

	srv := httptest.NewServer(handler.NewRouter(db, "static"))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	body := readBody(t, resp)
	if strings.Contains(body, "draft-only") {
		t.Error("visitor should not see draft posts")
	}
}

func TestHome_LoggedInNoFollows_SeesAllPostsAndHint(t *testing.T) {
	db := openTestDB(t)
	author := makeUser(t, db, "homeauthor3")
	loggedIn := makeUser(t, db, "homefollower3")
	catID := blogCatID(t, db)
	p := makePost(t, db, catID, author.ID, "all-posts-visible", "published")

	// Create a session for loggedIn user
	token, err := model.CreateSession(db, loggedIn.ID)
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	srv := httptest.NewServer(handler.NewRouter(db, "static"))
	defer srv.Close()

	req, _ := http.NewRequest("GET", srv.URL+"/", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: token})

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	body := readBody(t, resp)
	if !strings.Contains(body, p.Title) {
		t.Errorf("logged-in user with no follows should see all posts, missing %q", p.Title)
	}
	if !strings.Contains(body, "Follow curators") {
		t.Error("user with no follows should see follow-curators hint")
	}
	if !strings.Contains(body, "Recent Posts") {
		t.Error("user with no follows should see 'Recent Posts' heading")
	}
}

func TestHome_LoggedInWithFollows_SeesOnlyFollowedPosts(t *testing.T) {
	db := openTestDB(t)
	curator := makeUser(t, db, "homecurator4")
	other := makeUser(t, db, "homeother4")
	loggedIn := makeUser(t, db, "homefollower4")
	catID := blogCatID(t, db)

	curatorPost := makePost(t, db, catID, curator.ID, "curator-post-4", "published")
	otherPost := makePost(t, db, catID, other.ID, "other-post-4", "published")

	if err := model.Follow(db, loggedIn.ID, curator.ID); err != nil {
		t.Fatalf("Follow: %v", err)
	}

	token, err := model.CreateSession(db, loggedIn.ID)
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	srv := httptest.NewServer(handler.NewRouter(db, "static"))
	defer srv.Close()

	req, _ := http.NewRequest("GET", srv.URL+"/", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: token})

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	body := readBody(t, resp)
	// Check by slug (appears in the post card's href) since all posts share the same default title.
	if !strings.Contains(body, curatorPost.Slug) {
		t.Errorf("should see followed curator's post with slug %q", curatorPost.Slug)
	}
	if strings.Contains(body, otherPost.Slug) {
		t.Errorf("should not see unfollowed author's post with slug %q", otherPost.Slug)
	}
	if !strings.Contains(body, "Your Feed") {
		t.Error("user following curators should see 'Your Feed' heading")
	}
}

func TestHome_LoggedInWithFollows_NoFollowHint(t *testing.T) {
	db := openTestDB(t)
	curator := makeUser(t, db, "homecurator5")
	loggedIn := makeUser(t, db, "homefollower5")

	if err := model.Follow(db, loggedIn.ID, curator.ID); err != nil {
		t.Fatalf("Follow: %v", err)
	}

	token, err := model.CreateSession(db, loggedIn.ID)
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	srv := httptest.NewServer(handler.NewRouter(db, "static"))
	defer srv.Close()

	req, _ := http.NewRequest("GET", srv.URL+"/", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: token})

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	body := readBody(t, resp)
	if strings.Contains(body, "Follow curators") {
		t.Error("user already following curators should not see follow hint")
	}
}

func TestHome_HTMX_ReturnsFragment(t *testing.T) {
	db := openTestDB(t)
	srv := httptest.NewServer(handler.NewRouter(db, "static"))
	defer srv.Close()

	req, _ := http.NewRequest("GET", srv.URL+"/", nil)
	req.Header.Set("HX-Request", "true")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	body := readBody(t, resp)
	if strings.Contains(body, "<html") {
		t.Error("HTMX response should not contain <html> wrapper")
	}
}
