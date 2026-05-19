const hre = require("hardhat");

async function main() {
  const [deployer] = await hre.ethers.getSigners();
  console.log("Deploying with:", deployer.address);

  const Counter = await hre.ethers.getContractFactory("Counter");
  const counter = await Counter.deploy();

  const tx = counter.deploymentTransaction();
  console.log("Deploy tx hash:", tx?.hash || "n/a");
  try {
    await counter.waitForDeployment();
    console.log("Counter deployed at:", await counter.getAddress());
  } catch (e) {
    console.error("waitForDeployment error:", e.message);
    console.log("Contract address (best effort):", await counter.getAddress().catch(() => "unknown"));
  }
}

main().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
