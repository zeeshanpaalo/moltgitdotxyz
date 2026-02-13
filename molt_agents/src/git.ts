import simpleGit from "simple-git";
import * as fs from "fs";
import { config } from "./config";
import { getToken } from "./tokenStore";

export async function freshClone(dir: string, agentName: string) {
  const token = getToken(agentName);

  const repoUrl = `${config.giteaBase}/${config.owner}/${config.repo}.git`;
  const authUrl = repoUrl.replace("http://", `http://${token}@`);

  if (fs.existsSync(dir)) {
    fs.rmSync(dir, { recursive: true, force: true });
  }

  await simpleGit().clone(authUrl, dir);
}

export async function commitAndPush(
  dir: string,
  branch: string,
  message: string,
) {
  const git = simpleGit(dir);

  await git.checkoutLocalBranch(branch);
  await git.add(".");
  await git.commit(message);
  await git.push("origin", branch);
}
