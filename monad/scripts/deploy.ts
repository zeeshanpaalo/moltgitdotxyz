// import { buildModule } from "@nomicfoundation/hardhat-ignition/modules";

// // If you want, you can import viem helpers, but not needed for basic mint
// // import { getAddress } from "viem";

// export default buildModule("TestAgentModule", (m) => {
//   // ------------------- Deploying TestAgent to Mainnet -------------------

//   const metadataURI = "https://moltgit.xyz/metadata.json"; // TODO: replace with actual IPFS URI

//   // Deploy the TestAgent contract
//   const testAgent = m.contract("TestAgent", [
//     metadataURI, // constructor argument
//   ]);

//   return { testAgent };
// });




import hre from "hardhat";
import TestAgentModule from "../ignition/modules/TestAgent.js";

async function main() {
  // Connect to network
  const connection = await hre.network.connect();

  // Deploy the Ignition module
  const { testAgent } = await connection.ignition.deploy(TestAgentModule);

  // Log the deployed address
  console.log("TestAgent deployed at:", testAgent.address);
}

main().catch((error) => {
  console.error(error);
  process.exit(1);
});

