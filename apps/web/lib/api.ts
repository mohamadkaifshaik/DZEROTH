const API_URL = "http://localhost:8080";

const USER_ID = "bf5520a6-0bfe-441a-8699-34515c4e51cc";

async function request(path: string, options: RequestInit = {}) {
  const response = await fetch(`${API_URL}${path}`, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      "X-User-ID": USER_ID,
      ...options.headers,
    },
  });

  if (!response.ok) {
    const message = await response.text();

    throw new Error(message || `Request failed: ${response.status}`);
  }

  if (response.status === 204) {
    return null;
  }

  return response.json();
}

// --------------------
// User
// --------------------

export async function getMe() {
  return request("/api/v1/me");
}

// --------------------
// Inner Circle
// --------------------

export async function getInnerCircle() {
  return request("/api/v1/inner-circle");
}

export async function getInnerCircleFeed() {
  return request("/api/v1/feeds/inner-circle");
}

// --------------------
// Posts
// --------------------

export async function createPost(
  content: string,
  visibility: "inner_circle" | "public" = "inner_circle",
  interestIds: string[] = [],
) {
  return request("/api/v1/posts", {
    method: "POST",
    body: JSON.stringify({
      content,
      visibility,
      interest_ids: interestIds,
    }),
  });
}

// --------------------
// Comments
// --------------------

export async function getComments(postId: string) {
  return request(`/api/v1/posts/${postId}/comments`);
}

export async function createComment(postId: string, content: string) {
  return request(`/api/v1/posts/${postId}/comments`, {
    method: "POST",
    body: JSON.stringify({
      content,
    }),
  });
}

// --------------------
// Reactions
// --------------------

export async function getMyReactions(postId: string) {
  return request(`/api/v1/posts/${postId}/reactions/me`);
}

export async function addReaction(postId: string, type: string) {
  return request(`/api/v1/posts/${postId}/reactions`, {
    method: "POST",
    body: JSON.stringify({
      type,
    }),
  });
}

export async function removeReaction(postId: string, type: string) {
  return request(`/api/v1/posts/${postId}/reactions/${type}`, {
    method: "DELETE",
  });
}

export async function getInterests() {
  return request("/api/v1/interests");
}

export async function getMyInterests() {
  return request("/api/v1/me/interests");
}

export async function addInterest(interestId: string) {
  return request("/api/v1/me/interests", {
    method: "POST",
    body: JSON.stringify({
      interest_id: interestId,
    }),
  });
}

export async function removeInterest(interestId: string) {
  return request(`/api/v1/me/interests/${interestId}`, {
    method: "DELETE",
  });
}

export async function getDiscoveryFeed() {
  return request("/api/v1/feeds/discovery");
}
