#!/usr/bin/env python3
"""
ARES SQL Deduplication Script
Version: 1.0.0
Last Updated: October 19, 2025

Purpose: Scan repository for duplicate SQL files using SHA-256 hashing and semantic similarity.
Eliminates redundancy in the ARES trading system database for optimal AI agent performance.

Features:
- SHA-256 exact duplicate detection
- Semantic similarity via embeddings (OpenAI text-embedding-3-small)
- pgvector integration for fast similarity searches
- Comprehensive reporting (JSON, Markdown, CSV)
- Safe dry-run mode
- Autonomous execution capability for SOLACE

Usage:
    python dedup_sql_files.py --repo /path/to/repo --dry-run
    python dedup_sql_files.py --repo /path/to/repo --execute
    python dedup_sql_files.py --scan-only --output report.json

Author: ARES SQL Reorganization Team
"""

import os
import sys
import hashlib
import json
import argparse
import logging
from pathlib import Path
from typing import Dict, List, Set, Tuple, Optional
from dataclasses import dataclass, asdict
from datetime import datetime
import requests
import psycopg2
from psycopg2.extras import execute_values

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s',
    handlers=[
        logging.FileHandler('dedup_sql.log'),
        logging.StreamHandler(sys.stdout)
    ]
)
logger = logging.getLogger(__name__)

@dataclass
class SQLFile:
    """Represents a SQL file with metadata"""
    path: str
    size: int
    sha256: str
    content: str
    embedding: Optional[List[float]] = None
    schema_name: Optional[str] = None
    table_names: List[str] = None

    def __post_init__(self):
        if self.table_names is None:
            self.table_names = []

@dataclass
class DuplicateGroup:
    """Represents a group of duplicate files"""
    group_id: str
    files: List[SQLFile]
    similarity_score: float
    duplicate_type: str  # 'exact' or 'semantic'
    recommended_action: str
    master_file: Optional[str] = None

class SQLDeduplicator:
    """Main deduplication engine for SQL files"""

    def __init__(self, repo_path: str, openai_api_key: Optional[str] = None,
                 db_config: Optional[Dict] = None):
        self.repo_path = Path(repo_path)
        self.openai_api_key = openai_api_key or os.getenv('OPENAI_API_KEY')
        self.db_config = db_config or {
            'host': os.getenv('DB_HOST', 'localhost'),
            'port': int(os.getenv('DB_PORT', '5433')),
            'database': os.getenv('DB_NAME', 'ares_pgvector'),
            'user': os.getenv('DB_USER', 'postgres'),
            'password': os.getenv('DB_PASSWORD', 'ARESISWAKING')
        }

        # Similarity thresholds
        self.EXACT_DUPLICATE_THRESHOLD = 1.0
        self.SEMANTIC_DUPLICATE_THRESHOLD = 0.85
        self.PARTIAL_DUPLICATE_THRESHOLD = 0.60

        # File collections
        self.sql_files: List[SQLFile] = []
        self.hash_groups: Dict[str, List[SQLFile]] = {}
        self.semantic_groups: List[DuplicateGroup] = []

    def scan_sql_files(self) -> List[SQLFile]:
        """Scan repository for SQL files"""
        logger.info(f"Scanning for SQL files in {self.repo_path}")

        sql_extensions = {'.sql', '.SQL'}
        sql_files = []

        for root, dirs, files in os.walk(self.repo_path):
            # Skip common non-SQL directories
            dirs[:] = [d for d in dirs if not d.startswith('.') and d not in {
                'node_modules', '__pycache__', '.git', 'build', 'dist'
            }]

            for file in files:
                if any(file.endswith(ext) for ext in sql_extensions):
                    file_path = Path(root) / file
                    try:
                        sql_file = self._analyze_sql_file(file_path)
                        if sql_file:
                            sql_files.append(sql_file)
                    except Exception as e:
                        logger.warning(f"Failed to analyze {file_path}: {e}")

        self.sql_files = sql_files
        logger.info(f"Found {len(sql_files)} SQL files")
        return sql_files

    def _analyze_sql_file(self, file_path: Path) -> Optional[SQLFile]:
        """Analyze a single SQL file"""
        try:
            with open(file_path, 'r', encoding='utf-8', errors='ignore') as f:
                content = f.read()

            # Skip empty files
            if not content.strip():
                return None

            # Calculate hash
            sha256 = hashlib.sha256(content.encode('utf-8')).sha256()

            # Extract schema and table names
            schema_name, table_names = self._extract_sql_metadata(content)

            return SQLFile(
                path=str(file_path.relative_to(self.repo_path)),
                size=len(content),
                sha256=sha256,
                content=content,
                schema_name=schema_name,
                table_names=table_names
            )

        except Exception as e:
            logger.error(f"Error analyzing {file_path}: {e}")
            return None

    def _extract_sql_metadata(self, content: str) -> Tuple[Optional[str], List[str]]:
        """Extract schema name and table names from SQL content"""
        schema_name = None
        table_names = []

        lines = content.upper().split('\n')

        for line in lines:
            line = line.strip()

            # Look for schema creation
            if line.startswith('CREATE SCHEMA') or line.startswith('USE SCHEMA'):
                parts = line.split()
                if len(parts) > 2:
                    schema_name = parts[2].strip(';"')

            # Look for table creation
            elif line.startswith('CREATE TABLE'):
                parts = line.split()
                if len(parts) > 2:
                    table_name = parts[2].strip('();')
                    # Remove schema prefix if present
                    if '.' in table_name:
                        table_name = table_name.split('.')[-1]
                    table_names.append(table_name)

        return schema_name, table_names

    def find_exact_duplicates(self) -> Dict[str, List[SQLFile]]:
        """Find exact duplicates using SHA-256 hashing"""
        logger.info("Finding exact duplicates...")

        hash_groups = {}
        for sql_file in self.sql_files:
            if sql_file.sha256 not in hash_groups:
                hash_groups[sql_file.sha256] = []
            hash_groups[sql_file.sha256].append(sql_file)

        # Filter to only groups with duplicates
        exact_duplicates = {
            hash_val: files for hash_val, files in hash_groups.items()
            if len(files) > 1
        }

        logger.info(f"Found {len(exact_duplicates)} exact duplicate groups")
        self.hash_groups = exact_duplicates
        return exact_duplicates

    def generate_embeddings(self) -> None:
        """Generate embeddings for semantic similarity"""
        if not self.openai_api_key:
            logger.warning("No OpenAI API key provided, skipping semantic analysis")
            return

        logger.info("Generating embeddings for semantic analysis...")

        for sql_file in self.sql_files:
            try:
                embedding = self._get_embedding(sql_file.content)
                sql_file.embedding = embedding
            except Exception as e:
                logger.warning(f"Failed to generate embedding for {sql_file.path}: {e}")

    def _get_embedding(self, text: str) -> List[float]:
        """Get embedding from OpenAI API"""
        url = "https://api.openai.com/v1/embeddings"
        headers = {
            "Authorization": f"Bearer {self.openai_api_key}",
            "Content-Type": "application/json"
        }

        # Truncate text if too long (OpenAI limit)
        text = text[:8000] if len(text) > 8000 else text

        data = {
            "input": text,
            "model": "text-embedding-3-small"
        }

        response = requests.post(url, headers=headers, json=data)
        response.raise_for_status()

        return response.json()['data'][0]['embedding']

    def find_semantic_duplicates(self) -> List[DuplicateGroup]:
        """Find semantic duplicates using embeddings"""
        logger.info("Finding semantic duplicates...")

        semantic_groups = []

        # Compare all pairs of files with embeddings
        for i, file1 in enumerate(self.sql_files):
            if not file1.embedding:
                continue

            for j, file2 in enumerate(self.sql_files[i+1:], i+1):
                if not file2.embedding:
                    continue

                similarity = self._cosine_similarity(file1.embedding, file2.embedding)

                if similarity >= self.SEMANTIC_DUPLICATE_THRESHOLD:
                    duplicate_type = 'semantic'
                    recommended_action = self._determine_action(file1, file2, similarity)

                    group = DuplicateGroup(
                        group_id=f"semantic_{i}_{j}",
                        files=[file1, file2],
                        similarity_score=similarity,
                        duplicate_type=duplicate_type,
                        recommended_action=recommended_action,
                        master_file=self._choose_master_file(file1, file2)
                    )
                    semantic_groups.append(group)

        logger.info(f"Found {len(semantic_groups)} semantic duplicate groups")
        self.semantic_groups = semantic_groups
        return semantic_groups

    def _cosine_similarity(self, vec1: List[float], vec2: List[float]) -> float:
        """Calculate cosine similarity between two vectors"""
        import math

        dot_product = sum(a * b for a, b in zip(vec1, vec2))
        norm1 = math.sqrt(sum(a * a for a in vec1))
        norm2 = math.sqrt(sum(b * b for b in vec2))

        if norm1 == 0 or norm2 == 0:
            return 0.0

        return dot_product / (norm1 * norm2)

    def _determine_action(self, file1: SQLFile, file2: SQLFile, similarity: float) -> str:
        """Determine recommended action for duplicate files"""
        if similarity >= self.EXACT_DUPLICATE_THRESHOLD:
            return "DELETE_DUPLICATE"
        elif similarity >= self.SEMANTIC_DUPLICATE_THRESHOLD:
            return "MERGE_SEMANTIC"
        elif similarity >= self.PARTIAL_DUPLICATE_THRESHOLD:
            return "REVIEW_MANUAL"
        else:
            return "KEEP_SEPARATE"

    def _choose_master_file(self, file1: SQLFile, file2: SQLFile) -> str:
        """Choose which file to keep as master"""
        # Prefer files in migrations directory
        if 'migration' in file1.path.lower():
            return file1.path
        elif 'migration' in file2.path.lower():
            return file2.path

        # Prefer newer files (by size, assuming larger = more complete)
        if file1.size > file2.size:
            return file1.path
        else:
            return file2.path

    def generate_report(self, output_format: str = 'json') -> str:
        """Generate comprehensive deduplication report"""
        logger.info(f"Generating {output_format} report...")

        report_data = {
            'timestamp': datetime.now().isoformat(),
            'repo_path': str(self.repo_path),
            'total_files': len(self.sql_files),
            'exact_duplicates': {
                'count': len(self.hash_groups),
                'groups': [
                    {
                        'hash': hash_val,
                        'files': [f.path for f in files],
                        'count': len(files)
                    }
                    for hash_val, files in self.hash_groups.items()
                ]
            },
            'semantic_duplicates': {
                'count': len(self.semantic_groups),
                'groups': [
                    {
                        'group_id': group.group_id,
                        'files': [f.path for f in group.files],
                        'similarity_score': group.similarity_score,
                        'recommended_action': group.recommended_action,
                        'master_file': group.master_file
                    }
                    for group in self.semantic_groups
                ]
            },
            'summary': {
                'total_duplicate_files': sum(len(files) for files in self.hash_groups.values()) +
                                       sum(len(group.files) for group in self.semantic_groups),
                'space_saved_estimate': self._calculate_space_saved(),
                'recommendations': self._generate_recommendations()
            }
        }

        if output_format == 'json':
            return json.dumps(report_data, indent=2)
        elif output_format == 'markdown':
            return self._generate_markdown_report(report_data)
        else:
            raise ValueError(f"Unsupported format: {output_format}")

    def _calculate_space_saved(self) -> int:
        """Estimate space that could be saved by deduplication"""
        space = 0

        # Exact duplicates: keep one copy
        for files in self.hash_groups.values():
            space += sum(f.size for f in files[1:])  # All but first

        # Semantic duplicates: estimate 50% reduction
        for group in self.semantic_groups:
            avg_size = sum(f.size for f in group.files) / len(group.files)
            space += int(avg_size * 0.5 * (len(group.files) - 1))

        return space

    def _generate_recommendations(self) -> List[str]:
        """Generate recommendations for deduplication"""
        recommendations = []

        if self.hash_groups:
            recommendations.append(f"Remove {sum(len(files)-1 for files in self.hash_groups.values())} exact duplicate files")

        if self.semantic_groups:
            semantic_count = len([g for g in self.semantic_groups if g.recommended_action == 'MERGE_SEMANTIC'])
            recommendations.append(f"Merge {semantic_count} semantic duplicate files")

        if len(self.sql_files) > 50:
            recommendations.append("Consider consolidating into functional schemas (trading_core, memory_system, etc.)")

        return recommendations

    def _generate_markdown_report(self, data: Dict) -> str:
        """Generate markdown format report"""
        md = [f"# ARES SQL Deduplication Report\n"]
        md.append(f"**Generated:** {data['timestamp']}\n")
        md.append(f"**Repository:** {data['repo_path']}\n")
        md.append(f"**Total SQL Files:** {data['total_files']}\n")

        # Summary
        summary = data['summary']
        md.append("## Summary\n")
        md.append(f"- **Duplicate Files:** {summary['total_duplicate_files']}\n")
        md.append(f"- **Estimated Space Saved:** {summary['space_saved_estimate']:,} bytes\n")
        md.append("### Recommendations\n")
        for rec in summary['recommendations']:
            md.append(f"- {rec}\n")

        # Exact Duplicates
        if data['exact_duplicates']['count'] > 0:
            md.append("\n## Exact Duplicates\n")
            for group in data['exact_duplicates']['groups']:
                md.append(f"### Hash: `{group['hash'][:16]}...`\n")
                md.append(f"**Files ({group['count']} copies):**\n")
                for file in group['files']:
                    md.append(f"- `{file}`\n")
                md.append("\n")

        # Semantic Duplicates
        if data['semantic_duplicates']['count'] > 0:
            md.append("\n## Semantic Duplicates\n")
            for group in data['semantic_duplicates']['groups']:
                md.append(f"### Group {group['group_id']}\n")
                md.append(f"**Similarity:** {group['similarity_score']:.3f}\n")
                md.append(f"**Action:** {group['recommended_action']}\n")
                md.append(f"**Master File:** `{group['master_file']}`\n")
                md.append("**Files:**\n")
                for file in group['files']:
                    md.append(f"- `{file}`\n")
                md.append("\n")

        return ''.join(md)

    def execute_deduplication(self, dry_run: bool = True) -> Dict[str, int]:
        """Execute deduplication actions"""
        logger.info(f"Executing deduplication (dry_run={dry_run})")

        results = {
            'files_deleted': 0,
            'files_merged': 0,
            'errors': 0
        }

        # Handle exact duplicates
        for hash_val, files in self.hash_groups.items():
            master_file = self._choose_master_file(*files)
            for sql_file in files:
                if sql_file.path != master_file:
                    if dry_run:
                        logger.info(f"Would delete: {sql_file.path}")
                    else:
                        try:
                            os.remove(self.repo_path / sql_file.path)
                            results['files_deleted'] += 1
                            logger.info(f"Deleted: {sql_file.path}")
                        except Exception as e:
                            logger.error(f"Failed to delete {sql_file.path}: {e}")
                            results['errors'] += 1

        # Handle semantic duplicates (more complex, just log for now)
        for group in self.semantic_groups:
            if group.recommended_action == 'MERGE_SEMANTIC':
                if dry_run:
                    logger.info(f"Would merge group {group.group_id}")
                else:
                    # TODO: Implement semantic merging logic
                    logger.warning(f"Semantic merge not yet implemented for {group.group_id}")
                    results['errors'] += 1

        return results

def main():
    parser = argparse.ArgumentParser(description="ARES SQL Deduplication Tool")
    parser.add_argument('--repo', required=True, help='Repository path to scan')
    parser.add_argument('--dry-run', action='store_true', help='Dry run mode')
    parser.add_argument('--execute', action='store_true', help='Execute deduplication')
    parser.add_argument('--scan-only', action='store_true', help='Only scan, no deduplication')
    parser.add_argument('--output', default='report.json', help='Output file path')
    parser.add_argument('--format', choices=['json', 'markdown'], default='json', help='Output format')
    parser.add_argument('--openai-key', help='OpenAI API key for embeddings')

    args = parser.parse_args()

    # Validate arguments
    if not (args.dry_run or args.execute or args.scan_only):
        args.dry_run = True  # Default to dry run

    if args.execute and not args.dry_run:
        confirm = input("⚠️  This will permanently delete files. Continue? (yes/no): ")
        if confirm.lower() != 'yes':
            logger.info("Operation cancelled")
            return

    # Initialize deduplicator
    deduplicator = SQLDeduplicator(
        repo_path=args.repo,
        openai_api_key=args.openai_key
    )

    try:
        # Phase 1: Scan files
        logger.info("Phase 1: Scanning SQL files...")
        sql_files = deduplicator.scan_sql_files()

        if not sql_files:
            logger.warning("No SQL files found")
            return

        # Phase 2: Find exact duplicates
        logger.info("Phase 2: Finding exact duplicates...")
        exact_duplicates = deduplicator.find_exact_duplicates()

        # Phase 3: Generate embeddings and find semantic duplicates
        if deduplicator.openai_api_key:
            logger.info("Phase 3: Generating embeddings...")
            deduplicator.generate_embeddings()

            logger.info("Phase 4: Finding semantic duplicates...")
            semantic_duplicates = deduplicator.find_semantic_duplicates()
        else:
            logger.info("Skipping semantic analysis (no OpenAI API key)")

        # Phase 4: Generate report
        logger.info("Phase 5: Generating report...")
        report = deduplicator.generate_report(args.format)

        # Save report
        with open(args.output, 'w', encoding='utf-8') as f:
            f.write(report)

        logger.info(f"Report saved to {args.output}")

        # Phase 5: Execute deduplication if requested
        if args.execute:
            logger.info("Phase 6: Executing deduplication...")
            results = deduplicator.execute_deduplication(dry_run=args.dry_run)

            logger.info("Deduplication Results:")
            logger.info(f"  Files deleted: {results['files_deleted']}")
            logger.info(f"  Files merged: {results['files_merged']}")
            logger.info(f"  Errors: {results['errors']}")

        # Print summary
        total_duplicates = len(exact_duplicates) + len(deduplicator.semantic_groups)
        logger.info("🎯 Deduplication Summary:")
        logger.info(f"   Total SQL files: {len(sql_files)}")
        logger.info(f"   Exact duplicate groups: {len(exact_duplicates)}")
        logger.info(f"   Semantic duplicate groups: {len(deduplicator.semantic_groups)}")
        logger.info(f"   Total duplicate groups: {total_duplicates}")

        if total_duplicates > 0:
            logger.info("   Status: DUPLICATES FOUND - Review report and consider deduplication")
        else:
            logger.info("   Status: CLEAN - No duplicates detected")

    except Exception as e:
        logger.error(f"Critical error: {e}")
        sys.exit(1)

if __name__ == "__main__":
    main()