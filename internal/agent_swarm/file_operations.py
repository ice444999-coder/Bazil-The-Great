#!/usr/bin/env python3
"""
ARES File Operations Module - Direct file system access for SOLACE

This module provides file system operations for the SOLACE orchestration system,
enabling autonomous file reading, writing, backup, and restoration capabilities.
"""

import os
import shutil
from pathlib import Path
from datetime import datetime
from typing import List, Dict, Optional


def read_file(file_path: str) -> str:
    """
    Read file contents and return as string.
    
    Args:
        file_path: Path to file to read (absolute or relative)
        
    Returns:
        str: File contents as UTF-8 string
        
    Raises:
        FileNotFoundError: If file does not exist
        PermissionError: If file cannot be read
    """
    path = Path(file_path)
    
    if not path.exists():
        raise FileNotFoundError(f"File not found: {file_path}")
    
    if not path.is_file():
        raise ValueError(f"Path is not a file: {file_path}")
    
    try:
        # Try UTF-8 first
        return path.read_text(encoding='utf-8')
    except UnicodeDecodeError:
        # Fallback: read as binary and return message
        return f"[BINARY FILE: {file_path}, Size: {path.stat().st_size} bytes]"


def write_file(file_path: str, content: str) -> bool:
    """
    Write content to file, creating parent directories if needed.
    
    Args:
        file_path: Path to file to write (absolute or relative)
        content: String content to write
        
    Returns:
        bool: True if successful
        
    Raises:
        PermissionError: If file cannot be written
        OSError: If directory creation fails
    """
    path = Path(file_path)
    
    # Create parent directories if they don't exist
    path.parent.mkdir(parents=True, exist_ok=True)
    
    # Write file
    path.write_text(content, encoding='utf-8')
    
    return True


def list_directory(dir_path: str, recursive: bool = True, max_depth: int = 5) -> List[Dict]:
    """
    List directory contents with metadata.
    
    Args:
        dir_path: Path to directory to list
        recursive: If True, list subdirectories recursively
        max_depth: Maximum recursion depth (default: 5)
        
    Returns:
        List[Dict]: List of file/directory metadata dicts with keys:
            - path: str (relative to dir_path)
            - type: str ('file' or 'directory')
            - size: int (bytes, only for files)
            - modified: str (ISO format timestamp)
            
    Raises:
        FileNotFoundError: If directory does not exist
        NotADirectoryError: If path is not a directory
    """
    base_path = Path(dir_path)
    
    if not base_path.exists():
        raise FileNotFoundError(f"Directory not found: {dir_path}")
    
    if not base_path.is_dir():
        raise NotADirectoryError(f"Path is not a directory: {dir_path}")
    
    results = []
    
    def _scan_dir(current_path: Path, depth: int = 0):
        """Recursively scan directory."""
        if depth > max_depth:
            return
        
        try:
            for item in current_path.iterdir():
                try:
                    # Get relative path
                    rel_path = item.relative_to(base_path)
                    
                    # Get file stats
                    stat = item.stat()
                    modified = datetime.fromtimestamp(stat.st_mtime).isoformat()
                    
                    entry = {
                        'path': str(rel_path),
                        'type': 'directory' if item.is_dir() else 'file',
                        'modified': modified
                    }
                    
                    # Add size for files
                    if item.is_file():
                        entry['size'] = stat.st_size
                    
                    results.append(entry)
                    
                    # Recurse into subdirectories
                    if recursive and item.is_dir():
                        _scan_dir(item, depth + 1)
                        
                except PermissionError:
                    # Skip files/dirs we can't access
                    continue
                    
        except PermissionError:
            # Skip directories we can't access
            pass
    
    _scan_dir(base_path)
    return results


def create_backup(workspace_path: str, backup_base_dir: str = "C:\\ARES_Backups") -> str:
    """
    Create timestamped backup of workspace.
    
    Args:
        workspace_path: Path to workspace to backup
        backup_base_dir: Base directory for backups (default: C:\\ARES_Backups)
        
    Returns:
        str: Path to created backup directory
        
    Raises:
        FileNotFoundError: If workspace does not exist
        OSError: If backup creation fails
    """
    workspace = Path(workspace_path)
    
    if not workspace.exists():
        raise FileNotFoundError(f"Workspace not found: {workspace_path}")
    
    # Create backup base directory if needed
    backup_base = Path(backup_base_dir)
    backup_base.mkdir(parents=True, exist_ok=True)
    
    # Generate timestamp-based backup name
    timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
    backup_name = f"backup_{timestamp}"
    backup_path = backup_base / backup_name
    
    # Copy entire workspace
    shutil.copytree(
        workspace,
        backup_path,
        ignore=shutil.ignore_patterns(
            '.git', 'node_modules', 'bin', 'obj', '__pycache__',
            '*.pyc', '*.pyo', '*.exe', '*.dll', '.vs', '.vscode'
        )
    )
    
    return str(backup_path)


def restore_backup(backup_path: str, workspace_path: str) -> bool:
    """
    Restore workspace from backup.
    
    WARNING: This will DELETE the existing workspace and replace it with the backup!
    
    Args:
        backup_path: Path to backup directory to restore from
        workspace_path: Path to workspace to restore to
        
    Returns:
        bool: True if successful
        
    Raises:
        FileNotFoundError: If backup does not exist
        OSError: If restoration fails
    """
    backup = Path(backup_path)
    workspace = Path(workspace_path)
    
    if not backup.exists():
        raise FileNotFoundError(f"Backup not found: {backup_path}")
    
    if not backup.is_dir():
        raise NotADirectoryError(f"Backup path is not a directory: {backup_path}")
    
    # Remove existing workspace
    if workspace.exists():
        shutil.rmtree(workspace)
    
    # Copy backup to workspace location
    shutil.copytree(backup, workspace)
    
    return True


def _test_module():
    """Test the file operations module."""
    print("=" * 60)
    print("ARES File Operations Module - Test Suite")
    print("=" * 60)
    
    # Test 1: List current directory
    print("\n[TEST 1] Listing current directory (non-recursive)...")
    try:
        current_dir = Path(__file__).parent
        files = list_directory(str(current_dir), recursive=False)
        print(f"✅ Found {len(files)} items in {current_dir.name}/")
        for item in files[:5]:  # Show first 5
            type_icon = "📁" if item['type'] == 'directory' else "📄"
            size_str = f" ({item.get('size', 0)} bytes)" if item['type'] == 'file' else ""
            print(f"   {type_icon} {item['path']}{size_str}")
        if len(files) > 5:
            print(f"   ... and {len(files) - 5} more items")
    except Exception as e:
        print(f"❌ FAILED: {e}")
    
    # Test 2: Read this file
    print("\n[TEST 2] Reading this module file...")
    try:
        content = read_file(__file__)
        lines = content.split('\n')
        print(f"✅ Read {len(lines)} lines from {Path(__file__).name}")
        print(f"   First line: {lines[0][:60]}...")
    except Exception as e:
        print(f"❌ FAILED: {e}")
    
    # Test 3: Write test file
    print("\n[TEST 3] Writing test file...")
    try:
        test_file = Path(__file__).parent / "test_output.txt"
        test_content = f"ARES Test - {datetime.now().isoformat()}\nFile operations working!"
        write_file(str(test_file), test_content)
        print(f"✅ Wrote test file: {test_file.name}")
        
        # Read it back
        read_back = read_file(str(test_file))
        if test_content == read_back:
            print("✅ Content verification: PASS")
        else:
            print("❌ Content verification: FAIL")
        
        # Clean up
        test_file.unlink()
        print("✅ Cleaned up test file")
    except Exception as e:
        print(f"❌ FAILED: {e}")
    
    # Test 4: List with recursion
    print("\n[TEST 4] Listing parent directory (recursive, max_depth=2)...")
    try:
        parent_dir = Path(__file__).parent.parent
        files = list_directory(str(parent_dir), recursive=True, max_depth=2)
        file_count = sum(1 for f in files if f['type'] == 'file')
        dir_count = sum(1 for f in files if f['type'] == 'directory')
        print(f"✅ Found {file_count} files and {dir_count} directories")
        print(f"   Total items: {len(files)}")
    except Exception as e:
        print(f"❌ FAILED: {e}")
    
    print("\n" + "=" * 60)
    print("Test suite complete!")
    print("=" * 60)


if __name__ == "__main__":
    _test_module()
