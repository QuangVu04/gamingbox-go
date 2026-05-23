# Gaming Box API Documentation

Tài liệu chi tiết về tất cả các API đã được triển khai trong hệ thống **Gaming Box Backend (Go/Gin)**. Hệ thống API chạy trên base path `/api/v1` (mặc định tại `http://localhost:8080/api/v1`).

---

## Danh Mục Nhóm API
1. [Authentication & OAuth (Xác thực)](#1-authentication--oauth-xác-thực)
2. [Users & Profile (Thành viên & Trang cá nhân)](#2-users--profile-thành-viên--trang-cá-nhân)
3. [Games (Trò chơi & Tương tác)](#3-games-trò-chơi--tương-tác)
4. [Reviews (Đánh giá)](#4-reviews-đánh-giá)
5. [Lists (Danh sách Game cá nhân)](#5-lists-danh-sách-game-cá-nhân)
6. [Likes (Yêu thích chung)](#6-likes-yêu-thích-chung)
7. [Notifications (Thông báo)](#7-notifications-thông-báo)
8. [Admin (Quản trị viên)](#8-admin-quản-trị-viên)

---

## 1. Authentication & OAuth (Xác thực)

### POST `/auth/register`
* **Mô tả**: Đăng ký tài khoản thành viên mới bằng Email.
* **Mã trạng thái**: `201 Created`
* **Request Body**:
```json
{
  "email": "gamer@example.com",
  "username": "GamerX",
  "password": "strongPassword123"
}
```
* **Response dự kiến (Success)**:
```json
{
  "status": "success",
  "data": {
    "user": {
      "id": 1,
      "email": "gamer@example.com",
      "username": "GamerX",
      "avatar_url": null,
      "role": "user"
    },
    "access_token": "eyJhbGciOi...",
    "refresh_token": "def456...",
    "expires_in": 86400
  }
}
```

### POST `/auth/login`
* **Mô tả**: Đăng nhập bằng Email và Password.
* **Mã trạng thái**: `200 OK`
* **Request Body**:
```json
{
  "email": "gamer@example.com",
  "password": "strongPassword123",
  "remember_me": true
}
```
* **Response dự kiến (Success)**:
```json
{
  "status": "success",
  "data": {
    "user": {
      "id": 1,
      "email": "gamer@example.com",
      "username": "GamerX",
      "avatar_url": "https://avatar.url/1.png",
      "role": "user"
    },
    "access_token": "eyJhbGciOi...",
    "refresh_token": "def456...",
    "expires_in": 86400
  }
}
```

### POST `/auth/refresh`
* **Mô tả**: Làm mới Token truy cập hết hạn sử dụng Refresh Token.
* **Request Body**:
```json
{
  "refresh_token": "def456..."
}
```
* **Response dự kiến**:
```json
{
  "status": "success",
  "data": {
    "access_token": "newAccesseyJ...",
    "refresh_token": "newRefreshdef...",
    "expires_in": 86400
  }
}
```

### POST `/auth/forgot-password`
* **Mô tả**: Yêu cầu gửi mã OTP khôi phục mật khẩu vào Email.
* **Request Body**:
```json
{
  "email": "gamer@example.com"
}
```
* **Response dự kiến**:
```json
{
  "status": "success",
  "message": "Mã xác thực đã được gửi tới email của bạn"
}
```

### POST `/auth/verify-code`
* **Mô tả**: Xác thực mã OTP 6 số gửi từ email.
* **Request Body**:
```json
{
  "email": "gamer@example.com",
  "code": "123456"
}
```
* **Response dự kiến**:
```json
{
  "status": "success",
  "data": {
    "reset_token": "resetTokenSecret123"
  }
}
```

### POST `/auth/reset-password`
* **Mô tả**: Đặt lại mật khẩu mới.
* **Request Body**:
```json
{
  "reset_token": "resetTokenSecret123",
  "new_password": "NewStrongPassword123"
}
```
* **Response dự kiến**:
```json
{
  "status": "success",
  "message": "Đặt lại mật khẩu thành công"
}
```

### GET `/auth/steam`
* **Mô tả**: Đi đến trang xác thực Steam OpenID (OAuth).
* **Mã trạng thái**: `302 Redirect`

### GET `/auth/steam/callback`
* **Mô tả**: Xử lý phản hồi từ Steam OAuth.

---

## 2. Users & Profile (Thành viên & Trang cá nhân)

### GET `/users/profile?id={userId}`
* **Mô tả**: Lấy thông tin chi tiết trang cá nhân của thành viên.
* **Query Params**: `id` (ID của người dùng)
* **Response dự kiến**:
```json
{
  "status": "success",
  "data": {
    "id": 1,
    "email": "gamer@example.com",
    "username": "GamerX",
    "avatar_url": "https://avatar.url/1.png",
    "bio": "Đam mê game RPG",
    "role": "user",
    "status": "active",
    "follower_count": 12,
    "following_count": 8,
    "review_count": 4,
    "list_count": 2,
    "game_logs_count": 42,
    "average_rating": 4.5,
    "created_at": "2026-01-15T08:30:00Z"
  }
}
```

### GET `/users/me` (Yêu cầu Token)
* **Mô tả**: Lấy thông tin cá nhân của User đang đăng nhập.
* **Response dự kiến**: Giống như cấu trúc `/users/profile`.

### POST `/users/follow` (Yêu cầu Token)
* **Mô tả**: Follow hoặc Unfollow một người dùng khác.
* **Request Body**:
```json
{
  "userId": 2
}
```
* **Response dự kiến**:
```json
{
  "status": "success",
  "data": {
    "is_following": true
  }
}
```

### PUT `/users/favorite-games` (Yêu cầu Token)
* **Mô tả**: Cập nhật danh sách trò chơi yêu thích (Tối đa 4 game hiển thị trên đầu profile).
* **Request Body**:
```json
{
  "game_ids": [12, 45, 89]
}
```
* **Response dự kiến**:
```json
{
  "status": "success",
  "message": "Cập nhật danh sách game yêu thích thành công"
}
```

### GET `/users`
* **Mô tả**: Tìm kiếm thành viên công khai có phân trang và sắp xếp.
* **Query Params**:
  - `search` (từ khóa tìm kiếm theo tên hoặc bio, mặc định trống)
  - `page` (trang hiện tại, mặc định 1)
  - `limit` (số lượng trên mỗi trang, mặc định 10)
  - `sort` (sắp xếp: `active` - hoạt động tích cực nhất, `followers` - nhiều người theo dõi nhất, `newest` - tài khoản mới tạo)
* **Response dự kiến**:
```json
{
  "status": "success",
  "pagination": {
    "total_records": 1,
    "current_page": 1,
    "total_pages": 1,
    "limit": 10
  },
  "data": [
    {
      "id": 1,
      "email": "gamer@example.com",
      "username": "GamerX",
      "avatar_url": "https://avatar.url/1.png",
      "bio": "Đam mê game RPG",
      "role": "user",
      "follower_count": 12,
      "following_count": 8,
      "review_count": 4,
      "list_count": 2,
      "game_logs_count": 42,
      "average_rating_count": 4.5,
      "created_at": "2026-01-15T08:30:00Z",
      "recent_games": [
        {
          "id": 1,
          "steam_id": 1245620,
          "title": "Elden Ring",
          "poster": "https://cover.url/eldenring.jpg",
          "release_date": "2022-02-25T00:00:00Z",
          "price": 59.99,
          "is_free": false,
          "avg_rating": 4.8,
          "review_count": 425,
          "like_count": 890
        }
      ]
    }
  ]
}
```

---

## 3. Games (Trò chơi & Tương tác)

### GET `/games/trending`
* **Mô tả**: Danh sách Game thịnh hành phân trang.
* **Query Params**: `page` (mặc định 1), `limit` (mặc định 12)
* **Response dự kiến**:
```json
{
  "status": "success",
  "pagination": {
    "total_records": 120,
    "current_page": 1,
    "total_pages": 10,
    "limit": 12
  },
  "data": [
    {
      "game_id": 1,
      "title": "Elden Ring",
      "thumbnail": "https://cover.url/eldenring.jpg",
      "trending_score": 1500,
      "avg_rating": 4.8,
      "total_reviews": 425,
      "like_count": 890,
      "studios": ["FromSoftware"],
      "genres": ["Action", "RPG"]
    }
  ]
}
```

### GET `/games/popular`
* **Mô tả**: Lấy danh sách Game phổ biến dựa trên đánh giá cao nhất.

### GET `/games/search?q={query}`
* **Mô tả**: Tìm kiếm game theo tên trong database.
* **Query Params**: `q` (từ khóa tìm kiếm), `page`, `limit`

### GET `/games/:id`
* **Mô tả**: Lấy thông tin chi tiết một Game bằng ID.
* **Response dự kiến**:
```json
{
  "status": "success",
  "data": {
    "id": 1,
    "steam_id": 1245620,
    "title": "Elden Ring",
    "description": "Hành trình thế giới mở rộng lớn...",
    "release_date": "2022-02-25T00:00:00Z",
    "price": 59.99,
    "is_free": false,
    "avg_rating": 4.8,
    "review_count": 425,
    "like_count": 890,
    "plays_count": 1200,
    "playing_count": 120,
    "dropped_count": 15,
    "backlog_count": 350,
    "wishlist_count": 680,
    "studio": {
      "id": 5,
      "name": "FromSoftware"
    },
    "genres": [
      {"id": 1, "name": "Action"},
      {"id": 2, "name": "RPG"}
    ],
    "platforms": [
      {"id": 1, "name": "PC"},
      {"id": 2, "name": "PlayStation 5"}
    ],
    "images": [
      "https://cover.url/eldenring_cover.jpg",
      "https://cover.url/eldenring_header.jpg"
    ]
  }
}
```

### POST `/games/log` (Yêu cầu Token)
* **Mô tả**: Đánh dấu tình trạng chơi game (playing, completed, dropped, backlog).
* **Request Body**:
```json
{
  "game_id": 1,
  "status": "completed"
}
```
* **Response dự kiến**:
```json
{
  "status": "success",
  "message": "Đã lưu trạng thái chơi game"
}
```

### POST `/games/rate` (Yêu cầu Token)
* **Mô tả**: Đánh giá số điểm (từ 0.5 đến 5.0) cho Game.
* **Request Body**:
```json
{
  "game_id": 1,
  "rating": 4.5
}
```

---

## 4. Reviews (Đánh giá)

### POST `/reviews` (Yêu cầu Token)
* **Mô tả**: Tạo đánh giá (review) cho một trò chơi.
* **Request Body**:
```json
{
  "game_id": 1,
  "content": "Một kiệt tác tuyệt vời của thế giới mở...",
  "recommend": "recommend",
  "is_spoiler": false
}
```
* **Response dự kiến**:
```json
{
  "status": "success",
  "data": {
    "id": 12,
    "game_id": 1,
    "user_id": 1,
    "content": "Một kiệt tác tuyệt vời của thế giới mở...",
    "recommend": "recommend",
    "is_spoiler": false,
    "created_at": "2026-05-23T10:00:00Z"
  }
}
```

### PUT `/reviews/:id` (Yêu cầu Token)
* **Mô tả**: Chỉnh sửa bài đánh giá (Chỉ chủ sở hữu).

### GET `/reviews/trending`
* **Mô tả**: Lấy các bài viết đánh giá thịnh hành nhất trong tuần.

---

## 5. Lists (Danh sách Game cá nhân)

### GET `/lists/trending`
* **Mô tả**: Lấy danh sách List Game thịnh hành phân trang.
* **Response dự kiến**:
```json
{
  "status": "success",
  "pagination": {
    "total_records": 4,
    "current_page": 1,
    "total_pages": 1,
    "limit": 10
  },
  "data": [
    {
      "list_id": 1,
      "title": "Đồ họa đỉnh #",
      "author": {
        "id": 52,
        "username": "qyArHZg",
        "avatar": "https://avatar.url/default.png"
      },
      "game_count": 5,
      "thumbnails": [
        "https://cdn.steam.jpg",
        "https://cdn.steam2.jpg"
      ],
      "weekly_likes_count": 12,
      "total_likes": 25,
      "user_has_liked": false,
      "comment_count": 1,
      "created_at": "2026-05-20T10:00:00Z"
    }
  ]
}
```

### POST `/lists` (Yêu cầu Token)
* **Mô tả**: Tạo danh sách game cá nhân mới.
* **Request Body**:
```json
{
  "title": "Top Game Nhập Vai Hay Nhất",
  "description": "Các game thế giới mở cốt truyện sâu sắc...",
  "is_public": true,
  "entries": [
    { "game_id": 1, "note": "Hơn 200 tiếng chơi không chán!" },
    { "game_id": 8, "note": "Không khí rùng rợn siêu đỉnh." }
  ]
}
```
* **Response dự kiến**:
```json
{
  "status": "success",
  "data": {
    "id": 15,
    "title": "Top Game Nhập Vai Hay Nhất",
    "description": "Các game thế giới mở cốt truyện sâu sắc...",
    "game_count": 2,
    "like_count": 0,
    "comment_count": 0,
    "games": [
      { "game_id": 1, "title": "Elden Ring", "note": "Hơn 200 tiếng chơi không chán!" }
    ]
  }
}
```

### GET `/lists/:id`
* **Mô tả**: Chi tiết một danh sách game cụ thể.
* **Response dự kiến**:
```json
{
  "status": "success",
  "data": {
    "id": 1,
    "title": "Đồ họa đỉnh #",
    "description": "Danh sách này tổng hợp các game hay nhất...",
    "author": {
      "id": 52,
      "username": "qyArHZg",
      "avatar": "https://avatar.url/default.png"
    },
    "game_count": 5,
    "like_count": 25,
    "user_has_liked": false,
    "comment_count": 1,
    "games": [
      {
        "game_id": 2,
        "title": "Counter-Strike 2",
        "poster": "https://cdn.steam.jpg",
        "note": "Game cực hay"
      }
    ]
  }
}
```

### POST `/lists/:id/comments` (Yêu cầu Token)
* **Mô tả**: Đăng bình luận vào một danh sách game.
* **Request Body**:
```json
{
  "content": "Danh sách này chất lượng quá bạn ơi!",
  "parent_id": null
}
```
* **Response dự kiến**:
```json
{
  "status": "success",
  "data": {
    "id": 142,
    "user": {
      "username": "GamerX",
      "avatar": "https://avatar.url/1.png"
    },
    "content": "Danh sách này chất lượng quá bạn ơi!",
    "created_at": "2026-05-23T15:20:00Z"
  }
}
```

---

## 6. Likes (Yêu thích chung)

### POST `/games/:id/like` (Yêu cầu Token)
* **Mô tả**: Thích / Bỏ thích một Trò chơi.
* **Response dự kiến**:
```json
{
  "status": "success",
  "message": "Đã thực hiện thích/bỏ thích thành công"
}
```

### POST `/reviews/:id/like` (Yêu cầu Token)
* **Mô tả**: Thích / Bỏ thích một bài Đánh giá.

### POST `/lists/:id/like` (Yêu cầu Token)
* **Mô tả**: Thích / Bỏ thích một Danh sách game.

### POST `/comments/:id/like` (Yêu cầu Token)
* **Mô tả**: Thích / Bỏ thích một Bình luận.

---

## 7. Notifications (Thông báo)

### GET `/notifications` (Yêu cầu Token)
* **Mô tả**: Lấy danh sách thông báo hoạt động của người dùng (Có người like, follow, bình luận...).
* **Response dự kiến**:
```json
{
  "status": "success",
  "data": [
    {
      "id": 876,
      "sender": {
        "username": "AlexG",
        "avatar": "https://avatar.url/2.png"
      },
      "action_type": "like",
      "target_type": "list",
      "target_id": 1,
      "is_read": false,
      "created_at": "2026-05-23T14:00:00Z"
    }
  ]
}
```

---

## 8. Admin (Quản trị viên)

### GET `/admin/stats` (Yêu cầu Token Admin)
* **Mô tả**: Lấy thống kê tổng quan của hệ thống hiển thị trên Dashboard.
* **Response dự kiến**:
```json
{
  "status": "success",
  "data": {
    "total_users": { "value": "154", "change": "+5", "is_positive": true },
    "total_games": { "value": "452", "change": "+12", "is_positive": true },
    "total_reviews": { "value": "240", "change": "+8", "is_positive": true }
  }
}
```

### DELETE `/admin/users/:id` (Yêu cầu Token Admin)
* **Mô tả**: Xóa hoặc vô hiệu hóa tài khoản của một thành viên.
