"use client";

// =====================================================
// IMPORTS
// =====================================================

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";

import { request } from "@/lib/api";
import { supabase } from "@/lib/supabase";

import {
  addReaction,
  addInterest,
  createComment,
  createPost,
  getComments,
  getInnerCircleFeed,
  getDiscoveryFeed,
  getInterests,
  getMyReactions,
  removeReaction,
  getMyInterests,
  removeInterest,
} from "../lib/api";

// =====================================================
// TYPES
// =====================================================

type User = {
  id: string;
  username: string;
  display_name: string;
  avatar_url: string;
  bio: string;
};

type Post = {
  id: string;
  author_id: string;
  username: string;
  display_name: string;
  avatar_url: string;
  content: string;
  visibility: string;
  created_at: string;
};

type Interest = {
  id: string;
  name: string;
  slug: string;
  description: string;
};

type Comment = {
  id: string;
  post_id: string;
  author_id: string;
  username: string;
  display_name: string;
  content: string;
  created_at: string;
};

type ReactionType = "like" | "support" | "helpful";

// =====================================================
// MAIN COMPONENT
// =====================================================

export default function Home() {
  const router = useRouter();

  // ===================================================
  // USER & AUTH STATE
  // ===================================================

  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  // ===================================================
  // SPACE & FEED STATE
  // ===================================================

  const [space, setSpace] = useState<"inner_circle" | "discovery">(
    "inner_circle",
  );

  const [posts, setPosts] = useState<Post[]>([]);

  // ===================================================
  // POST CREATION STATE
  // ===================================================

  const [postText, setPostText] = useState("");
  const [posting, setPosting] = useState(false);

  // ===================================================
  // COMMENTS STATE
  // ===================================================

  const [openComments, setOpenComments] = useState<string | null>(null);

  const [comments, setComments] = useState<Record<string, Comment[]>>({});

  const [commentText, setCommentText] = useState("");

  // ===================================================
  // REACTIONS STATE
  // ===================================================

  const [reactionMenu, setReactionMenu] = useState<string | null>(null);

  const [myReactions, setMyReactions] = useState<
    Record<string, ReactionType[]>
  >({});

  // ===================================================
  // INTERESTS STATE
  // ===================================================

  const [interests, setInterests] = useState<Interest[]>([]);

  const [myInterestIds, setMyInterestIds] = useState<string[]>([]);

  const [showInterestPicker, setShowInterestPicker] = useState(false);

  const [loadingInterests, setLoadingInterests] = useState(false);

  // ===================================================
  // PROFILE EDITOR STATE
  // ===================================================

  const [showProfileEditor, setShowProfileEditor] = useState(false);

  const [editDisplayName, setEditDisplayName] = useState("");

  const [editBio, setEditBio] = useState("");

  const [editAvatarURL, setEditAvatarURL] = useState("");

  const [savingProfile, setSavingProfile] = useState(false);

  // ===================================================
  // AUTH & INITIAL PAGE LOADING
  // ===================================================

  useEffect(() => {
    async function checkSupabaseUser() {
      const {
        data: { user },
      } = await supabase.auth.getUser();

      console.log("Supabase ID:", user?.id);
      console.log("Supabase email:", user?.email);
    }

    checkSupabaseUser();
  }, []);

  async function loadPage() {
    const {
      data: { session },
    } = await supabase.auth.getSession();

    if (!session) {
      router.push("/login");
      return;
    }

    try {
      const [profileData, feedData] = await Promise.all([
        request("/api/v1/me"),
        request("/api/v1/feeds/inner-circle"),
      ]);

      // Save authenticated user's profile
      setUser(profileData);

      // Save posts for the default Inner Circle feed
      setPosts(feedData.posts || []);
    } catch (error) {
      console.error("Failed to load page data:", error);
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    loadPage();
    loadInterests();
  }, []);

  async function handleLogout() {
    await supabase.auth.signOut();

    router.push("/login");
  }

  // ===================================================
  // PROFILE FUNCTIONS
  // ===================================================

  function openProfileEditor() {
    if (!user) {
      return;
    }

    setEditDisplayName(user.display_name);
    setEditBio(user.bio);
    setEditAvatarURL(user.avatar_url);

    setShowProfileEditor(true);
  }

  async function handleProfileUpdate(
    displayName: string,
    bio: string,
    avatarURL: string,
  ) {
    try {
      setSavingProfile(true);
      setError("");

      const updatedUser = await request("/api/v1/me", {
        method: "PATCH",
        body: JSON.stringify({
          display_name: displayName,
          bio: bio,
          avatar_url: avatarURL,
        }),
      });

      setUser(updatedUser);

      setShowProfileEditor(false);
    } catch (error) {
      console.error("Failed to update profile:", error);

      setError("Could not update your profile.");
    } finally {
      setSavingProfile(false);
    }
  }

  // ===================================================
  // POST FUNCTIONS
  // ===================================================

  async function handleCreatePost(event: SubmitEvent) {
    event.preventDefault();

    const content = postText.trim();

    if (!content) {
      return;
    }

    try {
      setPosting(true);
      setError("");

      const newPost = await createPost(
        content,
        space === "discovery" ? "public" : "inner_circle",
        space === "discovery" ? myInterestIds : [],
      );

      // Add the new post to the top of the feed
      setPosts((currentPosts) => [
        {
          ...newPost,
          username: user?.username || "you",
          display_name: user?.display_name || "You",
          avatar_url: user?.avatar_url || "",
        },
        ...currentPosts,
      ]);

      setPostText("");
    } catch (err) {
      console.error(err);

      setError("Could not publish your post.");
    } finally {
      setPosting(false);
    }
  }

  // ===================================================
  // COMMENT FUNCTIONS
  // ===================================================

  async function toggleComments(postId: string) {
    if (openComments === postId) {
      setOpenComments(null);
      return;
    }

    try {
      const data = await getComments(postId);

      setComments((current) => ({
        ...current,
        [postId]: data.comments,
      }));

      setOpenComments(postId);
    } catch (err) {
      console.error(err);

      setError("Could not load comments.");
    }
  }

  async function handleCreateComment(postId: string) {
    const content = commentText.trim();

    if (!content) {
      return;
    }

    try {
      const newComment = await createComment(postId, content);

      // Add the new comment without reloading
      setComments((current) => ({
        ...current,
        [postId]: [...(current[postId] || []), newComment],
      }));

      setCommentText("");
    } catch (err) {
      console.error(err);

      setError("Could not add comment.");
    }
  }

  // ===================================================
  // REACTION FUNCTIONS
  // ===================================================

  async function loadMyReactions(postId: string) {
    try {
      const data = await getMyReactions(postId);

      setMyReactions((current) => ({
        ...current,
        [postId]: data.reactions.map(
          (reaction: { type: ReactionType }) => reaction.type,
        ),
      }));
    } catch (err) {
      console.error(err);
    }
  }

  async function handleReaction(postId: string, type: ReactionType) {
    const current = myReactions[postId] || [];

    const alreadyReacted = current.includes(type);

    try {
      if (alreadyReacted) {
        await removeReaction(postId, type);

        setMyReactions((existing) => ({
          ...existing,
          [postId]: existing[postId].filter((reaction) => reaction !== type),
        }));
      } else {
        await addReaction(postId, type);

        setMyReactions((existing) => ({
          ...existing,
          [postId]: [...(existing[postId] || []), type],
        }));
      }
    } catch (err) {
      console.error(err);

      setError("Could not update reaction.");
    }

    setReactionMenu(null);
  }

  async function toggleReactionMenu(postId: string) {
    if (reactionMenu === postId) {
      setReactionMenu(null);
      return;
    }

    await loadMyReactions(postId);

    setReactionMenu(postId);
  }

  // ===================================================
  // INTEREST FUNCTIONS
  // ===================================================

  async function loadInterests() {
    try {
      setLoadingInterests(true);

      const [allInterests, selectedInterests] = await Promise.all([
        getInterests(),
        getMyInterests(),
      ]);

      setInterests(allInterests.interests);

      setMyInterestIds(
        selectedInterests.interests.map((interest: Interest) => interest.id),
      );
    } catch (err) {
      console.error(err);

      setError("Could not load your interests.");
    } finally {
      setLoadingInterests(false);
    }
  }

  async function toggleInterest(interestId: string) {
    const isSelected = myInterestIds.includes(interestId);

    try {
      setError("");

      if (isSelected) {
        await removeInterest(interestId);

        setMyInterestIds((currentIds) =>
          currentIds.filter((id) => id !== interestId),
        );
      } else {
        await addInterest(interestId);

        setMyInterestIds((currentIds) => [...currentIds, interestId]);
      }

      // Refresh the Discovery feed immediately
      if (space === "discovery") {
        const feed = await getDiscoveryFeed();

        setPosts(feed.posts || []);
      }
    } catch (error) {
      console.error("Failed to update interests:", error);

      setError("Could not update your interests.");
    }
  }

  // ===================================================
  // SPACE NAVIGATION
  // ===================================================

  async function switchSpace(nextSpace: "inner_circle" | "discovery") {
    if (nextSpace === space) {
      return;
    }

    try {
      setLoading(true);
      setError("");

      if (nextSpace === "discovery") {
        const feed = await getDiscoveryFeed();

        setPosts(feed.posts);
      } else {
        const feed = await getInnerCircleFeed();

        setPosts(feed.posts);
      }

      setSpace(nextSpace);
    } catch (err) {
      console.error(err);

      setError("Unable to load this space.");
    } finally {
      setLoading(false);
    }
  }

  // ===================================================
  // LOADING SCREEN
  // ===================================================

  if (loading) {
    return (
      <main className="page">
        <div className="loading">Loading your space...</div>
      </main>
    );
  }

  // ===================================================
  // PAGE UI
  // ===================================================

  return (
    <main className="page">
      {/* =============================================
          HEADER
      ============================================== */}
      <header className="topbar">
        {/* Keep your existing header JSX here */}
        <div className="brand">
          <div className="brand-mark">R</div>

          <div>
            <h1>RealSpace</h1>
            <p>Social, without the noise.</p>
          </div>
        </div>

        {user && (
          <button type="button" className="profile" onClick={openProfileEditor}>
            <div className="profile-avatar">{user.display_name.charAt(0)}</div>

            <span>{user.display_name}</span>
          </button>
        )}
      </header>

      {/* =============================================
          MAIN LAYOUT
      ============================================== */}
      <div className="layout">
        {/* ===========================================
            SIDEBAR & NAVIGATION
        ============================================ */}
        <aside className="sidebar">
          {/* Keep your existing sidebar JSX here */}
          <nav>
            <button
              className={
                space === "inner_circle" ? "nav-item active" : "nav-item"
              }
              onClick={() => switchSpace("inner_circle")}
            >
              <span>○</span>
              Inner Circle
            </button>

            <button
              className={space === "discovery" ? "nav-item active" : "nav-item"}
              onClick={() => switchSpace("discovery")}
            >
              <span>✦</span>
              Discovery
            </button>
            <button onClick={handleLogout}>Logout</button>
          </nav>

          <div className="sidebar-note">
            <strong>Your space.</strong>

            <p>
              No follower counts.
              <br />
              No trending bait.
              <br />
              No endless feed.
            </p>
          </div>
        </aside>

        {/* ===========================================
            FEED
        ============================================ */}
        <section className="feed">
          {/* Feed heading */}
          <div className="feed-heading">
            <div>
              <p className="eyebrow">
                {space === "inner_circle"
                  ? "PRIVATE SPACE"
                  : "INTENTIONAL DISCOVERY"}
              </p>

              <h2>{space === "inner_circle" ? "Inner Circle" : "Discovery"}</h2>

              <p>
                {space === "inner_circle"
                  ? "Updates from people you chose."
                  : "Public posts from interests you chose."}
              </p>
            </div>

            <div className="circle-count">
              <span>●</span>
              {posts.length} {space === "inner_circle" ? "updates" : "posts"}
            </div>
          </div>

          {/* Discovery controls */}
          {space === "discovery" && (
            <div className="discovery-controls">
              <button
                className="interest-toggle"
                onClick={() => setShowInterestPicker((current) => !current)}
              >
                ✦
                {myInterestIds.length === 0
                  ? "Choose interests"
                  : `${myInterestIds.length} interests`}
              </button>

              {myInterestIds.length > 0 && (
                <div className="selected-interests">
                  {interests
                    .filter((interest) => myInterestIds.includes(interest.id))
                    .map((interest) => (
                      <span className="interest-chip" key={interest.id}>
                        {interest.name}
                      </span>
                    ))}
                </div>
              )}
            </div>
          )}

          {/* Post composer */}
          <form
            className="composer"
            onSubmit={async (event) => {
              event.preventDefault();

              const content = postText.trim();

              if (!content) {
                return;
              }
            }}
          >
            <div className="composer-avatar">
              {user?.display_name.charAt(0)}
            </div>

            <div className="composer-content">
              <textarea
                value={postText}
                onChange={(event) => setPostText(event.target.value)}
                placeholder="What's worth sharing?"
                maxLength={1000}
                rows={3}
              />

              <div className="composer-footer">
                <span>{postText.length}/1000</span>

                <button type="submit" disabled={posting || !postText.trim()}>
                  {posting ? "Sharing..." : "Share"}
                </button>
              </div>
            </div>
          </form>

          {/* Error message */}
          {error && <div className="error">{error}</div>}

          {/* Posts */}
          <div className="posts">
            {posts.map((post) => (
              <article className="post" key={post.id}>
                <div className="post-header">
                  <div className="post-avatar">
                    {post.display_name.charAt(0)}
                  </div>

                  <div className="post-author">
                    <strong>{post.display_name}</strong>

                    <span>@{post.username}</span>
                  </div>

                  <time>{formatDate(post.created_at)}</time>
                </div>

                <p className="post-content">{post.content}</p>

                <div className="post-footer">
                  <button onClick={() => toggleComments(post.id)}>
                    {openComments === post.id ? "Hide comments" : "Comment"}
                  </button>

                  <button onClick={() => toggleReactionMenu(post.id)}>
                    React
                  </button>
                </div>

                {openComments === post.id && (
                  <div className="comments">
                    <div className="comment-list">
                      {(comments[post.id] || []).map((comment) => (
                        <div className="comment" key={comment.id}>
                          <div className="comment-avatar">
                            {comment.display_name.charAt(0)}
                          </div>

                          <div>
                            <strong>{comment.display_name}</strong>

                            <p>{comment.content}</p>
                          </div>
                        </div>
                      ))}

                      {comments[post.id]?.length === 0 && (
                        <p className="no-comments">
                          Be the first to say something.
                        </p>
                      )}
                    </div>

                    <div className="comment-composer">
                      <input
                        value={commentText}
                        onChange={(event) => setCommentText(event.target.value)}
                        placeholder="Write a thoughtful comment..."
                        onKeyDown={(event) => {
                          if (event.key === "Enter" && !event.shiftKey) {
                            event.preventDefault();

                            handleCreateComment(post.id);
                          }
                        }}
                      />

                      <button
                        onClick={() => handleCreateComment(post.id)}
                        disabled={!commentText.trim()}
                      >
                        Send
                      </button>
                    </div>
                  </div>
                )}

                {reactionMenu === post.id && (
                  <div className="reaction-menu">
                    {(
                      [
                        ["like", "Like"],
                        ["support", "Support"],
                        ["helpful", "Helpful"],
                      ] as [ReactionType, string][]
                    ).map(([type, label]) => {
                      const selected = myReactions[post.id]?.includes(type);

                      return (
                        <button
                          key={type}
                          className={
                            selected ? "reaction selected" : "reaction"
                          }
                          onClick={() => handleReaction(post.id, type)}
                        >
                          {label}
                          {selected && " ✓"}
                        </button>
                      );
                    })}
                  </div>
                )}
              </article>
            ))}
          </div>

          {posts.length > 0 && (
            <div className="caught-up">
              <div className="caught-up-line" />

              <div className="caught-up-icon">✓</div>

              <h3>You're caught up.</h3>

              <p>That's everything from your Inner Circle for now.</p>

              <div className="caught-up-line" />
            </div>
          )}

          {posts.length === 0 && (
            <div className="empty">
              <div className="empty-icon">
                {space === "inner_circle" ? "○" : "✦"}
              </div>

              <h3>
                {space === "inner_circle"
                  ? "Your circle is quiet."
                  : myInterestIds.length === 0
                    ? "Choose what you want to discover."
                    : "Nothing new right now."}
              </h3>

              <p>
                {space === "inner_circle"
                  ? "When your people share something, it'll appear here."
                  : myInterestIds.length === 0
                    ? "Discovery only shows topics you intentionally choose."
                    : "You've seen everything matching your interests for now."}
              </p>

              {space === "discovery" && myInterestIds.length === 0 && (
                <button
                  type="button"
                  className="interest-button"
                  onClick={() => setShowInterestPicker(true)}
                >
                  Manage interests
                </button>
              )}
            </div>
          )}

          {/* Empty state */}
        </section>
      </div>

      {/* =============================================
          INTEREST PICKER
      ============================================= */}

      {showInterestPicker && (
        <div
          className="modal-backdrop"
          onClick={() => setShowInterestPicker(false)}
        >
          <div
            className="interest-modal"
            onClick={(event) => event.stopPropagation()}
          >
            <div className="interest-modal-header">
              <div>
                <h2>Your interests</h2>

                <p>Choose topics you want to see in Discovery.</p>
              </div>

              <button
                type="button"
                onClick={() => setShowInterestPicker(false)}
                aria-label="Close interests"
              >
                ×
              </button>
            </div>

            {loadingInterests ? (
              <p>Loading interests...</p>
            ) : (
              <div className="interest-list">
                {interests.map((interest) => {
                  const selected = myInterestIds.includes(interest.id);

                  return (
                    <button
                      key={interest.id}
                      type="button"
                      className={
                        selected ? "interest-item selected" : "interest-item"
                      }
                      onClick={() => toggleInterest(interest.id)}
                    >
                      <span>{interest.name}</span>

                      {selected && <span>✓</span>}
                    </button>
                  );
                })}
              </div>
            )}

            <div className="interest-modal-footer">
              <button
                type="button"
                className="interest-done"
                onClick={() => setShowInterestPicker(false)}
              >
                Done
              </button>
            </div>
          </div>
        </div>
      )}
      {showProfileEditor && user && (
        <div className="interest-overlay">
          <div className="interest-modal">
            <div className="interest-modal-header">
              <div>
                <p className="eyebrow">YOUR PROFILE</p>

                <h2>Edit profile</h2>

                <p>Update how people see you on RealSpace.</p>
              </div>

              <button
                type="button"
                className="modal-close"
                onClick={() => setShowProfileEditor(false)}
              >
                ×
              </button>
            </div>

            <div className="profile-editor">
              <label>
                Display name
                <input
                  value={editDisplayName}
                  onChange={(event) => setEditDisplayName(event.target.value)}
                  maxLength={100}
                />
              </label>

              <label>
                Bio
                <textarea
                  value={editBio}
                  onChange={(event) => setEditBio(event.target.value)}
                  maxLength={500}
                  rows={4}
                />
              </label>

              <label>
                Avatar URL
                <input
                  value={editAvatarURL}
                  onChange={(event) => setEditAvatarURL(event.target.value)}
                  placeholder="https://example.com/avatar.jpg"
                />
              </label>

              <button
                type="button"
                className="interest-done"
                onClick={() => {
                  handleProfileUpdate(editDisplayName, editBio, editAvatarURL);
                }}
                disabled={savingProfile}
              >
                {savingProfile ? "Saving..." : "Save profile"}
              </button>
            </div>
          </div>
        </div>
      )}
    </main>
  );
}

// =====================================================
// UTILITY FUNCTIONS
// =====================================================

function formatDate(date: string) {
  const value = new Date(date);

  return value.toLocaleString([], {
    month: "short",
    day: "numeric",
    hour: "numeric",
    minute: "2-digit",
  });
}
