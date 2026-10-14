package dockerfile_healthcheck

const (
	testHealthcheckCmdAppSmKMSLivezExit1               = "HEALTHCHECK CMD /app/sm-kms livez || exit 1"
	testHealthcheckCmdShouldBe                         = "HEALTHCHECK CMD should be"
	testDeploymentsSmKMSDockerfile                     = "deployments/sm-kms/Dockerfile"
	testNoHealthcheckInstructionIsAcceptableForNonPsID = "no HEALTHCHECK instruction is acceptable for non-PS-ID"
)
