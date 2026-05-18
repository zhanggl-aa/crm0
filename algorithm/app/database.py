"""
Database connection module with connection pooling for PostgreSQL.
"""

import os
import logging
from contextlib import contextmanager

import psycopg2
from psycopg2 import pool

logger = logging.getLogger(__name__)

# Database configuration from environment variables with defaults matching Go service
DB_HOST = os.getenv("DB_HOST", "localhost")
DB_PORT = int(os.getenv("DB_PORT", "5432"))
DB_USER = os.getenv("DB_USER", "postgres")
DB_PASSWORD = os.getenv("DB_PASSWORD", "123456")
DB_NAME = os.getenv("DB_NAME", "crm0_db")
DB_MIN_CONN = int(os.getenv("DB_MIN_CONN", "2"))
DB_MAX_CONN = int(os.getenv("DB_MAX_CONN", "10"))

_connection_pool: pool.ThreadedConnectionPool | None = None


def init_connection_pool() -> pool.ThreadedConnectionPool:
    """Initialize the connection pool."""
    global _connection_pool
    if _connection_pool is None:
        _connection_pool = pool.ThreadedConnectionPool(
            minconn=DB_MIN_CONN,
            maxconn=DB_MAX_CONN,
            host=DB_HOST,
            port=DB_PORT,
            user=DB_USER,
            password=DB_PASSWORD,
            dbname=DB_NAME,
        )
        logger.info(
            "Database connection pool created: %s:%d/%s (min=%d, max=%d)",
            DB_HOST, DB_PORT, DB_NAME, DB_MIN_CONN, DB_MAX_CONN,
        )
    return _connection_pool


def close_connection_pool() -> None:
    """Close all connections in the pool."""
    global _connection_pool
    if _connection_pool is not None:
        _connection_pool.closeall()
        _connection_pool = None
        logger.info("Database connection pool closed.")


@contextmanager
def get_connection():
    """
    Context manager that provides a database connection from the pool.
    Automatically commits on success and rolls back on exception.
    """
    pool_instance = init_connection_pool()
    conn = pool_instance.getconn()
    try:
        yield conn
        conn.commit()
    except Exception:
        conn.rollback()
        raise
    finally:
        pool_instance.putconn(conn)


@contextmanager
def get_cursor(dict_cursor: bool = False):
    """
    Context manager that provides a database cursor.
    If dict_cursor is True, returns a RealDictCursor for dict-like rows.
    """
    with get_connection() as conn:
        if dict_cursor:
            from psycopg2.extras import RealDictCursor
            cur = conn.cursor(cursor_factory=RealDictCursor)
        else:
            cur = conn.cursor()
        try:
            yield cur
        finally:
            cur.close()
