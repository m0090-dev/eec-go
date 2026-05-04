import re
import os
import argparse

def update_version(file_path: str, new_version: str):
    if not os.path.exists(file_path):
        print(f"Error: {file_path} not found.")
        return

    with open(file_path, "r", encoding="utf-8") as f:
        content = f.read()

    # VERSION = "x.y.z" だけを狙い撃ち
    new_content = re.sub(
    r'(VERSION\s*=\s*")[^"]+(")',
    lambda m: f'{m.group(1)}{new_version}{m.group(2)}',
    content
    )

    if new_content == content:
        print("No changes (version already same?)")
        return

    with open(file_path, "w", encoding="utf-8") as f:
        f.write(new_content)

    print(f"Updated VERSION -> {new_version}")

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Update Go VERSION constant")
    parser.add_argument("version", help="new version string (e.g. 1.2.3)")
    parser.add_argument(
        "--file",
        default="internal/ext/types/constants.go",
        help="path to constants.go"
    )

    args = parser.parse_args()
    update_version(args.file, args.version)
