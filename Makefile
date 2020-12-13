

generate:
	controller-gen paths=./pkg/apis/... crd:trivialVersions=true rbac:roleName=controller-perms output:crd:artifacts:config=config/crd