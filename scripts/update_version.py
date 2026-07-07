import re
import os
import argparse
import subprocess


def get_latest_git_tag() -> str:
    """
    現在のリポジトリから、vX.Y.Z形式の最新タグを取得する。
    タグが1つも無い場合は v0.0.0 を返す。
    """
    try:
        result = subprocess.run(
            ["git", "tag", "--sort=-v:refname"],
            capture_output=True,
            text=True,
            check=True,
        )
        tags = result.stdout.splitlines()
        pattern = re.compile(r"^v\d+\.\d+\.\d+$")
        for tag in tags:
            if pattern.match(tag.strip()):
                return tag.strip()
    except (subprocess.CalledProcessError, FileNotFoundError):
        pass
    return "v0.0.0"


def compute_next_version(latest_tag: str, bump: str) -> str:
    """
    auto-tag.ymlと同じ計算式で次のバージョンを算出する。
    (major.minor.patchのpatchを基本的にインクリメントする)
    """
    version_str = latest_tag.lstrip("v")
    major, minor, patch = (int(x) for x in version_str.split("."))

    if bump == "major":
        major += 1
        minor = 0
        patch = 0
    elif bump == "minor":
        minor += 1
        patch = 0
    else:
        patch += 1

    return f"{major}.{minor}.{patch}"


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
    parser.add_argument(
        "version",
        nargs="?",
        default=None,
        help="new version string (e.g. 1.2.3). 省略した場合はgitタグから自動計算する。"
    )
    parser.add_argument(
        "--bump",
        choices=["patch", "minor", "major"],
        default="patch",
        help="versionを省略した場合の上げ方(デフォルト: patch)"
    )
    parser.add_argument(
        "--file",
        default="internal/ext/types/constants.go",
        help="path to constants.go"
    )

    args = parser.parse_args()

    if args.version:
        target_version = args.version
    else:
        latest_tag = get_latest_git_tag()
        target_version = compute_next_version(latest_tag, args.bump)
        print(f"Latest git tag: {latest_tag} -> Next version: {target_version}")

    update_version(args.file, target_version)
