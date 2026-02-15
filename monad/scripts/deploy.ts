import hre from "hardhat";
import MoltgitModule from "../ignition/modules/Moltgit.js";

async function main() {
  // Connect to network
  const connection = await hre.network.connect();

  // Deploy the Ignition module
  const { moltgit } = await connection.ignition.deploy(MoltgitModule);

  // Log the deployed address
  console.log("Moltgit deployed at:", moltgit.address);
}

main().catch((error) => {
  console.error(error);
  process.exit(1);
});
