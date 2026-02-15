import axios from "axios";
import * as fs from "fs";
import * as path from "path";
import { config } from "./config";

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

  const res = await axios.post(
    `${config.API_BASE_URL}/user/sign_up/new?jsondata=true`,
    {
      email,
      username: agentName,
    },
  );

  console.log(res.data);
  const apiKey = res.data.api_key;

  fs.writeFileSync(tokenPath, apiKey);

  const giteaToken = res.data.gitea_token;
  fs.writeFileSync(
    path.join(TOKENS_DIR, `${agentName}.gitea.token`),
    giteaToken,
  );
  console.log(
    `✅ ${agentName} registered and token  and personal access token saved`,
  );
  return apiKey;
}

export async function initAgents() {
  const plannerToken = await registerAgent(
    config.plannerAgentName,
    config.plannerAgentEmail,
  );
  const builderToken = await registerAgent(
    config.builderAgentName,
    config.builderAgentEmail,
  );
  const reviewerToken = await registerAgent(
    config.reviewerAgentName,
    config.reviewerAgentEmail,
  );

  return {
    planner: plannerToken,
    builder: builderToken,
    reviewer: reviewerToken,
  };
}
