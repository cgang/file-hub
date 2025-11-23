# File Hub Web UI Design

This document outlines the design and implementation of the web-based UI layer of the File Hub project. The UI is embedded directly in the binary and provides a browser-based interface to interact with files stored via the WebDAV interface.

## Goals

- Provide a user-friendly web interface to browse, upload, and manage files
- Keep dependencies minimal and bundle size small
- Ensure responsive design that works on all device types
- Integrate seamlessly with existing WebDAV endpoints
- Package the UI directly in the binary for simple deployment
- Maintain security through proper authentication

## Technology Stack

### Frontend Framework
- **Svelte**: A modern frontend framework that compiles to highly efficient vanilla JavaScript
- **Vite**: Next-generation frontend build tool with instant server start and lightning-fast HMR

### Dependencies (Minimal)
- `svelte`: Core UI library with no runtime
- `vite`: For build tooling and dev server

### No Additional HTTP Libraries
- Using native browser fetch API for all WebDAV operations
- Direct integration with WebDAV HTTP methods (PROPFIND, GET, PUT, DELETE, MKCOL)

### Reactivity
- Using Svelte's built-in reactive declarations instead of hooks
- Direct state manipulation with Svelte's reactive assignments

## Project Structure

```
web/
├── index.html
├── package.json
├── vite.config.js
├── src/
│   ├── main.js
│   ├── components/
│   │   ├── FileBrowser.svelte
│   │   ├── FileCard.svelte
│   │   ├── NavigationBar.svelte
│   │   └── UploadComponent.svelte
│   ├── styles/
│   │   └── main.css
│   └── utils/
│       └── webdav.js
└── public/
    └── favicon.ico
```

## Components

### FileBrowser Component
The main container that displays files and directories in a grid or list view. Handles the primary file browsing experience with options to switch between different view modes.

### FileCard Component
Represents an individual file or directory with relevant metadata such as name, size, type, and date modified. Provides visual cues for different file types.

### NavigationBar Component
Shows the current path with breadcrumbs to help users navigate the file structure. Allows quick navigation to parent directories.

### UploadComponent
Provides drag-and-drop file upload functionality with visual feedback and progress indicators. Handles single and multiple file uploads.

### Sidebar Component (Optional)
Displays user information and storage quota usage. Could include quick navigation links or file operation controls.

## WebDAV Integration

### API Integration
- **Authentication**: Using HTTP Basic Authentication with stored credentials
- **Directory listing**: Using PROPFIND method to retrieve file/directory metadata
- **File operations**: Using GET/PUT/DELETE to download/upload/delete files
- **Navigation**: Using path-based URLs to navigate directories

### Implementation in `web/utils/webdav.js`
- `listDirectory(path)`: Get files/folders in a directory using PROPFIND
- `uploadFile(path, file)`: Upload a file using PUT
- `downloadFile(path)`: Download a file using GET
- `deleteFile(path)`: Delete a file using DELETE
- `createDirectory(path)`: Create a directory using MKCOL

## Key Features

### File Browsing
- View files and directories in a clean, intuitive interface
- Toggle between grid and list views
- Breadcrumb navigation for easy path tracking

### File Operations
- Upload files via drag-and-drop or traditional file picker
- Download files with a single click
- Delete files with confirmation
- Preview common file types (images, text files)

### Responsive Design
- Works on mobile, tablet, and desktop devices
- Adapts layout based on screen size
- Touch-friendly controls for mobile devices

### User Information
- Display storage quota usage
- Show user information (if applicable)

## Security Considerations

- Use HTTPS in production to protect credentials and data
- Implement proper authentication via HTTP Basic Auth
- Sanitize filenames and paths to prevent directory traversal
- Validate file types on upload to prevent malicious content

## Development Workflow

### Setup
1. Navigate to the `web/` directory
2. Install dependencies: `npm install`
3. Start development server: `npm run dev`

### Build Process for Integration
- The `make build` command automatically:
  - Builds the frontend assets using Vite
  - Embeds the assets in the Go binary
  - Creates a single executable serving both UI and API
- Bundle minimization and optimization happens automatically
- Asset fingerprinting for cache busting happens automatically

### Deployment
- Build the integrated binary with `make build` (this automatically builds frontend assets and embeds them)
- The single binary serves both the WebDAV API and the web interface
- Access the UI at the root path (e.g., `http://localhost:8080`)
- Access WebDAV API at `/webdav` path (e.g., `http://localhost:8080/webdav`)

## Future Enhancements

- File sharing capabilities
- Search functionality
- File versioning
- Preview for more file types (PDF, documents, etc.)
- Dark/light theme options
- Keyboard shortcuts for power users

## Performance Considerations

- Leverage Svelte's compile-time optimizations for minimal bundle size
- Implement virtual scrolling for large directories
- Use Svelte's reactive declarations for efficient state updates
- Cache directory listings where appropriate
- Load images lazily as they come into view

## Testing Strategy

- Unit tests for utility functions
- Component tests for UI functionality
- Integration tests for WebDAV API calls
- End-to-end tests for critical user flows