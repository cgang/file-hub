# File Hub Web UI

A minimal web-based interface to browse and manage files stored via WebDAV.

## Setup

1. Install dependencies:
   ```bash
   cd web
   npm install
   ```

2. Create a `.env` file in the `web` directory with your WebDAV configuration:
   ```env
   VITE_WEBDAV_URL=http://localhost:8080
   VITE_WEBDAV_USER=your_username
   VITE_WEBDAV_PASS=your_password
   ```

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