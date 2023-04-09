# Tagging instructions

Create Relase
```
TAG=0.0.5
git tag -a v${TAG} -m "Creating v${TAG} release"
git push origin v${TAG}
```

Delete Release
```
TAG=0.0.5
git tag -d v${TAG}
git push origin --delete v${TAG}
```