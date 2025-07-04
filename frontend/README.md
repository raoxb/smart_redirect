# Smart Redirect Admin Dashboard

React-based admin dashboard for managing Smart Redirect service.

## Features

- **Dashboard**: Real-time statistics and traffic overview
- **Link Management**: Create, edit, and manage redirect links
- **Target Configuration**: Configure multiple targets with weighted distribution
- **Statistics**: Detailed analytics and performance metrics
- **Batch Operations**: Import/export links via CSV
- **User Management**: Admin panel for user and permission management

## Tech Stack

- **React 18** with TypeScript
- **Vite** for fast development and building
- **Ant Design 5** for UI components
- **React Query** for server state management
- **Zustand** for client state management
- **React Router 6** for navigation
- **Recharts** for data visualization

## Development

### Prerequisites

- Node.js 18+
- npm or yarn

### Installation

```bash
# Install dependencies
npm install

# Start development server
npm run dev

# Build for production
npm run build

# Type checking
npm run type-check

# Linting
npm run lint
```

### Environment Variables

Create a `.env` file in the frontend directory:

```env
VITE_API_URL=http://localhost:8080
```

## Project Structure

```
src/
├── pages/          # Page components
├── components/     # Reusable UI components
├── services/       # API client and services
├── hooks/          # Custom React hooks
├── store/          # Zustand stores
├── types/          # TypeScript definitions
└── utils/          # Utility functions
```

## Available Scripts

- `npm run dev` - Start development server on port 3000
- `npm run build` - Build for production
- `npm run preview` - Preview production build
- `npm run lint` - Run ESLint
- `npm run lint:fix` - Fix ESLint errors
- `npm run type-check` - Run TypeScript type checking

## Key Components

### Pages

- **Login** - Authentication page
- **Dashboard** - Overview with statistics and charts
- **Links** - Link management with CRUD operations
- **LinkDetail** - Detailed view of a single link
- **Statistics** - Advanced analytics (coming soon)
- **Templates** - Link template management (coming soon)
- **Users** - User management (admin only, coming soon)

### State Management

- **authStore** - Authentication state and user info
- **React Query** - Server state caching and synchronization

### API Integration

All API calls are centralized in `src/services/api.ts` with proper TypeScript typing.

## Building for Production

```bash
# Build the application
npm run build

# The output will be in the dist/ directory
# Serve it with any static file server

# Example with serve:
npx serve -s dist
```

## Deployment

The frontend can be deployed to any static hosting service:

- **Nginx** - See nginx configuration in parent project
- **Vercel** - `vercel --prod`
- **Netlify** - `netlify deploy --prod`
- **S3 + CloudFront** - Upload dist/ to S3

## License

This project is proprietary software.