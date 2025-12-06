// WebDAV utility functions for File Hub web UI

// Base path for WebDAV requests
const WEBDAV_BASE_PATH = '/dav';

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
  const response = await fetch(`${WEBDAV_BASE_PATH}${path}`, {
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

    // Convert href to a relative path by removing the WebDAV URL prefix
    let relativeHref = decodeURIComponent(href);
    if (relativeHref.startsWith(WEBDAV_BASE_PATH)) {
      relativeHref = relativeHref.substring(WEBDAV_BASE_PATH.length);
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

/**
 * Upload a file
 * @param {string} path - Directory path to upload to
 * @param {File} file - File object to upload
 * @returns {Promise<Response>} Fetch response
 */
export async function uploadFile(path, file) {
  // Create the full file path
  const fullPath = path + file.name;

  const response = await fetch(`${WEBDAV_BASE_PATH}${fullPath}`, {
    method: 'PUT',
    headers: createRequestHeaders(),
    body: file
  });

  if (!response.ok) {
    handleHttpError(response, 'Failed to upload file');
  }

  return response;
}

/*
 * The following functions are not currently used in the UI but are kept for future expansion
 */

/*
 * Download a file
 * @param {string} path - Path to the file
 * @returns {Promise<Blob>} File blob
 *
export async function downloadFile(path) {
  const response = await fetch(`${WEBDAV_BASE_PATH}${path}`, {
    method: 'GET',
    headers: createRequestHeaders()
  });

  if (!response.ok) {
    throw new Error(`Failed to download file: ${response.status} ${response.statusText}`);
  }

  return response.blob();
}
*/

/*
 * Delete a file or directory
 * @param {string} path - Path to the file or directory
 * @returns {Promise<Response>} Fetch response
 *
export async function deleteFile(path) {
  const response = await fetch(`${WEBDAV_BASE_PATH}${path}`, {
    method: 'DELETE',
    headers: createRequestHeaders()
  });

  if (!response.ok) {
    throw new Error(`Failed to delete file: ${response.status} ${response.statusText}`);
  }

  return response;
}
*/

/*
 * Create a directory
 * @param {string} path - Parent directory path
 * @param {string} dirName - Name of the directory to create
 * @returns {Promise<Response>} Fetch response
 *
export async function createDirectory(path, dirName) {
  // Create the full directory path
  const fullPath = path + dirName;

  const response = await fetch(`${WEBDAV_BASE_PATH}${fullPath}`, {
    method: 'MKCOL',
    headers: createRequestHeaders()
  });

  if (!response.ok) {
    throw new Error(`Failed to create directory: ${response.status} ${response.statusText}`);
  }

  return response;
}
*/
