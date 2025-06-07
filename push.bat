git add .
git commit -m %1
git fetch sub-repo
git subtree pull --prefix=eec-deleter-go sub-repo main --squash
git push origin main

