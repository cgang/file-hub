export function getIcon(file) {
  if (file.type === 'directory') {
    return '📁';  // Folder icon
  }

  // Determine icon based on file extension
  const ext = file.name.split('.').pop().toLowerCase();
  switch (ext) {
    case 'pdf':
      return '📄';
    case 'jpg':
    case 'jpeg':
    case 'png':
    case 'gif':
    case 'webp':
      return '🖼️';
    case 'txt':
      return '📝';
    case 'doc':
    case 'docx':
      return '📝';
    case 'xls':
    case 'xlsx':
      return '📊';
    case 'ppt':
    case 'pptx':
      return '📊';
    case 'zip':
    case 'rar':
    case '7z':
      return '📦';
    case 'mp3':
    case 'wav':
    case 'flac':
      return '🎵';
    case 'mp4':
    case 'avi':
    case 'mov':
      return '🎬';
    default:
      return '📎';
  }
}

export function formatSize(bytes) {
  if (bytes === undefined || bytes === null) return '-';
  
  if (bytes < 1024) return bytes + ' B';
  else if (bytes < 1048576) return (bytes / 1024).toFixed(1) + ' KB';
  else if (bytes < 1073741824) return (bytes / 1048576).toFixed(1) + ' MB';
  else return (bytes / 1073741824).toFixed(1) + ' GB';
}
