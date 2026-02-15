# 97layerOS PWA Frontend

Next.js Progressive Web App for real-time agent intelligence visualization.

## Installation

```bash
cd frontend
npm install
```

## Development

```bash
npm run dev
```

Open [http://localhost:3000](http://localhost:3000) in your browser.

## Building for Production

```bash
npm run build
npm start
```

## Tech Stack

- **Next.js 14** - App Router with TypeScript
- **Tailwind CSS** - Utility-first CSS with 우화해 brand colors
- **WebSocket** - Real-time communication with FastAPI backend
- **PWA** - Progressive Web App capabilities (Phase 1 complete, manifest in Phase 2)

## Components

### Phase 1: Health Monitor

- `HealthMonitor.tsx` - Hybrid system status display
  - MacBook vs GCP VM status
  - Real-time WebSocket updates
  - Node health indicators
  - Last sync/heartbeat timestamps

### Phase 2 (Upcoming)

- Agent orchestration chat
- Thought process streaming
- Agent status indicators

### Phase 3 (Upcoming)

- Asset gallery
- File browser
- AI tagging system

## Brand Design System

**Colors** (Tailwind config):
- `brand-black`: #0A0A0A
- `brand-white`: #FAFAFA
- `brand-gold`: #D4AF37 (accent)
- `brand-gray-*`: 50-900 scale

**Aesthetic**: Minimal, high-end, 우화해 identity

## Environment Variables

Create `.env.local`:

```env
NEXT_PUBLIC_WS_URL=ws://localhost:8080/ws
NEXT_PUBLIC_API_URL=http://localhost:8080
```

## Testing

1. Start FastAPI backend: `cd ../execution/api && python main.py`
2. Start Next.js dev: `npm run dev`
3. Open http://localhost:3000
4. Verify WebSocket connection (green dot)
5. Modify `knowledge/system/sync_state.json` and watch real-time updates

## Deployment

### Development (Local Podman)
```bash
npm run dev
# Access via http://localhost:3000
```

### Production (GCP Cloud Run - Phase 4)
```bash
npm run build
# Docker + Cloud Run deployment config TBD
```

## Phase Roadmap

- ✅ **Phase 1**: Health Monitor + WebSocket
- 🔜 **Phase 2**: Agent Chat + Thought Streaming
- 🔜 **Phase 3**: Asset Gallery + File Browser
- 🔜 **Phase 4**: PWA Manifest + Cloud Deployment
