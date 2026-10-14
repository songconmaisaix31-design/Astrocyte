import ast
import io
from pathlib import Path
import subprocess
import sys
import tokenize

ROOT = Path(__file__).resolve().parents[3]
FOLDERS = ("events", "scripts/validation", "scripts/evidence", "scripts/experiment", "internal/adapters/importers", "tests/validation")


def without_docs(tree):
    for node in ast.walk(tree):
        if isinstance(node, (ast.Module, ast.ClassDef, ast.FunctionDef, ast.AsyncFunctionDef)):
            if node.body and isinstance(node.body[0], ast.Expr) and isinstance(node.body[0].value, ast.Constant) and isinstance(node.body[0].value.value, str):
                node.body.pop(0)
    return ast.dump(tree, include_attributes=False)


def check(baseline):
    files = sorted(path for folder in FOLDERS for path in (ROOT / folder).rglob("*.py"))
    frozen = 0
    for path in files:
        relative = path.relative_to(ROOT).as_posix()
        source = path.read_text(encoding="utf-8")
        tree = ast.parse(source)
        if ast.get_docstring(tree) or source.startswith("#!"):
            raise ValueError("Python header rejected: " + relative)
        for token in tokenize.generate_tokens(io.StringIO(source).readline):
            if token.type == tokenize.COMMENT and not any("\u4e00" <= char <= "\u9fff" for char in token.string):
                raise ValueError("Chinese comment required: " + relative + ":" + str(token.start[0]))
        for node in ast.walk(tree):
            if isinstance(node, ast.Try) and any(isinstance(child, (ast.Import, ast.ImportFrom)) for statement in node.body for child in ast.walk(statement)):
                raise ValueError("Conditional import rejected: " + relative + ":" + str(node.lineno))
        # 只比较内存中的 AST，不修改文件；全部 docstring 与行号不参与比较。
        if relative.startswith("scripts/experiment/"):
            original = subprocess.run(["git", "show", baseline + ":" + relative], cwd=ROOT, check=True, capture_output=True).stdout.decode("utf-8")
            if without_docs(ast.parse(original)) != without_docs(tree):
                raise ValueError("Frozen experiment AST changed: " + relative)
            frozen += 1
    print("Python parse/header/comments/imports PASS: " + str(len(files)) + " files; frozen AST PASS: " + str(frozen) + " files")


if __name__ == "__main__":
    check(sys.argv[1])
