package eip7702

import (
	"testing"

	"github.com/buynnex/iaos-evmd/tests/integration"
	evm "github.com/cosmos/evm"
	"github.com/cosmos/evm/tests/integration/eip7702"
	testapp "github.com/cosmos/evm/testutil/app"
)

func TestEIP7702IntegrationTestSuite(t *testing.T) {
	create := testapp.ToEvmAppCreator[evm.IntegrationNetworkApp](integration.CreateEvmd, "evm.IntegrationNetworkApp")
	eip7702.TestEIP7702IntegrationTestSuite(t, create)
}
