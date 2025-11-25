# File Hub Web UI

A minimal web-based interface to browse and manage files stored via WebDAV.

## Setup

1. Install dependencies:
   ```bash
   cd web
   npm install
   ```

2. The WebDAV URL is hardcoded to `/dav` and authentication is handled server-side, so no additional configuration is needed in the frontend.

3. Start the development server:
   ```bash
   npm run dev
   ```

## Features

- Browse files and directories
- Navigate through folders
- Upload new files
- View file details (size, last modified date)

## Build for Production

To build the static assets:

```bash
npm run build
```

The built files will be in the `dist` directory (relative to the main project root).