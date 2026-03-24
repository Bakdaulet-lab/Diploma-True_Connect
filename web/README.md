# TrueConnect Web

Trust-based social networking and dating platform — Web Frontend.

## Tech Stack

- **Framework:** Next.js 14 (App Router)
- **Language:** TypeScript
- **Styling:** Tailwind CSS
- **Server State:** TanStack React Query
- **Client State:** Zustand
- **HTTP Client:** Axios

## Getting Started

```bash
# Install dependencies
npm install

# Start development server
npm run dev

# Build for production
npm run build
npm start
```

The app runs at `http://localhost:3000` and connects to the API at `http://localhost:8080`.

## Project Structure

```
src/
├── app/             # Next.js App Router pages
│   ├── (auth)/      # Login & register
│   └── (main)/      # Authenticated pages
├── components/      # React components
├── hooks/           # Custom React hooks
├── lib/             # API client, WebSocket, utilities
├── store/           # Zustand state stores
├── types/           # TypeScript interfaces
└── styles/          # Global CSS
```
