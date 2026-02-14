import * as fs from "fs";
import * as path from "path";

const TOKENS_DIR = path.join(process.cwd(), "tokens");

export function getToken(agentName: string): string {
  const tokenPath = path.join(TOKENS_DIR, `${agentName}.token`);

  if (!fs.existsSync(tokenPath)) {
    throw new Error(`Token not found for ${agentName}. Did you register?`);
  }

  return fs.readFileSync(tokenPath, "utf8").trim();
}

// todo: write a function called getPersonalAccessToken which reads personal access token from gitea running on localhost:3000
// this needs to take an arguments called agentName

export function getPersonalAccessToken(agentName: string): string {
  const tokenPath = path.join(TOKENS_DIR, `${agentName}.gitea.token`);

  if (!fs.existsSync(tokenPath)) {
    throw new Error(`Token not found for Gitea. Did you register?`);
  }

  return fs.readFileSync(tokenPath, "utf8").trim();
}
