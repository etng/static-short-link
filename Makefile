VERSION=v0.1.0
run:
	go run cmd/build/main.go
save:
	git add .
	git commit -m '${shell cat .commit_msg}' || true
	git push origin master -u
	git tag -a ${VERSION} -m ${VERSION} -f
	git push origin ${VERSION} -f