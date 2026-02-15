import hre from "hardhat";
import { encodeFunctionData } from "viem";
import MolgitAbi from "../artifacts/contracts/Moltgit.sol/Moltgit.json";
import "dotenv/config";

const CONTRACT_ADDRESS = process.env.CONTRACT_ADDRESS as `0x${string}`;

export function wait(seconds: number): Promise<void> {
  return new Promise((resolve) => {
    setTimeout(resolve, seconds * 1000);
  });
}

async function main() {
  if (!CONTRACT_ADDRESS) {
    throw new Error("❌ CONTRACT_ADDRESS not set");
  }

  // Connect to network (HH3 way)
  const { provider } = await hre.network.connect();

  // Get accounts from configured private key
  const accounts = await provider.request({
    method: "eth_accounts",
  });

  if (!accounts || accounts.length === 0) {
    throw new Error("❌ No accounts available");
  }

  const from = accounts[0];
  console.log("👤 Using account:", from);

  // Encode mintToAgent() calldata
  const data = encodeFunctionData({
    abi: MolgitAbi.abi,
    functionName: "mint",
    args: ["0xDc02F29D475AE8945A661fcc5008844c4c8941c5"],
  });

  // Send transaction
  const txHash = await provider.request({
    method: "eth_sendTransaction",
    params: [
      {
        from,
        to: CONTRACT_ADDRESS,
        data,
      },
    ],
  });

  console.log("⏳ mintedToAgent tx sent:", txHash);

  // Wait for confirmation
  const receipt = await provider.request({
    method: "eth_getTransactionReceipt",
    params: [txHash],
  });

  console.log("✅ mintedToAgent confirmed in block", receipt.blockNumber);
}

main().catch((err) => {
  console.error("❌ Error:", err);
  process.exitCode = 1;
});
