// WebDAV utility functions for File Hub web UI

let _davPath = null;

/**
 * Fetch DAV path from /api/hello
 * @returns {Promise<string>} The DAV base path
 */
async function fetchDavPath() {
  if (_davPath) return _davPath;
  
  const response = await fetch('/api/hello');
  if (!response.ok) {
    throw new Error('Failed to fetch DAV path');
  }
  
  const data = await response.json();
  if (!data.dav) {
    throw new Error('DAV path not found in API response');
  }
  
  _davPath = data.dav;
  return _davPath;
}

/**
 * Get DAV path with caching
 * @returns {Promise<string>} The DAV base path
 */
export async function getDavPath() {
  return _davPath || fetchDavPath();
}

/**
 * Create authorization header for WebDAV requests
 * @returns {Object} Headers object with Content-Type
 */
function createRequestHeaders() {
  return {
    'Content-Type': 'application/octet-stream' // Default for file uploads
  };
}

/**
 * Handle HTTP errors by parsing response text and creating an appropriate error object
 * @param {Response} response - The fetch response object
 * @param {string} prefix - Operation-specific prefix for the error message
 * @returns {never} - Throws an error
 */
async function handleHttpError(response, prefix) {
  const text = await response.text();
  let errorText = text;
  try {
    const parser = new DOMParser();
    const xmlDoc = parser.parseFromString(text, 'application/xml');
    const errorElement = xmlDoc.getElementsByTagName('error')[0];
    if (errorElement) {
      errorText = errorElement.textContent || text;
    }
  } catch (parseError) {
    // If parsing fails, use the original text
  }
  throw new Error(`${prefix}: ${response.status} ${response.statusText} - ${errorText}`);
}

/**
 * List directory contents using PROPFIND
 * @param {string} path - Path to list
 * @returns {Promise<Array>} Array of file/directory objects
 */
export async function listDirectory(path) {
  const davPath = await getDavPath();
const response = await fetch(`${davPath}${path}`, {
    method: 'PROPFIND',
    headers: {
      ...createRequestHeaders(),
      'Depth': '1'  // Only get immediate children
    }
  });

  if (!response.ok) {
    handleHttpError(response, 'Failed to list directory');
  }

  const responseText = await response.text();

  // Parse the XML response from WebDAV
  const parser = new DOMParser();
  const xmlDoc = parser.parseFromString(responseText, 'application/xml');

  // Extract file and directory information
  const responses = xmlDoc.getElementsByTagName('response');
  const items = [];

  for (let i = 0; i < responses.length; i++) {
    const response = responses[i];
    const href = response.getElementsByTagName('href')[0]?.textContent;
    if (!href) continue;

    const propstat = response.getElementsByTagName('propstat')[0];
    if (!propstat) continue;

    const prop = propstat.getElementsByTagName('prop')[0];
    if (!prop) continue;

    const displayName = prop.getElementsByTagName('displayname')[0]?.textContent || '';
    const contentType = prop.getElementsByTagName('getcontenttype')[0]?.textContent;
    const contentLength = prop.getElementsByTagName('getcontentlength')[0]?.textContent;
    const lastModified = prop.getElementsByTagName('getlastmodified')[0]?.textContent;
    const isCollection = prop.getElementsByTagName('iscollection')[0]?.textContent === '1';

    // Convert href to a relative path by removing the WebDAV URL prefix
    let relativeHref = decodeURIComponent(href);
    const davPath = await getDavPath();
    if (relativeHref.startsWith(davPath)) {
      relativeHref = relativeHref.substring(davPath.length);
      // Ensure path starts with / after removing prefix
      if (!relativeHref.startsWith('/')) {
        relativeHref = '/' + relativeHref;
      }
    }
    // Always ensure directory paths end with '/' (except root)
    if (isCollection && !relativeHref.endsWith('/')) {
      relativeHref = relativeHref + '/';
    }

    // The path variable passed to this function is already the path relative to WebDAV root
    // Skip the entry for the current directory itself
    if (relativeHref === path || relativeHref === path + '/') {
      continue;
    }

    // Skip the parent directory indicator if present
    if (displayName === '..') continue;

    items.push({
      name: displayName,
      path: relativeHref,
      type: isCollection ? 'directory' : 'file',
      size: contentLength ? parseInt(contentLength) : undefined,
      lastModified: lastModified,
      contentType: contentType
    });
  }

  return items;
}

/**
 * Upload a file
 * @param {string} path - Directory path to upload to
 * @param {File} file - File object to upload
 * @returns {Promise<Response>} Fetch response
 */
export async function uploadFile(path, file) {
  // Create the full file path
  const fullPath = path + file.name;

  const davPath = await getDavPath();
const response = await fetch(`${davPath}${fullPath}`, {
    method: 'PUT',
    headers: createRequestHeaders(),
    body: file
  });

  if (!response.ok) {
    handleHttpError(response, 'Failed to upload file');
  }

  return response;
}
