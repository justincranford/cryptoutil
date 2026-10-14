package dockerfile_healthcheck

const (
	testHealthcheckCmdAppSmKMSLivezExec                = `HEALTHCHECK CMD ["/app/sm-kms", "livez"]`
	testHealthcheckCmdShouldBe                         = "HEALTHCHECK CMD should be"
	testDeploymentsSmKMSDockerfile                     = "deployments/sm-kms/Dockerfile"
	testNoHealthcheckInstructionIsAcceptableForNonPsID = "no HEALTHCHECK instruction is acceptable for non-PS-ID"
)
