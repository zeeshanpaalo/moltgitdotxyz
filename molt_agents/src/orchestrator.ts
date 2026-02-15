import { plannerLoop } from "./planner";
import { builderLoop } from "./builder";
import { reviewerLoop } from "./reviewer";
import { initAgents } from "./register";

let isRunning = false;

async function run() {
  if (isRunning) {
    console.log("⏳ Previous cycle still running, skipping...");
    return;
  }

  isRunning = true;

  try {
    console.log("🚀 Agent cycle started at", new Date().toISOString());

    await plannerLoop();
    await builderLoop();
    await reviewerLoop();

    console.log("✅ Agent cycle completed");
  } catch (err) {
    console.error("❌ Loop error:", err);
  } finally {
    isRunning = false;
  }
}

async function start() {
  try {
    console.log("🔧 Initializing agents (registering if needed)...");

    // Ensures tokens exist in /tokens folder
    await initAgents();

    console.log("🤖 Agents ready to perform action");

    // Run immediately when script starts
    await run();

    // Schedule every 15 minutes
    setInterval(run, 10 * 60 * 1000);

    console.log("⏱ Agents scheduled every 15 minutes...");
  } catch (err) {
    console.error("❌ Failed to start agents:", err);
  }
}

start();
