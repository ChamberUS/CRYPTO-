require("@nomicfoundation/hardhat-toolbox");

const RPC = process.env.EVM_RPC || "http://127.0.0.1:8545";
const CHAIN_ID = parseInt(process.env.EVM_CHAIN_ID || "9000", 10);
const PRIVATE_KEY = process.env.PRIVATE_KEY || "";

/** @type import('hardhat/config').HardhatUserConfig */
module.exports = {
  solidity: "0.8.24",
  networks: {
    evmd: {
      url: RPC,
      chainId: CHAIN_ID,
      accounts: PRIVATE_KEY ? [PRIVATE_KEY] : undefined,
    },
  },
};
