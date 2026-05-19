package distribution

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/buynnex/iaos-evmd/tests/integration"
	evm "github.com/cosmos/evm"
	"github.com/cosmos/evm/tests/integration/precompiles/distribution"
	testapp "github.com/cosmos/evm/testutil/app"
)

func TestDistributionPrecompileTestSuite(t *testing.T) {
	create := testapp.ToEvmAppCreator[evm.DistributionPrecompileApp](integration.CreateEvmd, "evm.DistributionPrecompileApp")
	s := distribution.NewPrecompileTestSuite(create)
	suite.Run(t, s)
}

func TestDistributionPrecompileIntegrationTestSuite(t *testing.T) {
	create := testapp.ToEvmAppCreator[evm.DistributionPrecompileApp](integration.CreateEvmd, "evm.DistributionPrecompileApp")
	distribution.TestPrecompileIntegrationTestSuite(t, create)
}
