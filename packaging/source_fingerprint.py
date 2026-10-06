"""Hash build inputs, not generated artifacts, private keys, or test evidence."""
from pathlib import Path
import hashlib

def source_index(root: Path) -> dict:
    paths = []
    for directory in ('manager', 'packaging', 'tests', '.github/workflows'):
        for p in (root/directory).rglob('*'):
            if p.is_file() and 'evidence' not in p.parts and '__pycache__' not in p.parts:
                if p.suffix in ('.go','.js','.css','.html','.json','.py','.sh','.yml','.yaml','.ini','.conf','.cmd') or p.name=='go.mod':
                    paths.append(p)
    paths.extend(root/n for n in ('VERSION','project.json') if (root/n).is_file())
    return {p.relative_to(root).as_posix():hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted(paths)}
