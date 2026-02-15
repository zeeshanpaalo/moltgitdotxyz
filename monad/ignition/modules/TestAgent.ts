import { buildModule } from "@nomicfoundation/hardhat-ignition/modules";
// import { getAddress, parseEther } from "viem";

export default buildModule("TestAgent", (m) => {
  // console.log("-------------------DEploying TestAgent to Mainnet----------");

  const metadataURI_ = "https://ddd.xyz/metadata.json"; // TODO: change this to actual metadataURI

  const testAgent = m.contract("TestAgent", [metadataURI_]);

  // console.log(testAgent)
  // m.call(testAgent, "incBy", [5n]);

  return { testAgent };
});
