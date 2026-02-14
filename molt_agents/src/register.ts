import axios from "axios";
import * as fs from "fs";
import * as path from "path";

const BACKEND_URL = "http://localhost:3000"; // your moltgit backend

const TOKENS_DIR = path.join(process.cwd(), "tokens");

if (!fs.existsSync(TOKENS_DIR)) {
  fs.mkdirSync(TOKENS_DIR);
}

async function registerAgent(
  agentName: string,
  email: string,
): Promise<string> {
  console.log("registerAgent called with", agentName, email);
  const tokenPath = path.join(TOKENS_DIR, `${agentName}.token`);

  // If token already stored, reuse it
  if (fs.existsSync(tokenPath)) {
    const existingToken = fs.readFileSync(tokenPath, "utf8");
    console.log(`🔑 Loaded existing token for ${agentName}`);
    return existingToken;
  }

  console.log(`📝 Registering ${agentName}...`);

  const res = await axios.post(`${BACKEND_URL}/user/sign_up/new?jsondata=true`, {
    email,
    username: agentName,
  });

  console.log(res.data);
  const apiKey = res.data.api_key;

  fs.writeFileSync(tokenPath, apiKey);

  console.log(`✅ ${agentName} registered and token saved`);

  return apiKey;
}

export async function initAgents() {
  const plannerToken = await registerAgent("planner-1", "planner@example.com");
  const builderToken = await registerAgent("builder-1", "builder@example.com");
  const reviewerToken = await registerAgent("reviewer-1", "reviewer@example.com");

  return {
    planner: plannerToken,
    builder: builderToken,
    reviewer: reviewerToken,
  };
}
