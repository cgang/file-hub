// WebDAV utility functions for File Hub web UI

// Configuration - these would typically come from environment or user settings
const WEBDAV_URL = import.meta.env.VITE_WEBDAV_URL || '/webdav';
const WEBDAV_USER = import.meta.env.VITE_WEBDAV_USER || 'user';
const WEBDAV_PASS = import.meta.env.VITE_WEBDAV_PASS || 'password';

// Create authorization header
function getAuthHeader() {
  const credentials = btoa(`${WEBDAV_USER}:${WEBDAV_PASS}`);
  return {
    'Authorization': `Basic ${credentials}`,
    'Content-Type': 'application/octet-stream' // Default for file uploads
  };
}

// Function to list directory contents using PROPFIND
export async function listDirectory(path) {
  // Ensure path starts with /
  if (!path.startsWith('/')) {
    path = '/' + path;
  }
  
  // Ensure path ends with /
  if (!path.endsWith('/')) {
    path = path + '/';
  }

  const response = await fetch(`${WEBDAV_URL}${path}`, {
    method: 'PROPFIND',
    headers: {
      ...getAuthHeader(),
      'Depth': '1'  // Only get immediate children
    }
  });

  if (!response.ok) {
    throw new Error(`Failed to list directory: ${response.status} ${response.statusText}`);
  }

  const responseText = await response.text();
  
  // Parse the XML response from WebDAV
  const parser = new DOMParser();
  const xmlDoc = parser.parseFromString(responseText, 'text/xml');
  
  // Extract file and directory information
  const responses = xmlDoc.getElementsByTagName('response');
  const items = [];
  
  for (let i = 0; i < responses.length; i++) {
    const response = responses[i];
    const href = response.getElementsByTagName('href')[0]?.textContent;
    
    // Skip the entry for the current directory itself
    if (!href || decodeURIComponent(href).endsWith(path) && !decodeURIComponent(href).includes('/', href.length - path.length)) {
      continue;
    }
    
    const propstat = response.getElementsByTagName('propstat')[0];
    if (!propstat) continue;
    
    const prop = propstat.getElementsByTagName('prop')[0];
    if (!prop) continue;
    
    const displayName = prop.getElementsByTagName('displayname')[0]?.textContent || '';
    const contentType = prop.getElementsByTagName('getcontenttype')[0]?.textContent;
    const contentLength = prop.getElementsByTagName('getcontentlength')[0]?.textContent;
    const lastModified = prop.getElementsByTagName('getlastmodified')[0]?.textContent;
    const resourceType = prop.getElementsByTagName('resourcetype')[0];
    const isCollection = resourceType?.getElementsByTagName('collection').length > 0;
    
    // Extract the actual name from the href
    const pathParts = decodeURIComponent(href).split('/').filter(part => part !== '');
    const currentPathParts = path.split('/').filter(part => part !== '');
    
    let name;
    if (pathParts.length > currentPathParts.length) {
      name = pathParts[currentPathParts.length];
    } else {
      continue; // Skip if it's not a direct child
    }
    
    // Skip the parent directory indicator if present
    if (name === '..') continue;

    items.push({
      name: displayName || name,
      path: decodeURIComponent(href),
      type: isCollection ? 'directory' : 'file',
      size: contentLength ? parseInt(contentLength) : undefined,
      lastModified: lastModified,
      contentType: contentType
    });
  }
  
  return items;
}

// Function to upload a file
export async function uploadFile(path, file) {
  // Ensure path starts with /
  if (!path.startsWith('/')) {
    path = '/' + path;
  }
  
  // If path doesn't end with /, it's a file path, so get the directory
  if (!path.endsWith('/')) {
    const pathParts = path.split('/');
    pathParts.pop(); // Remove the last part (filename)
    path = pathParts.join('/') + '/';
  }
  
  // Create the full file path
  const fullPath = path + file.name;
  
  const response = await fetch(`${WEBDAV_URL}${fullPath}`, {
    method: 'PUT',
    headers: getAuthHeader(),
    body: file
  });

  if (!response.ok) {
    throw new Error(`Failed to upload file: ${response.status} ${response.statusText}`);
  }
  
  return response;
}

// Function to download a file
export async function downloadFile(path) {
  // Ensure path starts with /
  if (!path.startsWith('/')) {
    path = '/' + path;
  }
  
  const response = await fetch(`${WEBDAV_URL}${path}`, {
    method: 'GET',
    headers: getAuthHeader()
  });

  if (!response.ok) {
    throw new Error(`Failed to download file: ${response.status} ${response.statusText}`);
  }
  
  return response.blob();
}

// Function to delete a file or directory
export async function deleteFile(path) {
  // Ensure path starts with /
  if (!path.startsWith('/')) {
    path = '/' + path;
  }
  
  const response = await fetch(`${WEBDAV_URL}${path}`, {
    method: 'DELETE',
    headers: getAuthHeader()
  });

  if (!response.ok) {
    throw new Error(`Failed to delete file: ${response.status} ${response.statusText}`);
  }
  
  return response;
}

// Function to create a directory
export async function createDirectory(path, dirName) {
  // Ensure base path starts with /
  if (!path.startsWith('/')) {
    path = '/' + path;
  }
  
  // Ensure path ends with /
  if (!path.endsWith('/')) {
    path = path + '/';
  }
  
  // Create the full directory path
  const fullPath = path + dirName;
  
  const response = await fetch(`${WEBDAV_URL}${fullPath}`, {
    method: 'MKCOL',
    headers: getAuthHeader()
  });

  if (!response.ok) {
    throw new Error(`Failed to create directory: ${response.status} ${response.statusText}`);
  }
  
  return response;
}