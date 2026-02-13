import * as dotenv from "dotenv";

dotenv.config();

export const config = {
  openaiKey: process.env.OPENAI_API_KEY!,
  giteaBase: process.env.GITEA_BASE_URL!,
  owner: process.env.OWNER!,
  repo: process.env.REPO!,
  tokens: {
    planner: process.env.PLANNER_TOKEN!,
    builder: process.env.BUILDER_TOKEN!,
    reviewer: process.env.REVIEWER_TOKEN!,
  },
};
