# WOOHWAHAE CMS Backend

Simple Flask-based content management system for WOOHWAHAE website.

## Setup

1. Install dependencies:
```bash
cd website/backend
pip3 install -r requirements.txt
```

2. Configure environment:
```bash
cp .env.example .env
# Edit .env with your settings
```

3. Run the server:
```bash
python3 app.py
```

Server will run on `http://localhost:5000`

## API Endpoints

### Authentication
- `POST /api/auth/login` - Login with password
- `POST /api/auth/logout` - Logout
- `GET /api/auth/check` - Check auth status

### Archive Management
- `GET /api/archive` - Get all posts
- `POST /api/archive` - Create new post (auth required)
- `PUT /api/archive/<slug>` - Update post (auth required)
- `DELETE /api/archive/<slug>` - Delete post (auth required)

### Media Upload
- `POST /api/upload` - Upload image (auth required)

## Example: Create New Post

```bash
curl -X POST http://localhost:5000/api/archive \
  -H "Content-Type: application/json" \
  -d '{
    "slug": "issue-004-minimal-life",
    "title": "미니멀 라이프",
    "date": "2026.02.18",
    "issue": "Issue 004",
    "preview": "덜어냄으로 채우는 삶.",
    "category": "Essay",
    "content": "<p>미니멀 라이프에 대한 생각...</p>"
  }'
```

## Security Notes

- Change `ADMIN_PASSWORD` in production
- Use HTTPS in production
- Consider adding rate limiting
- Consider JWT tokens for better auth
