#!/usr/bin/env python3
from pathlib import Path
import os


def read_dotenv() -> dict[str, str]:
    path = Path(".env")
    if not path.exists():
        return {}

    values: dict[str, str] = {}
    for raw_line in path.read_text().splitlines():
        line = raw_line.strip()
        if not line or line.startswith("#") or "=" not in line:
            continue
        key, value = line.split("=", 1)
        values[key.strip()] = value.strip().strip('"').strip("'")
    return values


def env_value(values: dict[str, str], key: str, fallback: str) -> str:
    return os.environ.get(key) or values.get(key) or fallback


def sql_identifier(value: str) -> str:
    return '"' + value.replace('"', '""') + '"'


def sql_literal(value: str) -> str:
    return "'" + value.replace("'", "''") + "'"


dotenv = read_dotenv()
user = env_value(dotenv, "POSTGRES_USER", "dsa_slate")
password = env_value(dotenv, "POSTGRES_PASSWORD", "change-me")

print(f"ALTER ROLE {sql_identifier(user)} WITH PASSWORD {sql_literal(password)};")
