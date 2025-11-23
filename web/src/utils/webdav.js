// WebDAV utility functions for File Hub web UI

export let davBasePath = '/webdav';

// Create authorization header
function getAuthHeader() {
  // Return minimal headers without auth since authentication is handled server-side
  return {
    'Content-Type': 'application/octet-stream' // Default for file uploads
  };
}

// Function to list directory contents using PROPFIND
export async function listDirectory(path) {
  const response = await fetch(`${davBasePath}${path}`, {
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

    if (!href) continue;

    // Convert href to a relative path by removing the WebDAV_URL prefix
    let relativeHref = decodeURIComponent(href);
    if (relativeHref.startsWith(davBasePath)) {
      relativeHref = relativeHref.substring(davBasePath.length);
      // Ensure path starts with / after removing prefix
      if (!relativeHref.startsWith('/')) {
        relativeHref = '/' + relativeHref;
      }
    }
    // Always ensure directory paths end with '/' (except root)
    if (response.getElementsByTagName('iscollection')[0]?.textContent === '1' && relativeHref !== '/' && !relativeHref.endsWith('/')) {
      relativeHref = relativeHref + '/';
    }

    // The path variable passed to this function is already the path relative to WebDAV root
    // Skip the entry for the current directory itself
    if (relativeHref === path || relativeHref === path + '/') {
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
    const isCollection = prop.getElementsByTagName('iscollection')[0]?.textContent === '1';

    // Extract the actual name from the relative path
    const hrefParts = relativeHref.split('/').filter(part => part !== '');
    const currentPathParts = path.split('/').filter(part => part !== '');

    let name;
    if (hrefParts.length > currentPathParts.length) {
      name = hrefParts[currentPathParts.length];
    } else {
      continue; // Skip if it's not a direct child
    }

    // Skip the parent directory indicator if present
    if (name === '..') continue;

    items.push({
      name: displayName || name,
      path: relativeHref,
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
  // Create the full file path
  const fullPath = path + file.name;
  
  const response = await fetch(`${davBasePath}${fullPath}`, {
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
  const response = await fetch(`${davBasePath}${path}`, {
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
  const response = await fetch(`${davBasePath}${path}`, {
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
  // Create the full directory path
  const fullPath = path + dirName;
  
  const response = await fetch(`${davBasePath}${fullPath}`, {
    method: 'MKCOL',
    headers: getAuthHeader()
  });

  if (!response.ok) {
    throw new Error(`Failed to create directory: ${response.status} ${response.statusText}`);
  }
  
  return response;
}
