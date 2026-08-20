"use client";

import { useEffect, useState } from "react";

import {
  addReaction,
  addInterest,
  createComment,
  createPost,
  getComments,
  getInnerCircleFeed,
  getMe,
  getDiscoveryFeed,
  getInterests,
  getMyReactions,
  removeReaction,
  getMyInterests,
  removeInterest,
} from "../lib/api";

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

export default function Home() {
  const [user, setUser] = useState<User | null>(null);
  const [posts, setPosts] = useState<Post[]>([]);
  const [postText, setPostText] = useState("");
  const [loading, setLoading] = useState(true);
  const [posting, setPosting] = useState(false);
  const [error, setError] = useState("");
  const [openComments, setOpenComments] = useState<string | null>(null);

  const [comments, setComments] = useState<Record<string, Comment[]>>({});

  const [commentText, setCommentText] = useState("");

  const [reactionMenu, setReactionMenu] = useState<string | null>(null);

  const [myReactions, setMyReactions] = useState<
    Record<string, ReactionType[]>
  >({});

  const [space, setSpace] = useState<"inner_circle" | "discovery">(
    "inner_circle",
  );

  const [interests, setInterests] = useState<Interest[]>([]);

  const [myInterestIds, setMyInterestIds] = useState<string[]>([]);

  const [showInterestPicker, setShowInterestPicker] = useState(false);

  const [loadingInterests, setLoadingInterests] = useState(false);

  async function loadPage() {
    try {
      setLoading(true);
      setError("");

      const [userData, feedData] = await Promise.all([
        getMe(),
        getInnerCircleFeed(),
      ]);

      setUser(userData);
      setPosts(feedData.posts);
    } catch (err) {
      console.error(err);
      setError("Unable to load your space.");
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    loadPage();
    loadInterests();
  }, []);

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

  if (loading) {
    return (
      <main className="page">
        <div className="loading">Loading your space...</div>
      </main>
    );
  }

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

  async function toggleInterest(interestId: string) {
    const selected = myInterestIds.includes(interestId);

    try {
      if (selected) {
        await removeInterest(interestId);

        setMyInterestIds((current) =>
          current.filter((id) => id !== interestId),
        );
      } else {
        await addInterest(interestId);

        setMyInterestIds((current) => [...current, interestId]);
      }

      if (space === "discovery") {
        const feed = await getDiscoveryFeed();

        setPosts(feed.posts);
      }
    } catch (err) {
      console.error(err);

      setError("Could not update interests.");
    }
  }

  return (
    <main className="page">
      <header className="topbar">
        <div className="brand">
          <div className="brand-mark">R</div>

          <div>
            <h1>RealSpace</h1>
            <p>Social, without the noise.</p>
          </div>
        </div>

        {user && (
          <div className="profile">
            <div className="profile-avatar">{user.display_name.charAt(0)}</div>

            <span>{user.display_name}</span>
          </div>
        )}
      </header>

      <div className="layout">
        <aside className="sidebar">
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

        <section className="feed">
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
              {posts.length} updates
            </div>
          </div>

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

          {error && <div className="error">{error}</div>}

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
              <div className="empty-icon">○</div>

              <h3>Your circle is quiet.</h3>

              <p>When your people share something, it'll appear here.</p>
            </div>
          )}
        </section>
      </div>

      {showInterestPicker && (
        <div className="interest-overlay">
          <div className="interest-modal">
            <div className="interest-modal-header">
              <div>
                <p className="eyebrow">YOUR DISCOVERY</p>

                <h2>Choose your interests</h2>

                <p>You decide what belongs in your Discovery space.</p>
              </div>

              <button
                className="modal-close"
                onClick={() => setShowInterestPicker(false)}
              >
                ×
              </button>
            </div>

            <div className="interest-grid">
              {loadingInterests ? (
                <div className="loading">Loading interests...</div>
              ) : (
                interests.map((interest) => {
                  const selected = myInterestIds.includes(interest.id);

                  return (
                    <button
                      key={interest.id}
                      className={
                        selected
                          ? "interest-option selected"
                          : "interest-option"
                      }
                      onClick={() => toggleInterest(interest.id)}
                    >
                      <div>
                        <strong>{interest.name}</strong>

                        <span>{interest.description}</span>
                      </div>

                      <div className="interest-check">
                        {selected ? "✓" : "+"}
                      </div>
                    </button>
                  );
                })
              )}
            </div>

            <button
              className="interest-done"
              onClick={() => {
                setShowInterestPicker(false);

                // if (space === "discovery") {
                //   switchSpace("discovery");
                // }
              }}
            >
              Done
            </button>
          </div>
        </div>
      )}
    </main>
  );
}

function formatDate(date: string) {
  const value = new Date(date);

  return value.toLocaleString([], {
    month: "short",
    day: "numeric",
    hour: "numeric",
    minute: "2-digit",
  });
}
