"""
💾 Investigation storage layer
Provides persistence for investigations and analyses using SQLite
"""

import json
import logging
import sqlite3
from datetime import datetime
from pathlib import Path
from typing import List, Optional

from .investigation import Investigation, InvestigationStatus

logger = logging.getLogger(__name__)


class InvestigationStorage:
    """Storage for Sift investigations using SQLite"""

    def __init__(self, db_path: str = "/tmp/sift_investigations.db"):
        """Initialize storage"""
        self.db_path = Path(db_path)
        self.db_path.parent.mkdir(parents=True, exist_ok=True)
        self._init_db()

    def _init_db(self) -> None:
        """Initialize database schema"""
        with sqlite3.connect(self.db_path) as conn:
            conn.execute(
                """
                CREATE TABLE IF NOT EXISTS investigations (
                    id TEXT PRIMARY KEY,
                    name TEXT NOT NULL,
                    labels TEXT NOT NULL,
                    start_time TEXT NOT NULL,
                    end_time TEXT,
                    status TEXT NOT NULL,
                    analyses TEXT NOT NULL,
                    created_at TEXT NOT NULL,
                    updated_at TEXT NOT NULL,
                    metadata TEXT NOT NULL
                )
            """
            )
            conn.commit()
            logger.info(f"💾 Initialized investigation storage at {self.db_path}")

    def save_investigation(self, investigation: Investigation) -> None:
        """Save or update an investigation"""
        investigation.updated_at = datetime.utcnow()
        data = investigation.to_dict()

        with sqlite3.connect(self.db_path) as conn:
            conn.execute(
                """
                INSERT OR REPLACE INTO investigations
                (id, name, labels, start_time, end_time, status, analyses, created_at, updated_at, metadata)
                VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
            """,
                (
                    data["id"],
                    data["name"],
                    json.dumps(data["labels"]),
                    data["start_time"],
                    data["end_time"],
                    data["status"],
                    json.dumps(data["analyses"]),
                    data["created_at"],
                    data["updated_at"],
                    json.dumps(data["metadata"]),
                ),
            )
            conn.commit()
            logger.info(f"💾 Saved investigation {investigation.id}")

    def get_investigation(self, investigation_id: str) -> Optional[Investigation]:
        """Get an investigation by ID"""
        with sqlite3.connect(self.db_path) as conn:
            conn.row_factory = sqlite3.Row
            cursor = conn.execute("SELECT * FROM investigations WHERE id = ?", (investigation_id,))
            row = cursor.fetchone()

            if row:
                data = {
                    "id": row["id"],
                    "name": row["name"],
                    "labels": json.loads(row["labels"]),
                    "start_time": row["start_time"],
                    "end_time": row["end_time"],
                    "status": row["status"],
                    "analyses": json.loads(row["analyses"]),
                    "created_at": row["created_at"],
                    "updated_at": row["updated_at"],
                    "metadata": json.loads(row["metadata"]),
                }
                return Investigation.from_dict(data)
            return None

    def list_investigations(self, limit: int = 10, status: Optional[InvestigationStatus] = None) -> List[Investigation]:
        """List investigations"""
        with sqlite3.connect(self.db_path) as conn:
            conn.row_factory = sqlite3.Row

            if status:
                cursor = conn.execute(
                    "SELECT * FROM investigations WHERE status = ? ORDER BY created_at DESC LIMIT ?",
                    (status.value, limit),
                )
            else:
                cursor = conn.execute("SELECT * FROM investigations ORDER BY created_at DESC LIMIT ?", (limit,))

            investigations = []
            for row in cursor.fetchall():
                data = {
                    "id": row["id"],
                    "name": row["name"],
                    "labels": json.loads(row["labels"]),
                    "start_time": row["start_time"],
                    "end_time": row["end_time"],
                    "status": row["status"],
                    "analyses": json.loads(row["analyses"]),
                    "created_at": row["created_at"],
                    "updated_at": row["updated_at"],
                    "metadata": json.loads(row["metadata"]),
                }
                investigations.append(Investigation.from_dict(data))

            return investigations

    def delete_investigation(self, investigation_id: str) -> bool:
        """Delete an investigation"""
        with sqlite3.connect(self.db_path) as conn:
            cursor = conn.execute("DELETE FROM investigations WHERE id = ?", (investigation_id,))
            conn.commit()
            deleted = cursor.rowcount > 0
            if deleted:
                logger.info(f"🗑️ Deleted investigation {investigation_id}")
            return deleted
